import { apiClient } from '@/api';
import type { SubmissionServiceInterface } from '@/apps/admin/application/service/submission';
import type {
	RejudgeSubmissionPayload,
	Submission,
	SubmissionDetail
} from '@/apps/admin/domain/submission';

export class SubmissionService implements SubmissionServiceInterface {
	fetchSubmissions(contestProjectId: number): Promise<Submission[]> {
		return apiClient
			.getSubmissionsByContestProject({
				params: {
					id: contestProjectId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch submissions');
				}

				return data;
			});
	}
	fetchSubmissionDetail(id: string): Promise<SubmissionDetail> {
		return apiClient
			.getSubmission({
				params: {
					id
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
	rejudgeSubmission(id: string, payload: RejudgeSubmissionPayload): Promise<void> {
		return apiClient
			.rejudgeSubmission(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to rejudge submission');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	ignoreSubmission(id: string): Promise<void> {
		return apiClient
			.ignoreSubmission(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to ignore submission');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	unignoreSubmission(id: string): Promise<void> {
		return apiClient
			.unignoreSubmission(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to unignore submission');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	setJudgmentStaffNote(judgmentId: string, staffNote: string): Promise<void> {
		return apiClient
			.setJudgmentStaffNote({ staffNote }, { params: { id: judgmentId } })
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to set staff note');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
}
