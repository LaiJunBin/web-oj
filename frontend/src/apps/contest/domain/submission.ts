import { UserApi, UserApiSchemas } from '@/api/gen/User';
import z, { ZodObject } from 'zod';
import type { Storage } from '@apps/common/domain/storage';
import { container } from '@/apps/contest/container';

export type Submission = z.infer<typeof UserApiSchemas.User_UserSpec_Submission>;
export type SubmissionStatus = z.infer<typeof UserApiSchemas.User_UserSpec_SubmissionStatus>;
export type SubmissionDetail = z.infer<typeof UserApiSchemas.User_UserSpec_SubmissionDetail>;
export type SubmissionGradingItem = z.infer<
	typeof UserApiSchemas.User_UserSpec_SubmissionGradingItem
>;
export type SubmissionTestcase = z.infer<typeof UserApiSchemas.User_UserSpec_SubmissionTestcase>;

export type SubmissionUpdateEventData = {
	id: string;
	score: number;
	status: SubmissionStatus;
	gradingItems: SubmissionGradingItem[];
};

// The submission-update SSE event omits `label` on grading items — the backend
// stream only sends id/score/status/testcases. Accept it as optional here;
// otherwise every event fails validation and the live update is silently dropped.
const SubmissionUpdateGradingItemSchema = (
	UserApiSchemas.User_UserSpec_SubmissionGradingItem as unknown as ZodObject<
		any,
		any,
		any,
		any,
		any
	>
).partial({ label: true });

export const SubmissionUpdateEventDataSchema = z
	.object({
		score: z.number(),
		status: UserApiSchemas.User_UserSpec_SubmissionStatus,
		gradingItems: z.array(SubmissionUpdateGradingItemSchema)
	})
	.passthrough() as unknown as ZodObject<any, 'strip', any, SubmissionUpdateEventData, any>;

export interface SubmissionsStorageInterface extends Storage<Submission[]> {}
export const useSubmissionsStorage = () =>
	container.get('SubmissionsStorage') as SubmissionsStorageInterface;

export interface SubmissionStorageInterface extends Storage<SubmissionDetail> {}
export const useSubmissionStorage = () =>
	container.get('SubmissionStorage') as SubmissionStorageInterface;
