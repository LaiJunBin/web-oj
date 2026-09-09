import type { SubmissionsApiSchemas } from '@/api/gen/Submissions';
import type { Storage } from '@/apps/common/domain/storage';
import type { z } from 'zod';
import { container } from '@/apps/admin/container';

export type Submission = z.infer<typeof SubmissionsApiSchemas.Admin_SubmissionSpec_Submission>;
export type SubmissionDetail = z.infer<
	typeof SubmissionsApiSchemas.Admin_SubmissionSpec_SubmissionDetail
>;
export type Judgment = z.infer<typeof SubmissionsApiSchemas.Admin_JudgmentSpec_Judgment>;

export type RejudgeSubmissionPayload = {
	reason: string;
};

export interface SubmissionsStorageInterface extends Storage<Submission[]> {}
export const useSubmissionsStorage = () =>
	container.get('SubmissionsStorage') as SubmissionsStorageInterface;

export interface SubmissionStorageInterface extends Storage<SubmissionDetail> {}
export const useSubmissionStorage = () =>
	container.get('SubmissionStorage') as SubmissionStorageInterface;
