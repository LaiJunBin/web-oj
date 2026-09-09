import { describe, expect, test, vi } from 'vitest';
import {
	createEventReframer,
	createSseStream,
	resolveUpstream,
	UPSTREAM_ERROR_EVENT,
	UPSTREAM_UNREACHABLE_STATUS,
	type ChunkHandler
} from '$lib/server/sse-proxy';

const enc = new TextEncoder();

/**
 * Parses an SSE byte stream the way EventSource does, so the assertions below
 * are about what the *browser* ends up with rather than what we happened to
 * enqueue. Frames whose data buffer stays empty are dropped, matching the spec
 * (and event-source-polyfill): a frame that is nothing but `event:` or a
 * comment never reaches a listener.
 */
function parseSse(text: string): { name: string; data: string }[] {
	const events: { name: string; data: string }[] = [];
	for (const block of text.split('\n\n')) {
		if (block === '') continue;
		let name = 'message';
		const data: string[] = [];
		for (const line of block.split('\n')) {
			if (line === '' || line.startsWith(':')) continue;
			const i = line.indexOf(':');
			const field = i === -1 ? line : line.slice(0, i);
			let value = i === -1 ? '' : line.slice(i + 1);
			if (value.startsWith(' ')) value = value.slice(1);
			if (field === 'event') name = value;
			else if (field === 'data') data.push(value);
		}
		if (data.length === 0) continue;
		events.push({ name, data: data.join('\n') });
	}
	return events;
}

/** Feeds bytes to a handler, cut at the given offsets, and parses what comes out. */
function through(handler: ChunkHandler, eventName: string, bytes: Uint8Array, splits: number[]) {
	const out: Uint8Array[] = [];
	const ctrl = {
		enqueue: (c: Uint8Array) => {
			out.push(c);
		}
	} as unknown as ReadableStreamDefaultController;
	let prev = 0;
	for (const at of [...splits, bytes.length]) {
		handler(eventName, ctrl, bytes.subarray(prev, at));
		prev = at;
	}
	return parseSse(new TextDecoder().decode(Buffer.concat(out.map((c) => Buffer.from(c)))));
}

/** A Response whose body delivers the given chunks, then ends. */
function streamingResponse(chunks: string[], status = 200): Response {
	return new Response(
		new ReadableStream({
			start(c) {
				for (const chunk of chunks) c.enqueue(enc.encode(chunk));
				c.close();
			}
		}),
		{ status }
	);
}

async function drain(stream: ReadableStream): Promise<string> {
	return await new Response(stream).text();
}

describe('createEventReframer', () => {
	const frame = (payload: object) => enc.encode(`data: ${JSON.stringify(payload)}\n\n`);

	test('survives a cut at every possible byte boundary', () => {
		const log = { log: 'starting judgment', timestamp: 't' };
		const bytes = frame(log);
		const broken: number[] = [];
		for (let at = 1; at < bytes.length; at++) {
			const events = through(createEventReframer(), 'message', bytes, [at]);
			const ok =
				events.length === 1 &&
				events[0].name === 'message' &&
				JSON.parse(events[0].data).log === log.log;
			if (!ok) broken.push(at);
		}
		// The naive `event: name\n` + raw chunk this replaced failed 49 of these
		// 50 offsets — every cut except the one landing after the blank line.
		expect(broken).toEqual([]);
	});

	test('keeps a multi-byte character split across chunks intact', () => {
		const payload = { level: 'error', message: '建置失敗:找不到組態檔' };
		const bytes = frame(payload);
		const midCharacter = bytes.indexOf(enc.encode('建')[0]) + 1;

		const events = through(createEventReframer(), 'log', bytes, [midCharacter]);

		expect(events).toEqual([{ name: 'log', data: JSON.stringify(payload) }]);
	});

	test('reassembles a frame larger than the backend copy buffer', () => {
		// The generated Go handler copies through a 4096-byte buffer and flushes
		// each read, so a frame this size is guaranteed to arrive in pieces.
		const payload = { log: 'x'.repeat(5000) };
		const bytes = frame(payload);
		expect(bytes.length).toBeGreaterThan(4096);

		const events = through(createEventReframer(), 'message', bytes, [4096]);

		expect(events).toHaveLength(1);
		expect(JSON.parse(events[0].data).log).toBe(payload.log);
	});

	test('splits several frames delivered in a single chunk', () => {
		const bytes = enc.encode('data: {"n":1}\n\ndata: {"n":2}\n\n');

		const events = through(createEventReframer(), 'time', bytes, []);

		expect(events).toEqual([
			{ name: 'time', data: '{"n":1}' },
			{ name: 'time', data: '{"n":2}' }
		]);
	});

	test('withholds a frame that never completes rather than emitting a partial', () => {
		const bytes = enc.encode('data: {"n":1}\n\ndata: {"trunc');

		const events = through(createEventReframer(), 'time', bytes, []);

		expect(events).toEqual([{ name: 'time', data: '{"n":1}' }]);
	});
});

describe('createSseStream', () => {
	test('opens with a heartbeat and relays every upstream', async () => {
		const abortCtl = new AbortController();
		const responses = {
			time: Promise.resolve(streamingResponse(['data: {"left":"1h"}\n\n'])),
			scoreboard: Promise.resolve(streamingResponse(['data: [1,2]\n\n']))
		};

		const text = await drain(
			createSseStream(abortCtl, 'test', responses, {
				time: createEventReframer(),
				scoreboard: createEventReframer()
			})
		);

		expect(text.startsWith(': keepalive\n\n')).toBe(true);
		expect(parseSse(text)).toEqual(
			expect.arrayContaining([
				{ name: 'time', data: '{"left":"1h"}' },
				{ name: 'scoreboard', data: '[1,2]' }
			])
		);
		expect(abortCtl.signal.aborted).toBe(false);
	});

	test('reports a late non-200 upstream instead of letting it die in silence', async () => {
		const abortCtl = new AbortController();
		const responses: Record<string, Promise<Response>> = {
			time: Promise.resolve(streamingResponse(['data: {"left":"1h"}\n\n'])),
			// Resolves after the race is already decided by `time`, which is the
			// hole this covers: resolveUpstream only ever checked the winner.
			scoreboard: new Promise((resolve) =>
				setTimeout(
					() =>
						resolve(new Response('{"success":false,"message":"Unauthorized"}', { status: 401 })),
					10
				)
			)
		};

		const upstream = await resolveUpstream(responses, abortCtl);
		expect(upstream.ok, 'the race is won by the healthy stream').toBe(true);

		const events = parseSse(
			await drain(
				createSseStream(abortCtl, 'test', responses, {
					time: createEventReframer(),
					scoreboard: createEventReframer()
				})
			)
		);

		expect(events).toContainEqual({
			name: UPSTREAM_ERROR_EVENT,
			data: JSON.stringify({ stream: 'scoreboard', status: 401 })
		});
		// The error body must never be relayed as if it were stream content.
		expect(events.some((e) => e.data.includes('Unauthorized'))).toBe(false);
		expect(abortCtl.signal.aborted, 'the whole connection is taken down').toBe(true);
	});

	test('several failing upstreams are announced once, but all are logged', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		const abortCtl = new AbortController();
		const responses: Record<string, Promise<Response>> = {
			time: Promise.resolve(streamingResponse(['data: {"left":"1h"}\n\n'])),
			scoreboard: Promise.resolve(new Response('{}', { status: 401 })),
			clarification: Promise.resolve(new Response('{}', { status: 403 }))
		};

		const events = parseSse(
			await drain(
				createSseStream(abortCtl, 'test', responses, {
					time: createEventReframer(),
					scoreboard: createEventReframer(),
					clarification: createEventReframer()
				})
			)
		);

		// The client tears the whole connection down on the first report, so a
		// second frame would at best be ignored and at worst counted twice.
		const reports = events.filter((e) => e.name === UPSTREAM_ERROR_EVENT);
		expect(reports, 'one announcement').toHaveLength(1);
		expect(
			warn.mock.calls.filter((c) => String(c[0]).includes('answered')),
			'both logged'
		).toHaveLength(2);
		warn.mockRestore();
	});

	test('does not announce an error when every upstream is healthy', async () => {
		const abortCtl = new AbortController();
		const events = parseSse(
			await drain(
				createSseStream(
					abortCtl,
					'test',
					{ time: Promise.resolve(streamingResponse(['data: {"left":"1h"}\n\n'])) },
					{ time: createEventReframer() }
				)
			)
		);

		expect(events.map((e) => e.name)).toEqual(['time']);
	});
});

describe('an upstream fetch that never produces a Response', () => {
	const reports = (text: string) => parseSse(text).filter((e) => e.name === UPSTREAM_ERROR_EVENT);

	/** A body that starts fine and then errors, the way a reset socket behaves. */
	function bodyThatBreaks(): Response {
		return new Response(
			new ReadableStream({
				start(c) {
					c.enqueue(enc.encode('data: {"n":1}\n\n'));
					setTimeout(() => c.error(new Error('socket hang up')), 5);
				}
			}),
			{ status: 200 }
		);
	}

	test('a refused connection is announced instead of looking healthy', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		const abortCtl = new AbortController();

		// Before this, the browser got exactly ": keepalive\n\n" — a clean 200 and
		// EOF it could not tell apart from a working stream, so it reconnected
		// once a second forever with nothing on screen to say why.
		const text = await drain(
			createSseStream(
				abortCtl,
				'test',
				{ message: Promise.reject(new TypeError('fetch failed')) },
				{ message: createEventReframer() }
			)
		);

		expect(reports(text)).toEqual([
			{
				name: UPSTREAM_ERROR_EVENT,
				data: JSON.stringify({ stream: 'message', status: UPSTREAM_UNREACHABLE_STATUS })
			}
		]);
		expect(abortCtl.signal.aborted).toBe(true);
		warn.mockRestore();
	});

	test('a body that breaks after it started is announced too', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		const abortCtl = new AbortController();

		const text = await drain(
			createSseStream(
				abortCtl,
				'test',
				{ message: Promise.resolve(bodyThatBreaks()) },
				{ message: createEventReframer() }
			)
		);

		expect(parseSse(text)).toContainEqual({ name: 'message', data: '{"n":1}' });
		expect(reports(text)).toHaveLength(1);
		expect(reports(text)[0].data).toContain(String(UPSTREAM_UNREACHABLE_STATUS));
		warn.mockRestore();
	});

	test('several failing upstreams still announce once, and the siblings are cut off', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		const abortCtl = new AbortController();
		let siblingChunks = 0;

		// A sibling that would keep producing forever if nobody stopped it.
		const sibling = new ReadableStream({
			start(c) {
				const t = setInterval(() => {
					try {
						c.enqueue(enc.encode('data: {"n":1}\n\n'));
						siblingChunks++;
					} catch {
						clearInterval(t);
					}
				}, 10);
				setTimeout(() => clearInterval(t), 500);
			}
		});

		const text = await drain(
			createSseStream(
				abortCtl,
				'test',
				{
					time: Promise.resolve(new Response(sibling, { status: 200 })),
					scoreboard: Promise.reject(new TypeError('fetch failed')),
					clarification: Promise.reject(new TypeError('fetch failed'))
				},
				{
					time: createEventReframer(),
					scoreboard: createEventReframer(),
					clarification: createEventReframer()
				}
			)
		);

		expect(reports(text), 'one announcement for the whole connection').toHaveLength(1);
		expect(abortCtl.signal.aborted).toBe(true);
		const atClose = siblingChunks;
		await new Promise((r) => setTimeout(r, 200));
		expect(siblingChunks, 'the sibling stopped rather than reading on detached').toBe(atClose);
		warn.mockRestore();
	});

	test('our own teardown is not mistaken for a fault', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		const abortCtl = new AbortController();
		const aborted = new DOMException('This operation was aborted', 'AbortError');

		const text = await drain(
			createSseStream(
				abortCtl,
				'test',
				{ message: Promise.reject(aborted) },
				{ message: createEventReframer() }
			)
		);

		expect(reports(text), 'reporting this would invent a fault').toEqual([]);
		warn.mockRestore();
	});

	test('resolveUpstream cuts the siblings loose before answering 500', async () => {
		const abortCtl = new AbortController();

		await expect(
			resolveUpstream(
				{
					a: Promise.reject(new TypeError('fetch failed')),
					b: new Promise<Response>(() => {})
				},
				abortCtl
			)
		).rejects.toThrow();

		expect(abortCtl.signal.aborted, 'the pending sibling must not be left reading').toBe(true);
	});
});
