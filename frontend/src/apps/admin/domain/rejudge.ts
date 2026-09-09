import type { JudgmentsApiSchemas } from '@/api/gen/Judgments';
import type { Storage } from '@/apps/common/domain/storage';
import type { z } from 'zod';
import { container } from '@/apps/admin/container';

export type Rejudge = z.infer<typeof JudgmentsApiSchemas.Admin_JudgmentSpec_Rejudge>;
export type RejudgeDetail = z.infer<typeof JudgmentsApiSchemas.Admin_JudgmentSpec_RejudgeDetail>;

export type SubmissionComparisonDetail = z.infer<
	typeof JudgmentsApiSchemas.Admin_JudgmentSpec_SubmissionComparisonDetail
>;
export type JudgmentOutcomeDetail = z.infer<
	typeof JudgmentsApiSchemas.Admin_JudgmentSpec_JudgmentOutcomeDetail
>;
export type GradingItemOutcome = z.infer<
	typeof JudgmentsApiSchemas.Admin_JudgmentSpec_GradingItemOutcome
>;
export type TestcaseOutcome = z.infer<
	typeof JudgmentsApiSchemas.Admin_JudgmentSpec_TestcaseOutcome
>;

export interface RejudgesStorageInterface extends Storage<Rejudge[]> {}
export const useRejudgesStorage = () =>
	container.get('RejudgesStorage') as RejudgesStorageInterface;

export interface RejudgeStorageInterface extends Storage<RejudgeDetail> {}
export const useRejudgeStorage = () => container.get('RejudgeStorage') as RejudgeStorageInterface;

export interface RejudgeComparisonStorageInterface extends Storage<SubmissionComparisonDetail> {}
export const useRejudgeComparisonStorage = () =>
	container.get('RejudgeComparisonStorage') as RejudgeComparisonStorageInterface;
