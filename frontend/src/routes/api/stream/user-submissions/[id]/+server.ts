import { type RequestHandler } from '@sveltejs/kit';
import { createAbortController, createSseStream, sseResponse } from '$lib/server/sse-proxy';
import { startBackendFetches } from '$lib/server/backend';

export const GET: RequestHandler = async ({ params, request, platform }) => {
	const { id } = params;
	const abortCtl = createAbortController(request, platform);

	const responses = startBackendFetches(
		{ submission: `/user/stream-submissions/${id}` },
		request,
		abortCtl.signal
	);

	// Deliberately do NOT resolveUpstream() here: the backend only flushes its
	// response headers on the first submission-update event, so awaiting them
	// would leave the client request pending forever on an already-judged
	// submission. createSseStream emits a heartbeat immediately, so the client
	// connection opens right away; backend events are piped through as they
	// arrive (and while idle the heartbeat keeps the connection alive).
	//
	// A consequence is that this route cannot forward a renewed auth cookie the
	// way the resolveUpstream()-based routes do (its headers are already sent by
	// the time the backend renews). That is fine: the backend renews without
	// rotating the refresh token, so each reconnect re-renews server-side from
	// the still-valid refresh cookie — the stream survives regardless.
	const handlers = {
		submission: (_eventName: string, ctrl: ReadableStreamDefaultController, chunk: Uint8Array) => {
			ctrl.enqueue(chunk);
		}
	};

	return sseResponse(createSseStream(abortCtl, `user-submissions/${id}`, responses, handlers));
};
