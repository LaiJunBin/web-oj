import { Judge_workersApiSchemas } from '@/api/gen/Judge_workers';
import { JudgmentsApiSchemas } from '@/api/gen/Judgments';
import { SubmissionsApiSchemas } from '@/api/gen/Submissions';
import type { Storage } from '@/apps/common/domain/storage';
import { container } from '@/apps/admin/container';
import { z } from 'zod';

// 系統狀態頁的資料:評分佇列快照(GET + SSE)、最近評分錯誤、提交速率。

export type JudgeQueueStatus = z.infer<
	typeof Judge_workersApiSchemas.Admin_JudgeWorkerSpec_JudgeQueueStatus
>;
export type JudgmentError = z.infer<typeof JudgmentsApiSchemas.Admin_JudgmentSpec_JudgmentError>;
export type SubmissionRatePoint = z.infer<
	typeof SubmissionsApiSchemas.Admin_SubmissionSpec_SubmissionRatePoint
>;

// 各 stream 每則事件的內容(與對應 GET 端點相同的包裝)。三條在 BFF 聚合為
// /api/stream/system-status 的單一連線,以事件名稱區分。
export const JudgeQueueEventDataSchema = z
	.object({
		data: Judge_workersApiSchemas.Admin_JudgeWorkerSpec_JudgeQueueStatus,
		success: z.boolean().default(true)
	})
	.passthrough();

export const JudgmentErrorsEventDataSchema = z
	.object({
		data: z.array(JudgmentsApiSchemas.Admin_JudgmentSpec_JudgmentError),
		success: z.boolean().default(true)
	})
	.passthrough();

export const SubmissionRateEventDataSchema = z
	.object({
		data: z.array(SubmissionsApiSchemas.Admin_SubmissionSpec_SubmissionRatePoint),
		success: z.boolean().default(true)
	})
	.passthrough();

export interface JudgeQueueStorageInterface extends Storage<JudgeQueueStatus> {}
export const useJudgeQueueStorage = () =>
	container.get('JudgeQueueStorage') as JudgeQueueStorageInterface;

export interface JudgmentErrorsStorageInterface extends Storage<JudgmentError[]> {}
export const useJudgmentErrorsStorage = () =>
	container.get('JudgmentErrorsStorage') as JudgmentErrorsStorageInterface;

export interface SubmissionRateStorageInterface extends Storage<SubmissionRatePoint[]> {}
export const useSubmissionRateStorage = () =>
	container.get('SubmissionRateStorage') as SubmissionRateStorageInterface;
