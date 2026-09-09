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
		time: `/user/contest-projects/${id}/stream-time`,
		scoreboard: `/user/contest-projects/${id}/stream-scoreboard`,
		clarification: `/user/contest-projects/${id}/stream-clarification`
	};

	const responses = startBackendFetches(requests, request, abortCtl.signal);

	const upstream = await resolveUpstream(responses, abortCtl);
	if (!upstream.ok) return upstream.clientResponse;

	// Each stream gets its own reframer so a large scoreboard message split across
	// transport chunks is reassembled before being labelled (see createEventReframer).
	const handlers = {
		time: createEventReframer(),
		scoreboard: createEventReframer(),
		clarification: createEventReframer()
	};

	return sseResponse(
		createSseStream(abortCtl, `user-contests/${id}`, responses, handlers),
		upstream.setCookies
	);
};
