import { type RequestHandler } from '@sveltejs/kit';
import {
	createAbortController,
	createEventReframer,
	createSseStream,
	resolveUpstream,
	sseResponse
} from '$lib/server/sse-proxy';
import { startBackendFetches } from '$lib/server/backend';

// Aggregates the three system-status streams into one browser connection, each
// labelled with its own event name (the contest proxies' multiplex pattern).
// The parameters are fixed here on purpose: the dashboard is the only consumer,
// and the backend answers 400 at connect time for a rejected parameter — so a
// bad value would fail the whole aggregate loudly instead of silently.
export const GET: RequestHandler = async ({ request, platform }) => {
	const abortCtl = createAbortController(request, platform);

	const requests = {
		'judge-queue': `/judge-queue/stream`,
		'judgment-errors': `/judgment-errors/stream?limit=10`,
		'submission-rate': `/submission-rate/stream?minutes=10`
	};

	const responses = startBackendFetches(requests, request, abortCtl.signal);

	const upstream = await resolveUpstream(responses, abortCtl);
	if (!upstream.ok) return upstream.clientResponse;

	const handlers = {
		'judge-queue': createEventReframer(),
		'judgment-errors': createEventReframer(),
		'submission-rate': createEventReframer()
	};

	return sseResponse(
		createSseStream(abortCtl, 'system-status', responses, handlers),
		upstream.setCookies
	);
};
