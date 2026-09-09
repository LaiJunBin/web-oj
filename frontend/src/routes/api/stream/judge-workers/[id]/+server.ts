import { type RequestHandler } from '@sveltejs/kit';
import {
	createAbortController,
	createEventReframer,
	createSseStream,
	resolveUpstream,
	sseResponse
} from '$lib/server/sse-proxy';
import { startBackendFetches } from '$lib/server/backend';

export const GET: RequestHandler = async ({ params, request, platform }) => {
	const { id } = params;
	const abortCtl = createAbortController(request, platform);

	const requests = {
		message: `/judge-workers/${id}/stream-logs`
	};

	const responses = startBackendFetches(requests, request, abortCtl.signal);

	const upstream = await resolveUpstream(responses, abortCtl);
	if (!upstream.ok) return upstream.clientResponse;

	// The reframer, not a raw `event: message\n` + chunk: a worker log frame
	// larger than the backend's 4096-byte copy buffer is split mid-message on
	// the wire, and the injected label would land inside a `data:` line (see
	// createEventReframer). The event name stays `message` — three consumers
	// listen for the default event.
	const handlers = {
		message: createEventReframer()
	};

	return sseResponse(
		createSseStream(abortCtl, `judge-workers/${id}`, responses, handlers),
		upstream.setCookies
	);
};
