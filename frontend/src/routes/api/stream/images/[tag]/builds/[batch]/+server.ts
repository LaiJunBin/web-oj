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
	const { tag, batch } = params;
	const abortCtl = createAbortController(request, platform);

	const requests = {
		log: `/images/${tag}/builds/${batch}/stream-logs`
	};

	const responses = startBackendFetches(requests, request, abortCtl.signal);

	const upstream = await resolveUpstream(responses, abortCtl);
	if (!upstream.ok) return upstream.clientResponse;

	// The reframer, not a per-line label: decoding each chunk on its own turned
	// a multi-byte character split across chunks into mojibake, and labelling
	// every line meant an arbitrary chunk boundary could cut a `data:` line in
	// two and label the second half as a fresh field. Build logs are exactly
	// where that bites — long lines, and Chinese in the output. The event name
	// stays `log`.
	const handlers = {
		log: createEventReframer()
	};

	return sseResponse(
		createSseStream(abortCtl, `images/${tag}/builds/${batch}`, responses, handlers),
		upstream.setCookies
	);
};
