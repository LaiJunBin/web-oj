import {
	type HttpResponseResolver,
	type PathParams,
	type DefaultBodyType,
	delay,
	HttpHandler,
	http,
	type StrictResponse,
	HttpResponse
} from 'msw';
import { apiClient } from '@/api';
import { handlers } from './handlers';
import { login } from './handlers/login';
import { streamUserProjectTime } from './handlers/stream-user-project-time';
import { streamUserProjectClarification } from './handlers/stream-user-project-clarification';
import { streamUserProjectScoreboard } from './handlers/stream-user-project-scoreboard';
import { streamLogs } from './handlers/stream-logs';
import { streamJudgeWorkerLogs } from './handlers/stream-judge-worker-logs';
import { getProfile } from '@/mocks/handlers/get-profile';

const baseURL = import.meta.env.VITE_API_URL;

const API: Record<(typeof apiClient)['api'][number]['alias'], string> = Object.fromEntries(
	Object.keys(apiClient)
		.map(() =>
			(
				apiClient.api as Array<{
					method: string;
					path: string;
					alias: string;
				}>
			).map((api) => [api.alias, `${api.method.toUpperCase()} ${api.path}`])
		)
		.flat()
);

export type StrictResponseOf<T extends (...args: any) => any> = StrictResponse<
	Awaited<ReturnType<T>>
>;
const OverwriteHandlerMap: Record<
	string,
	HttpResponseResolver<PathParams, DefaultBodyType, undefined>
> = {
	[API.login]: login,
	[API.getProfile]: getProfile,
	[API.streamUserProjectTime]: streamUserProjectTime,
	[API.streamUserProjectClarification]: streamUserProjectClarification,
	[API.streamUserProjectScoreboard]: streamUserProjectScoreboard,

	[API.streamPublicProjectTime]: streamUserProjectTime,
	[API.streamPublicProjectScoreboard]: streamUserProjectScoreboard,

	[API.streamLogs]: streamLogs,
	[API.streamJudgeWorkerLogs]: streamJudgeWorkerLogs
};

const delayMiddleware = (resolver: HttpResponseResolver) => {
	return async (input: Parameters<HttpResponseResolver>[0]) => {
		await delay(500);
		return resolver(input);
	};
};

const applyMiddleware = (
	handler: HttpHandler,
	middleware: (resolver: HttpResponseResolver) => HttpResponseResolver
) => {
	const {
		info: { path, method },
		resolver
	} = handler as typeof handler & {
		info: {
			path: string;
			method: string;
		};
		resolver: HttpResponseResolver;
	};
	return http[method.toLowerCase() as keyof typeof http](path, middleware(resolver));
};

const applyDelayMiddleware = (handler: HttpHandler) => {
	return applyMiddleware(handler, delayMiddleware);
};

const enhancedHandlers = handlers
	.map((handler: HttpHandler) => {
		handler.info.path = baseURL + handler.info.path;
		if (OverwriteHandlerMap[handler.info.header]) {
			const { method } = handler.info;
			const resolver = OverwriteHandlerMap[handler.info.header];
			return http[(method as string).toLowerCase() as keyof typeof http](
				handler.info.path,
				resolver
			);
		} else {
			const { method } = handler.info;
			// msw marks `resolver` as protected, and its return type is a wide union
			// (Response | iterable | undefined). The generated handlers' resolvers all
			// synchronously return a JSON HttpResponse, so unwrap via one deliberate
			// cast, forward the resolver info, and re-wrap the body with success:true.
			const originalResolver = (
				handler as unknown as {
					resolver: HttpResponseResolver<PathParams, DefaultBodyType>;
				}
			).resolver;
			return http[(method as string).toLowerCase() as keyof typeof http](
				handler.info.path,
				(async (info) => {
					const response = (await originalResolver(info)) as Response | undefined;
					if (!response) {
						return HttpResponse.json({ success: false }, { status: 500 });
					}
					const json = await response.json();
					console.log(json);
					return HttpResponse.json({
						...json,
						success: true
					});
				}) as HttpResponseResolver
			);
		}
	})
	.map(applyDelayMiddleware);

export { enhancedHandlers as handlers };
