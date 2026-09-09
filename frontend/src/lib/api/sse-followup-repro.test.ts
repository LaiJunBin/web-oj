/**
 * The SSE supervisor, exercised against the REAL event-source-polyfill over a
 * real HTTP server.
 *
 * The unit tests next door drive a hand-written fake, which is fast and precise
 * but cannot tell you what the polyfill actually does — and the three problems
 * this file covers were all mis-diagnosed by reading its source. Only running it
 * settles them: that a connect-time non-200 makes it close *terminally* rather
 * than retry, that it stops parsing a chunk once a listener closes it, and that
 * a silent-but-healthy reconnect produces no event of any kind.
 */
import { afterEach, describe, expect, test, vi } from 'vitest';
import { createServer, type Server, type ServerResponse } from 'node:http';
import { EventSourcePolyfill } from 'event-source-polyfill';
import { createManagedStream } from './managed-stream';

const servers: Server[] = [];
afterEach(() => {
	for (const s of servers.splice(0)) s.close();
});

async function serve(handler: (res: ServerResponse, hit: number) => void): Promise<string> {
	let hits = 0;
	const server = createServer((_req, res) => handler(res, ++hits));
	servers.push(server);
	await new Promise<void>((resolve) => server.listen(0, '127.0.0.1', resolve));
	const addr = server.address() as { port: number };
	return `http://127.0.0.1:${addr.port}/`;
}
function hitCounter() {
	let n = 0;
	return { bump: () => ++n, get: () => n };
}
const wait = (ms: number) => new Promise((r) => setTimeout(r, ms));
/**
 * Opens the response the way the proxy does: headers plus an immediate heartbeat.
 * The heartbeat is not decoration — Node holds the headers back until something
 * is written, so without it the browser never sees the connection open at all.
 */
const sse = (res: ServerResponse) => {
	res.writeHead(200, { 'Content-Type': 'text/event-stream' });
	res.write(': keepalive\n\n');
};
const errorFrame = (stream: string, status: number) =>
	`event: upstream-error\ndata: ${JSON.stringify({ stream, status })}\n\n`;

describe('connect-time failures reach the supervisor', () => {
	test('a 500 is retried on the backoff schedule and reported', async () => {
		// Left to itself the polyfill closes terminally here — one request, no
		// retry, nothing on screen. That is what freezes a countdown for good when
		// the backend is restarted mid-contest.
		const hits = hitCounter();
		const url = await serve((res) => {
			hits.bump();
			res.writeHead(500, { 'Content-Type': 'application/json' });
			res.end('{"message":"boom"}');
		});

		const onDegraded = vi.fn();
		const stop = createManagedStream({
			connect: () => new EventSourcePolyfill(url),
			on: { time: () => {} },
			onDegraded
		});

		await wait(3_500);
		stop();

		expect(hits.get(), 'retried at 1s and 2s, not abandoned after one try').toBeGreaterThanOrEqual(
			3
		);
		expect(onDegraded).toHaveBeenCalledWith(
			expect.objectContaining({ stream: 'connection', status: 500, terminal: false })
		);
	}, 20_000);

	test('a 404 is reported at once and never retried', async () => {
		// This is a contest whose activation window closed, or a scoreboard the
		// organiser un-published. Retrying would put an alarming banner on a venue
		// screen for a contest that simply ended.
		const hits = hitCounter();
		const url = await serve((res) => {
			hits.bump();
			res.writeHead(404, { 'Content-Type': 'application/json' });
			res.end('{"message":"not found"}');
		});

		const onDegraded = vi.fn();
		const stop = createManagedStream({
			connect: () => new EventSourcePolyfill(url),
			on: { time: () => {} },
			onDegraded
		});

		await wait(3_500);
		stop();

		expect(hits.get(), 'exactly one attempt').toBe(1);
		expect(onDegraded).toHaveBeenCalledWith(
			expect.objectContaining({ status: 404, terminal: true })
		);
	}, 20_000);
});

describe('simultaneous upstream failures', () => {
	test('two error frames on one connection produce a single retry', async () => {
		const hits = hitCounter();
		const url = await serve((res, n) => {
			hits.bump();
			sse(res);
			if (n === 1) {
				res.write(errorFrame('scoreboard', 500));
				res.write(errorFrame('clarification', 502));
			}
			// Later connections stay open and silent.
		});

		const stop = createManagedStream({
			connect: () => new EventSourcePolyfill(url),
			on: { time: () => {} }
		});

		await wait(2_500);
		stop();

		// A second counted failure would have armed a second timer and opened a
		// third connection within this window.
		expect(hits.get(), 'the initial connection plus exactly one retry').toBe(2);
	}, 20_000);
});

describe('recovery on a silent stream', () => {
	test('a healthy reconnect clears the notice even when no data ever arrives', async () => {
		// judge-workers and clarification deliver nothing until something changes
		// upstream, so waiting for a named event would leave the notice up forever.
		const url = await serve((res, n) => {
			sse(res);
			if (n <= 3) res.write(errorFrame('message', 500));
			// From the 4th connection the upstream is healthy — and says nothing.
		});

		let degraded = false;
		const stop = createManagedStream({
			connect: () => new EventSourcePolyfill(url),
			on: { message: () => {} },
			onDegraded: () => (degraded = true),
			onRecovered: () => (degraded = false),
			settleAfterMs: 1_000
		});

		// 1s + 2s + 4s of backoff, then the healthy-but-silent connection settles.
		await wait(9_000);
		stop();

		expect(degraded, 'the connection has been open and quiet long enough').toBe(false);
	}, 25_000);
});

describe('an upstream the proxy cannot reach at all', () => {
	test('the browser backs off and says so, instead of hammering in silence', async () => {
		// This is what a backend restarted mid-contest produces: the proxy's fetch
		// rejects, so it can only report a synthetic 502. Before that report
		// existed the browser saw 200 + heartbeat + clean EOF — indistinguishable
		// from a healthy stream — and reconnected once a second, forever, with
		// nothing on screen.
		const hits = hitCounter();
		const url = await serve((res) => {
			hits.bump();
			sse(res);
			res.write(errorFrame('message', 502));
		});

		const onDegraded = vi.fn();
		const stop = createManagedStream({
			connect: () => new EventSourcePolyfill(url),
			on: { message: () => {} },
			onDegraded
		});

		await wait(8_000);
		stop();

		// Unsupervised this would be ~8 connections; backing off 1s/2s/4s gives 4.
		expect(hits.get(), 'backed off rather than once a second').toBeLessThanOrEqual(5);
		expect(onDegraded).toHaveBeenCalledWith(
			expect.objectContaining({ status: 502, terminal: false })
		);
	}, 25_000);
});
