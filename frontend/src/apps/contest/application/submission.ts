import { container } from '@/apps/contest/container';
import type { SubmissionStorageInterface } from '@apps/contest/domain/submission';
import { Application } from '@apps/common/application/base';
import type { SubmissionServiceInterface } from './service/submission';

export class SubmissionApplication extends Application {
	constructor(
		private readonly submissionService: SubmissionServiceInterface,

		private readonly submissionStorage: SubmissionStorageInterface
	) {
		super();
	}

	fetchSubmission(id: string) {
		return this.submissionService
			.fetchSubmission(id)
			.then((submission) => {
				this.submissionStorage.set(submission);
			})
			.catch((error) => {
				this.submissionStorage.setError(error);
			});
	}

	cancelSubmission(id: string) {
		return this.submissionService.cancelSubmission(id);
	}
}

export const useSubmissionApp = () => container.get(SubmissionApplication);
