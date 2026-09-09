import { type RequestHandler } from '@sveltejs/kit';
import {
	createAbortController,
	createEventReframer,
	createSseStream,
	sseResponse
} from '$lib/server/sse-proxy';
import { startBackendFetches } from '$lib/server/backend';

export const GET: RequestHandler = async ({ request, platform }) => {
	const abortCtl = createAbortController(request, platform);

	const responses = startBackendFetches(
		{ message: `/judge-workers/stream` },
		request,
		abortCtl.signal
	);

	// Deliberately do NOT resolveUpstream() here: this backend stream is purely
	// event-driven (it writes nothing until a worker changes state), and gin
	// holds the response headers back until the first body byte — so awaiting
	// the response left the browser's request hanging on an idle system. Same
	// trap, same treatment as user-submissions/[id]: createSseStream emits a
	// heartbeat immediately, so the connection opens at once and backend events
	// flow through as they arrive.
	//
	// The cost is the same one that route already accepts: no renewed auth
	// cookie can be forwarded (headers are long gone by the time the backend
	// renews), and an upstream error can no longer be answered with its real
	// status. createSseStream validates the upstream instead and reports it as
	// an upstream-error event; the backend renews without rotating the refresh
	// token, so each reconnect re-renews server-side and the stream survives.
	const handlers = {
		message: createEventReframer()
	};

	return sseResponse(createSseStream(abortCtl, 'judge-workers', responses, handlers));
};
