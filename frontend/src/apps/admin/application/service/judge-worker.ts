import type {
	CreateJudgeWorkerPayload,
	JudgeWorker,
	SystemDefaults
} from '@/apps/admin/domain/judge-worker';

export interface JudgeWorkerServiceInterface {
	fetchJudgeWorkers(): Promise<JudgeWorker[]>;
	fetchJudgePriorityStrategies(): Promise<string[]>;
	fetchSystemDefaults(): Promise<SystemDefaults>;
	createJudgeWorker(payload: CreateJudgeWorkerPayload): Promise<void>;
	deleteJudgeWorker(id: string): Promise<void>;
	startJudgeWorker(id: string): Promise<void>;
	stopJudgeWorker(id: string): Promise<void>;
}
