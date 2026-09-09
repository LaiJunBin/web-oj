import { type RequestHandler } from '@sveltejs/kit';
import { createAbortController, createSseStream, sseResponse } from '$lib/server/sse-proxy';
import { startBackendFetches } from '$lib/server/backend';

export const GET: RequestHandler = async ({ params, request, platform }) => {
	const { id } = params;
	const abortCtl = createAbortController(request, platform);

	const responses = startBackendFetches(
		{ submission: `/contest-projects/${id}/stream-submissions` },
		request,
		abortCtl.signal
	);

	// Deliberately do NOT resolveUpstream() here: the backend only writes on a
	// submission update and gin withholds the response headers until the first
	// body byte, so awaiting the response left this request hanging until
	// somebody submitted. Same trap and same treatment as user-submissions/[id]
	// — createSseStream heartbeats immediately, so the connection opens at once.
	// The trade-off (no renewed cookie forwarded, no connect-time status) is
	// covered there; createSseStream still reports an upstream error as an
	// upstream-error event.
	//
	// The chunk is forwarded byte for byte, with no `event:` label. That is not
	// an oversight and must not be "made consistent" with the reframed routes:
	// a byte-identical relay cannot corrupt framing, and the consumer
	// (dashboard/submissions) listens on the DEFAULT message event — labelling
	// these frames would stop it receiving anything at all.
	const handlers = {
		submission: (_eventName: string, ctrl: ReadableStreamDefaultController, chunk: Uint8Array) => {
			ctrl.enqueue(chunk);
		}
	};

	return sseResponse(
		createSseStream(abortCtl, `contest-projects/${id}/submissions`, responses, handlers)
	);
};
