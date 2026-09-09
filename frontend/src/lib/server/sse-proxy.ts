export async function* streamToAsyncIterable(
	stream: ReadableStream<Uint8Array>,
	signal?: AbortSignal
): AsyncGenerator<Uint8Array> {
	const reader = stream.getReader();
	// Cancel the reader immediately when aborted so the pending reader.read()
	// resolves with {done:true} right away instead of waiting for the next chunk.
	const cancelReader = () => reader.cancel().catch(() => {});
	signal?.addEventListener('abort', cancelReader, { once: true });
	try {
		while (true) {
			const { done, value } = await reader.read();
			if (done) return;
			yield value;
		}
	} finally {
		signal?.removeEventListener('abort', cancelReader);
		reader.cancel().catch(() => {});
	}
}

/**
 * True for the rejection an aborted fetch or stream reader produces.
 *
 * Deliberately checks the name rather than `instanceof Error`: an abort rejects
 * with a `DOMException`, and whether that extends `Error` depends on the
 * runtime — it does under Node, it does not under jsdom. Getting this wrong is
 * not cosmetic: everything below treats an unrecognised rejection as a real
 * upstream fault, so a mis-classified abort would announce a broken connection
 * every time a viewer closed their tab.
 */
function isAbortError(e: unknown): boolean {
	return typeof e === 'object' && e !== null && (e as { name?: unknown }).name === 'AbortError';
}

type UpstreamOk = { ok: true; response: Response; setCookies: string[] };
type UpstreamReject = { ok: false; clientResponse: Response };
export type UpstreamResult = UpstreamOk | UpstreamReject;

/**
 * Reads any Set-Cookie headers off an upstream response.
 *
 * The backend's authorization middleware renews an expired access token from
 * the refresh cookie and hands the new access_token back via Set-Cookie on the
 * connection's response headers. Those headers land here, on the proxy's fetch
 * — not on the browser — so unless we forward them the browser keeps its stale
 * (now deleted) cookie and the next SSE reconnect fails, freezing the stream.
 * getSetCookie() is the standard multi-value accessor (undici / Node 18.14+).
 */
function readSetCookies(response: Response): string[] {
	const headers = response.headers as Headers & { getSetCookie?: () => string[] };
	return typeof headers.getSetCookie === 'function' ? headers.getSetCookie() : [];
}

/**
 * Resolves an upstream fetch promise and validates the response.
 * Returns { ok: true, response, setCookies } on success (status 200), where
 * setCookies are any renewed auth cookies to forward back to the browser.
 * Returns { ok: false, clientResponse } on abort or non-200 status — the caller
 * should return clientResponse directly to the browser.
 */
export async function resolveUpstream(
	responses: Record<string, Promise<Response>>,
	abortCtl: AbortController
): Promise<UpstreamResult> {
	let resp: Response;
	try {
		resp = await Promise.race(Object.values(responses));
	} catch (e) {
		if (isAbortError(e)) {
			return { ok: false, clientResponse: new Response(null, { status: 200 }) };
		}
		// The route is about to answer 500 and nobody will ever read the sibling
		// fetches — but they are detached promises still holding open sockets and
		// reading bodies. Take them down before leaving.
		abortCtl.abort();
		throw e;
	}
	if (resp.status !== 200) {
		return { ok: false, clientResponse: await upstreamErrorResponse(resp, abortCtl) };
	}
	return { ok: true, response: resp, setCookies: readSetCookies(resp) };
}

/**
 * Aborts the controller, then returns the upstream error as a JSON response.
 * Use when the upstream status is not 200.
 */
export async function upstreamErrorResponse(
	resp: Response,
	abortCtl: AbortController
): Promise<Response> {
	const data = await resp.json().catch(() => ({}));
	abortCtl.abort();
	return new Response(JSON.stringify(data), {
		status: resp.status,
		headers: { 'Content-Type': 'application/json' }
	});
}

/** Maps a named URL record into parallel fetch promises sharing the same RequestInit. */
export function startFetches(
	requests: Record<string, string>,
	fetch: (url: string, init: RequestInit) => Promise<Response>,
	init: RequestInit
): Record<string, Promise<Response>> {
	return Object.fromEntries(
		Object.entries(requests).map(([name, url]) => [name, fetch(url, init)])
	);
}

/**
 * Creates an AbortController wired to client disconnect signals.
 * Use this instead of inline listeners to handle the dev/prod platform difference.
 */
export function createAbortController(
	request: Request,
	platform: App.Platform | undefined
): AbortController {
	const abortCtl = new AbortController();
	request.signal.addEventListener('abort', () => abortCtl.abort());
	// Only available under adapter-node (node build/index.js); undefined in vite dev/preview.
	// Without this, SSE connections will not be torn down on browser disconnect in production.
	platform?.req.once('close', () => abortCtl.abort());
	return abortCtl;
}

export type ChunkHandler = (
	eventName: string,
	ctrl: ReadableStreamDefaultController,
	chunk: Uint8Array
) => void;

/**
 * Shared encoder for the frames this module emits on its own behalf (heartbeat,
 * upstream errors) and for the reframer's output.
 *
 * Everything enqueued must be bytes, not a string. SvelteKit's node adapter
 * happens to accept both (it ends up in `res.write()`), but a plain
 * `new Response(stream)` — which is what a unit test reaches for — rejects a
 * string chunk outright with "Received non-Uint8Array chunk". Keeping one type
 * on the stream costs nothing and keeps the proxy testable with standard APIs.
 */
const encoder = new TextEncoder();

/**
 * Event name used to tell the browser that one of the upstreams behind a
 * multiplexed route answered with an error. Deliberately not `error`:
 * EventSource routes an event literally named `error` to `onerror`, where it
 * would be indistinguishable from an ordinary connection drop.
 */
export const UPSTREAM_ERROR_EVENT = 'upstream-error';

/**
 * The status reported when the upstream fetch never produced a Response at all —
 * the backend process is down, the connection was refused, DNS failed, the
 * socket reset. There is no real status to forward, so we say what a proxy that
 * could not reach its upstream says. It is deliberately a retryable one: the
 * backend will very likely be back, and the client's existing classification
 * turns it into backoff rather than a terminal notice.
 */
export const UPSTREAM_UNREACHABLE_STATUS = 502;

/**
 * Builds the frame announcing a failed upstream.
 *
 * It must carry a `data:` line. EventSource only dispatches an event once its
 * data buffer is non-empty, so a frame that is nothing but `event:` (or `id:`)
 * is parsed and then silently discarded — the client would never hear about it.
 */
function upstreamErrorFrame(streamName: string, status: number): Uint8Array {
	return encoder.encode(
		`event: ${UPSTREAM_ERROR_EVENT}\ndata: ${JSON.stringify({ stream: streamName, status })}\n\n`
	);
}

/**
 * Creates a stateful {@link ChunkHandler} that labels an upstream SSE stream one
 * *message* at a time instead of one transport chunk at a time.
 *
 * The naive approach — prepend `event: <name>\n` to every raw chunk — corrupts
 * any message whose bytes span more than one chunk (e.g. a large scoreboard
 * payload), because a `ReadableStream` chunk boundary is arbitrary and the
 * injected prefix can land in the middle of a `data:` line. Small payloads
 * (the time tick) almost always arrive whole, which is why they never break.
 *
 * Here we buffer decoded text and only emit a labelled event once a full
 * message boundary (a blank line, `\n\n`) has arrived; any trailing partial
 * stays buffered for the next chunk. A streaming TextDecoder also keeps
 * multi-byte UTF-8 intact when a character is split across chunks. The upstream
 * backend terminates every message with `\n\n` (see the Go SSE presenters), so
 * `\n\n` is the definitive frame delimiter.
 *
 * Each stream must get its own reframer instance (its own buffer); call this
 * once per event name.
 */
export function createEventReframer(): ChunkHandler {
	const decoder = new TextDecoder();
	let buffer = '';
	return (eventName, ctrl, chunk) => {
		buffer += decoder.decode(chunk, { stream: true });
		let idx: number;
		while ((idx = buffer.indexOf('\n\n')) !== -1) {
			const message = buffer.slice(0, idx);
			buffer = buffer.slice(idx + 2);
			// Skip empty separators / stray blank lines; only forward real messages.
			if (message.length === 0) continue;
			ctrl.enqueue(encoder.encode(`event: ${eventName}\n${message}\n\n`));
		}
	};
}

/**
 * Creates a ReadableStream that fans out each upstream response to its handler,
 * handling heartbeat, error logging, and cleanup automatically.
 */
export function createSseStream(
	abortCtl: AbortController,
	logName: string,
	responses: Record<string, Promise<Response>>,
	handlers: Record<string, ChunkHandler>
): ReadableStream {
	return new ReadableStream({
		async start(controller) {
			const stopHeartbeat = startHeartbeat(
				() => controller.enqueue(encoder.encode(': keepalive\n\n')),
				abortCtl.signal
			);
			// Several upstreams can fail at once, but the client only needs telling
			// once — it takes the whole connection down on the first report, so a
			// second frame is at best ignored and at worst counted as a second
			// failure. Every failure is still logged; only the announcement is
			// one-shot, and a rejection and a non-200 share the same allowance.
			let reported = false;
			const report = (eventName: string, status: number) => {
				if (!reported) {
					reported = true;
					controller.enqueue(upstreamErrorFrame(eventName, status));
				}
				abortCtl.abort();
			};

			/** Validates one upstream and pumps its body into the client stream. */
			const relay = async (eventName: string, handle: ChunkHandler) => {
				const response = await responses[eventName];
				// Every upstream is validated here, not just the one resolveUpstream()
				// happened to race first. Reading a non-200's body as stream content
				// corrupts nothing — the reframer buffers a JSON error body forever,
				// waiting for a `\n\n` that never arrives — but that is the bad part:
				// the sub-stream dies in silence while its siblings keep flowing, so
				// the countdown ticks on beside a frozen scoreboard and the browser is
				// none the wiser. Say so, then drop the whole connection and let the
				// client decide when to come back.
				if (response.status !== 200) {
					console.warn(`[stream ${logName}] upstream "${eventName}" answered ${response.status}`);
					await response.body?.cancel().catch(() => {});
					report(eventName, response.status);
					return;
				}
				const body = response.body;
				if (!body) return;
				for await (const chunk of streamToAsyncIterable(body, abortCtl.signal)) {
					handle(eventName, controller, chunk);
				}
			};

			try {
				await Promise.all(
					Object.entries(handlers).map(async ([eventName, handle]) => {
						try {
							await relay(eventName, handle);
						} catch (e) {
							// Our own teardown: the client went away, or a sibling
							// upstream already failed and aborted the group. Reporting
							// this would invent a fault that never happened.
							if (isAbortError(e)) return;
							// The fetch never produced a Response — the backend process
							// is down, the connection was refused, DNS failed, the socket
							// reset. Without this the browser received a clean
							// 200 + heartbeat + EOF, indistinguishable from healthy, and
							// reconnected once a second forever with nothing on screen to
							// say why. (This also catches a body that starts fine and
							// then errors mid-read.)
							console.warn(`[stream ${logName}] upstream "${eventName}" failed`, e);
							report(eventName, UPSTREAM_UNREACHABLE_STATUS);
						}
					})
				);
			} catch (e) {
				// Every task catches its own now, so this is a backstop for anything
				// thrown outside them.
				if (isAbortError(e)) {
					console.debug(`[stream ${logName}] upstream aborted (client disconnected)`);
				} else {
					console.warn(`[stream ${logName}] unexpected stream error`, e);
				}
			} finally {
				stopHeartbeat();
				try {
					controller.close();
				} catch {}
			}
		},
		cancel() {
			abortCtl.abort();
		}
	});
}

/** Wraps a ReadableStream in a standard SSE Response. */
export function sseResponse(stream: ReadableStream, setCookies: string[] = []): Response {
	const headers = new Headers({
		'Content-Type': 'text/event-stream',
		'Cache-Control': 'no-cache',
		Connection: 'keep-alive'
	});
	// Forward any renewed auth cookies from the upstream (see readSetCookies).
	// append (not set) so multiple Set-Cookie headers survive as distinct entries.
	for (const cookie of setCookies) {
		headers.append('set-cookie', cookie);
	}
	return new Response(stream, { headers });
}

/**
 * Sends an immediate keepalive on connect, then periodically, to prevent
 * proxies from closing an idle SSE connection.
 *
 * Returns a cleanup function to stop the interval.
 */
export function startHeartbeat(
	enqueue: () => void,
	signal: AbortSignal,
	intervalMs = 15_000
): () => void {
	enqueue();
	const timer = setInterval(() => {
		if (signal.aborted) {
			clearInterval(timer);
			return;
		}
		try {
			enqueue();
		} catch {
			clearInterval(timer);
		}
	}, intervalMs);

	return () => clearInterval(timer);
}
