import type {
	JudgeQueueStatus,
	JudgmentError,
	SubmissionRatePoint
} from '@/apps/admin/domain/system-status';

export interface SystemStatusServiceInterface {
	fetchJudgeQueue(): Promise<JudgeQueueStatus>;
	fetchJudgmentErrors(limit: number): Promise<JudgmentError[]>;
	fetchSubmissionRate(minutes: number, contestProjectId?: number): Promise<SubmissionRatePoint[]>;
}
