import { apiClient } from '@/api';
import type { SubmissionServiceInterface } from '@/apps/contest/application/service/submission';
import type { SubmissionDetail } from '@/apps/contest/domain/submission';

export class SubmissionService implements SubmissionServiceInterface {
	fetchSubmission(id: string): Promise<SubmissionDetail> {
		return apiClient
			.getUserSubmission({
				params: {
					submissionId: id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch submission');
				}

				return data;
			});
	}

	cancelSubmission(id: string): Promise<void> {
		return apiClient
			.cancelUserSubmission(undefined, {
				params: {
					submissionId: id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to cancel submission');
				}
			});
	}
}
