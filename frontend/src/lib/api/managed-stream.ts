import { EventSourcePolyfill } from 'event-source-polyfill';

/**
 * Why a connection stopped delivering.
 *
 * `stream` names the multiplexed upstream the proxy blamed, or `connection` when
 * what failed was the initial HTTP response itself. `terminal` says retrying
 * cannot help, so the page can say what is actually true instead of promising a
 * reconnection that will never come.
 */
export type StreamFailure = { stream: string; status: number; terminal: boolean };

export type ManagedStreamOptions = {
	/** Opens a fresh connection. Called again on every supervised retry. */
	connect: () => EventSourcePolyfill;
	/** Listeners to attach, by event name. Receiving any of them clears the failure streak. */
	on: Record<string, (event: MessageEvent) => void>;
	/**
	 * Called once a notice is warranted — a terminal failure, or a streak of
	 * `degradedAfter` retryable ones — and not again until recovery.
	 */
	onDegraded?: (failure: StreamFailure) => void;
	/** Called when the stream is delivering again after `onDegraded`. */
	onRecovered?: () => void;
	/**
	 * Every error event, before this module classifies it. `stop` ends supervision
	 * entirely — the root layout uses it to send an expired session to /login.
	 */
	onError?: (event: Event, stop: () => void) => void;
	/**
	 * Pulls this stream's data back over REST.
	 *
	 * A reconnected stream does not replay what it already sent. The backend's
	 * scoreboard and judge-worker streams push only on change — measured: a
	 * scoreboard stream sends nothing at all for at least ten seconds after
	 * connecting, while the time stream ticks immediately — so after any
	 * reconnect the page goes on showing whatever it had until somebody
	 * submits. On a venue big screen that is the frozen scoreboard we set out
	 * to fix, arriving by a different road.
	 *
	 * REST does not depend on the stream, so this is also the only thing that
	 * helps when an upstream accepted the connection and then went quiet —
	 * which no stream-level signal can tell apart from a healthy silent one.
	 */
	onResync?: () => void;
	/** Consecutive retryable failures before `onDegraded` fires. */
	degradedAfter?: number;
	/** How long a connection must stay quiet before it counts as healthy. */
	settleAfterMs?: number;
	/** While degraded, how often to fall back to `onResync`. */
	resyncEveryMs?: number;
};

const FIRST_RETRY_MS = 1_000;
const MAX_RETRY_MS = 30_000;
const SETTLE_MS = 5_000;
const RESYNC_EVERY_MS = 30_000;

/** 1s, 2s, 4s, 8s, 16s, then 30s for every attempt after that. */
export function retryDelay(attempt: number): number {
	return Math.min(FIRST_RETRY_MS * 2 ** Math.max(0, attempt - 1), MAX_RETRY_MS);
}

/**
 * Statuses where retrying cannot help, because nothing is broken.
 *
 * A contest stream answers 404 once its activation window closes, and a public
 * scoreboard answers 404 once the organiser un-publishes it — both are a contest
 * reaching the end of its normal life, not a fault. 401/403 mean the session is
 * refused, which reconnecting will not change either. Retrying any of these
 * would hang a "connection lost" banner on a venue big screen for a contest that
 * simply finished, and knock on the backend every 30 seconds until someone
 * turned the screen off.
 */
export function isTerminalStatus(status: number): boolean {
	return status === 401 || status === 403 || status === 404;
}

/**
 * Opens an SSE connection and supervises the failures the polyfill handles
 * badly or not at all.
 *
 * A mid-stream drop is left alone: the polyfill reconnects on its own and that
 * path works, so stepping in would only open a second connection beside the one
 * it is already reopening. The two it cannot handle are:
 *
 * - **An `upstream-error` frame.** The proxy has already sent
 *   `200 text/event-stream` by the time it learns one of the streams it
 *   multiplexes is bad, so all it can do is announce the failure and close. The
 *   browser sees a clean EOF with no status, and the polyfill resets its backoff
 *   to one second on every successful open — one request per second, per tab,
 *   for as long as the upstream stays broken, with nothing on screen to say so.
 * - **A connect-time HTTP failure.** Here the polyfill does the opposite: a
 *   response that is not `200 text/event-stream` makes it close *terminally*,
 *   before dispatching the error. One request, no retry, no UI. A backend
 *   restarted mid-contest would freeze the countdown permanently and silently.
 *
 * Returns the teardown function; call it from `$effect`'s cleanup.
 */
export function createManagedStream(options: ManagedStreamOptions): () => void {
	const {
		connect,
		on,
		onDegraded,
		onRecovered,
		onError,
		onResync,
		degradedAfter = 3,
		settleAfterMs = SETTLE_MS,
		resyncEveryMs = RESYNC_EVERY_MS
	} = options;

	let source: EventSourcePolyfill | null = null;
	let retryTimer: ReturnType<typeof setTimeout> | null = null;
	let settleTimer: ReturnType<typeof setTimeout> | null = null;
	let resyncTimer: ReturnType<typeof setInterval> | null = null;
	let failures = 0;
	let degraded = false;
	let stopped = false;
	/** Connections opened so far. The first one needs no resync — the page just loaded its own data. */
	let opens = 0;

	const clearRetry = () => {
		if (retryTimer !== null) {
			clearTimeout(retryTimer);
			retryTimer = null;
		}
	};
	const clearSettle = () => {
		if (settleTimer !== null) {
			clearTimeout(settleTimer);
			settleTimer = null;
		}
	};
	const clearResync = () => {
		if (resyncTimer !== null) {
			clearInterval(resyncTimer);
			resyncTimer = null;
		}
	};

	const stop = () => {
		stopped = true;
		clearRetry();
		clearSettle();
		clearResync();
		source?.close();
		source = null;
	};

	/** The stream is delivering again — drop the streak and take the notice down. */
	const noteHealthy = () => {
		failures = 0;
		clearResync();
		if (degraded) {
			degraded = false;
			onRecovered?.();
		}
	};

	const open = () => {
		if (stopped) return;
		const es = connect();
		source = es;
		// Everything below belongs to THIS connection. Capturing `es` rather than
		// reading the shared `source` keeps a late event from a replaced
		// connection from closing its successor, and `settled` makes a connection
		// fail at most once however many error frames arrive on it.
		let settled = false;
		const mine = () => !stopped && source === es && !settled;

		const fail = (failure: StreamFailure) => {
			if (!mine()) return;
			settled = true;
			clearSettle();
			// Close before the polyfill can schedule its own one-second retry,
			// which would otherwise race ours and defeat the backoff entirely.
			es.close();
			source = null;
			failures += 1;
			if (failure.terminal) stopped = true;

			console.warn(
				`[stream] "${failure.stream}" answered ${failure.status}; ` +
					(failure.terminal ? 'not retrying' : `retry ${failures} in ${retryDelay(failures)}ms`)
			);
			// A terminal failure earns its notice immediately: waiting for a streak
			// that can never grow would leave the page silent for good.
			if (!degraded && (failure.terminal || failures >= degradedAfter)) {
				degraded = true;
				onDegraded?.(failure);
				// The stream stays the primary path; this is only a fallback for
				// when it has been down a while, so a venue board is a little
				// stale rather than frozen. Never for a terminal failure — there
				// is nothing to come back to, and polling it would be the same
				// pointless knocking the terminal classification exists to avoid.
				if (!failure.terminal && onResync) {
					clearResync();
					resyncTimer = setInterval(() => onResync(), resyncEveryMs);
				}
			}

			if (stopped) return;
			clearRetry();
			retryTimer = setTimeout(open, retryDelay(failures));
		};

		// Data reaching the page is NOT on its own evidence that the connection is
		// healthy, and treating it that way defeated the whole notice. On a
		// multiplexed route the proxy relays the working upstreams while it is
		// still validating the others — so a broken scoreboard arrives as: a few
		// time ticks, then the error frame, then close. Clearing the streak on
		// that first tick meant the streak never grew past one, the page
		// reconnected every second forever, and nothing was ever shown. Only the
		// settle window says healthy.
		for (const [name, handler] of Object.entries(on)) {
			es.addEventListener(name, (event) => {
				if (!mine()) return;
				handler(event as MessageEvent);
			});
		}

		es.addEventListener('upstream-error', (event) => {
			let stream = 'unknown';
			let status = 0;
			try {
				const parsed = JSON.parse((event as MessageEvent).data);
				stream = String(parsed.stream);
				status = Number(parsed.status);
			} catch {
				// Keep the placeholders; a malformed report is still a report.
			}
			fail({ stream, status, terminal: isTerminalStatus(status) });
		});

		// Opening proves only that the proxy answered `200 text/event-stream`. It
		// says nothing about the upstreams — those are validated afterwards, and
		// the heartbeat that arrives first is sent before that validation too. But
		// a bad upstream is reported within milliseconds, so a connection that has
		// been open this long without a word is as healthy as we can know it to
		// be. It is also the only recovery signal a silent stream will ever
		// produce: judge-workers and clarification deliver nothing at all until
		// something changes upstream, so waiting for data would leave the notice
		// up forever.
		//
		// The honest limit: an upstream that accepted the connection and then went
		// quiet is indistinguishable from a healthy silent one, and our own 15s
		// heartbeat keeps the polyfill's 45s watchdog from noticing either.
		es.addEventListener('open', () => {
			if (!mine()) return;
			opens += 1;
			// Pull the data back over REST on every RE-open. The first open needs
			// no resync — whatever mounted this stream fetched its own data — but
			// every one after it follows a gap the stream will not fill in, since
			// the backend pushes only on change. This fires for the polyfill's own
			// silent reconnects too, which is the point: those are the ones nothing
			// else here even sees.
			if (opens > 1) onResync?.();
			clearSettle();
			settleTimer = setTimeout(() => {
				if (!mine()) return;
				noteHealthy();
			}, settleAfterMs);
		});

		es.onerror = (event) => {
			onError?.(event as Event, stop);
			if (!mine()) return;
			// `status` is present only on the polyfill's ConnectionEvent, i.e. the
			// INITIAL response was not `200 text/event-stream` and it has already
			// closed itself for good. Anything else is a mid-stream drop it will
			// reopen by itself — supervising that would double the connections.
			// Note the test is "is there a status", not "is it 200": a wrong
			// content-type on a 200 lands here too.
			const status = (event as MessageEvent & { status?: number }).status;
			if (status === undefined) return;
			fail({ stream: 'connection', status, terminal: isTerminalStatus(status) });
		};
	};

	open();
	return stop;
}
