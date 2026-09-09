import type { Storage } from '@/apps/common/domain/storage';
import { container } from '@/apps/admin/container';
import type { z } from 'zod';
import type { ContestProjectApiSchemas } from '@/api/gen/ContestProject';
import type { SubmissionsApiSchemas } from '@/api/gen/Submissions';

type ContestProjectExtraFields = {
	submissions: z.infer<typeof SubmissionsApiSchemas.Admin_SubmissionSpec_Submission>[];
};
export type ContestProject = z.infer<
	typeof ContestProjectApiSchemas.Admin_ContestProjectSpec_ContestProject
> &
	ContestProjectExtraFields;

export type UserScore = z.infer<typeof ContestProjectApiSchemas.Admin_ContestProjectSpec_UserScore>;
export type RejudgeContestProjectPayload = z.infer<
	typeof ContestProjectApiSchemas.Admin_ContestProjectSpec_RejudgeContestProjectPayload
>;
export type JudgmentAnalytics = z.infer<
	typeof ContestProjectApiSchemas.Admin_ContestProjectSpec_JudgmentAnalytics
>;
export type JudgmentStatusCount = z.infer<
	typeof ContestProjectApiSchemas.Admin_ContestProjectSpec_JudgmentStatusCount
>;

export interface ContestProjectStorageInterface extends Storage<ContestProject> {}
export const useContestProjectStorage = () =>
	container.get('ContestProjectStorage') as ContestProjectStorageInterface;

export interface UserScoresStorageInterface extends Storage<UserScore[]> {}
export const useUserScoresStorage = () =>
	container.get('UserScoresStorage') as UserScoresStorageInterface;

export interface JudgmentAnalyticsStorageInterface extends Storage<JudgmentAnalytics> {}
export const useJudgmentAnalyticsStorage = () =>
	container.get('JudgmentAnalyticsStorage') as JudgmentAnalyticsStorageInterface;
