import type { Submission, SubmissionDetail } from '@apps/contest/domain/submission';

export interface SubmissionServiceInterface {
	fetchSubmission(id: string): Promise<SubmissionDetail>;
	cancelSubmission(id: string): Promise<void>;
}
