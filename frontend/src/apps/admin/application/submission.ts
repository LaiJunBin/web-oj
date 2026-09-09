import { Application } from '@/apps/common/application/base';
import type { SubmissionServiceInterface } from '@/apps/admin/application/service/submission';
import type {
	RejudgeSubmissionPayload,
	SubmissionsStorageInterface,
	SubmissionStorageInterface
} from '@/apps/admin/domain/submission';
import { container } from '@/apps/admin/container';

export class SubmissionApplication extends Application {
	constructor(
		private readonly submissionService: SubmissionServiceInterface,

		private readonly submissionsStorage: SubmissionsStorageInterface,
		private readonly submissionStorage: SubmissionStorageInterface
	) {
		super();
	}

	fetchSubmissions(contestProjectId: number) {
		return this.submissionService
			.fetchSubmissions(contestProjectId)
			.then((submissions) => {
				this.submissionsStorage.set(submissions);
			})
			.catch((error) => {
				this.submissionsStorage.setError(error);
			});
	}

	fetchSubmission(id: string) {
		return this.submissionService
			.fetchSubmissionDetail(id)
			.then((submission) => {
				this.submissionStorage.set(submission);
			})
			.catch((error) => {
				this.submissionStorage.setError(error);
			});
	}

	rejudgeSubmission(id: string, payload: RejudgeSubmissionPayload) {
		return this.submissionService.rejudgeSubmission(id, payload);
	}

	ignoreSubmission(id: string) {
		return this.submissionService.ignoreSubmission(id);
	}

	unignoreSubmission(id: string) {
		return this.submissionService.unignoreSubmission(id);
	}

	setJudgmentStaffNote(judgmentId: string, staffNote: string) {
		return this.submissionService.setJudgmentStaffNote(judgmentId, staffNote);
	}
}

export const useSubmissionApp = () => container.get(SubmissionApplication);
