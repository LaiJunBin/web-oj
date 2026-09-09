import { EventSourcePolyfill } from 'event-source-polyfill';

export const streamPublicContest = (id: number) =>
	new EventSourcePolyfill(`/api/stream/public-contests/${id}`);
export const streamUserContest = (id: number) =>
	new EventSourcePolyfill(`/api/stream/user-contests/${id}`);
export const streamBuildImageLog = (tag: string, batch: number) =>
	new EventSourcePolyfill(`/api/stream/images/${tag}/builds/${batch}`);
export const streamJudgeWorkers = () => new EventSourcePolyfill(`/api/stream/judge-workers`);
export const streamSystemStatus = () => new EventSourcePolyfill(`/api/stream/system-status`);
export const streamJudgeWorkerLog = (id: number) =>
	new EventSourcePolyfill(`/api/stream/judge-workers/${id}`);
export const streamContestProjectSubmissions = (id: number) =>
	new EventSourcePolyfill(`/api/stream/contest-projects/${id}/submissions`);
export const streamUserSubmission = (id: string) =>
	new EventSourcePolyfill(`/api/stream/user-submissions/${id}`);
