import { beforeEach, describe, expect, test, vi } from 'vitest';

vi.mock('$lib/server/backend', () => ({ startBackendFetches: vi.fn() }));

import { startBackendFetches } from '$lib/server/backend';
import { GET as judgeWorkers } from './judge-workers/+server';
import { GET as judgeWorkerLog } from './judge-workers/[id]/+server';
import { GET as buildLog } from './images/[tag]/builds/[batch]/+server';
import { GET as projectSubmissions } from './contest-projects/[id]/submissions/+server';
import { GET as userSubmission } from './user-submissions/[id]/+server';

const enc = new TextEncoder();
const mocked = vi.mocked(startBackendFetches);

/** An upstream 200 whose body arrives in exactly these pieces. */
function upstream(...pieces: string[]): Promise<Response> {
	return Promise.resolve(
		new Response(
			new ReadableStream({
				start(c) {
					for (const piece of pieces) c.enqueue(enc.encode(piece));
					c.close();
				}
			}),
			{ status: 200 }
		)
	);
}

/** An upstream 200 whose single frame is cut into two chunks at a byte offset. */
function latin1Upstream(text: string, at: number): Promise<Response> {
	const bytes = enc.encode(text);
	return Promise.resolve(
		new Response(
			new ReadableStream({
				start(c) {
					c.enqueue(bytes.subarray(0, at));
					c.enqueue(bytes.subarray(at));
					c.close();
				}
			}),
			{ status: 200 }
		)
	);
}

const call = (
	handler: (event: never) => Promise<Response> | Response,
	params: Record<string, string> = {}
) =>
	handler({
		params,
		request: new Request('http://localhost/api/stream/x'),
		platform: undefined
	} as never);

async function bodyOf(response: Response | Promise<Response>): Promise<string> {
	return await (await response).text();
}

/** Field names present across all frames, in order — what the browser keys on. */
function eventNames(sse: string): string[] {
	return [...sse.matchAll(/^event: (.+)$/gm)].map((m) => m[1]);
}

beforeEach(() => mocked.mockReset());

describe('reframed routes keep their event names', () => {
	test('judge-workers/[id] labels worker logs `message`, whole', async () => {
		const payload = { log: 'x'.repeat(5000), timestamp: 't' };
		const frame = `data: ${JSON.stringify(payload)}\n\n`;
		mocked.mockReturnValue({ message: latin1Upstream(frame, 4096) });

		const sse = await bodyOf(call(judgeWorkerLog, { id: '1' }));

		expect(eventNames(sse)).toEqual(['message']);
		expect(JSON.parse(/^data: (.+)$/m.exec(sse)![1]).log).toBe(payload.log);
	});

	test('images build log labels frames `log` with Chinese intact', async () => {
		const payload = { level: 'error', message: '建置失敗:找不到組態檔' };
		const frame = `data: ${JSON.stringify(payload)}\n\n`;
		const midCharacter = enc.encode(frame).indexOf(enc.encode('建')[0]) + 1;
		mocked.mockReturnValue({ log: latin1Upstream(frame, midCharacter) });

		const sse = await bodyOf(call(buildLog, { tag: 'demo', batch: '1' }));

		expect(eventNames(sse)).toEqual(['log']);
		expect(JSON.parse(/^data: (.+)$/m.exec(sse)![1]).message).toBe(payload.message);
	});

	test('judge-workers list labels frames `message`', async () => {
		mocked.mockReturnValue({ message: upstream('data: {"workers":[]}\n\n') });

		const sse = await bodyOf(call(judgeWorkers));

		expect(eventNames(sse)).toEqual(['message']);
	});
});

describe('passthrough routes stay unlabelled', () => {
	// Both consumers listen on the DEFAULT message event. Adding an `event:`
	// line here — "for consistency" with the reframed routes — would silently
	// stop them receiving anything.
	test('contest-projects/[id]/submissions relays bytes verbatim', async () => {
		mocked.mockReturnValue({ submission: upstream('data: {"id":"a"}\n\n') });

		const sse = await bodyOf(call(projectSubmissions, { id: '1' }));

		expect(eventNames(sse)).toEqual([]);
		expect(sse).toContain('data: {"id":"a"}\n\n');
	});

	test('user-submissions/[id] relays bytes verbatim', async () => {
		mocked.mockReturnValue({ submission: upstream('data: {"id":"a"}\n\n') });

		const sse = await bodyOf(call(userSubmission, { id: 'a' }));

		expect(eventNames(sse)).toEqual([]);
		expect(sse).toContain('data: {"id":"a"}\n\n');
	});
});

describe('routes fronting a silent upstream open immediately', () => {
	// These backends write nothing until an event occurs, and gin withholds the
	// response headers until the first body byte — so awaiting the upstream left
	// the browser's request hanging on an idle system.
	test.each([
		['judge-workers', () => call(judgeWorkers), 'message'],
		['contest-projects/[id]/submissions', () => call(projectSubmissions, { id: '1' }), 'submission']
	])('%s answers before the upstream does', async (_name, invoke, key) => {
		mocked.mockReturnValue({ [key]: new Promise<Response>(() => {}) });

		const response = await Promise.race([
			invoke(),
			new Promise<never>((_, reject) => setTimeout(() => reject(new Error('hung')), 1000))
		]);

		expect(response.status).toBe(200);
		expect(response.headers.get('Content-Type')).toBe('text/event-stream');
		const { value } = await response.body!.getReader().read();
		expect(new TextDecoder().decode(value)).toBe(': keepalive\n\n');
	});
});
