import { apiClient } from '@/api';
import type { RejudgeServiceInterface } from '@/apps/admin/application/service/rejudge';
import type {
	Rejudge,
	RejudgeDetail,
	SubmissionComparisonDetail
} from '@/apps/admin/domain/rejudge';

export class RejudgeService implements RejudgeServiceInterface {
	fetchRejudges(): Promise<Rejudge[]> {
		return apiClient.listRejudges().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch submissions');
			}

			return data;
		});
	}
	fetchRejudgesByContestProject(contestProjectId: number): Promise<Rejudge[]> {
		return apiClient
			.listRejudgesByContestProject({
				params: {
					id: contestProjectId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch rejudges');
				}

				return data;
			});
	}
	fetchRejudgeDetail(id: string): Promise<RejudgeDetail> {
		return apiClient
			.getRejudge({
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch rejudge');
				}

				return data;
			});
	}
	fetchSubmissionComparison(
		rejudgeId: string,
		submissionId: string
	): Promise<SubmissionComparisonDetail> {
		return apiClient
			.getRejudgeSubmissionComparison({
				params: {
					id: rejudgeId,
					submissionId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch submission comparison');
				}

				return data;
			});
	}
	cancelRejudge(id: string): Promise<void> {
		return apiClient
			.cancelRejudge(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to cancel rejudge');
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
	approveRejudge(id: string): Promise<void> {
		return apiClient
			.approveRejudge(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to approve rejudge');
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
	rejectRejudge(id: string): Promise<void> {
		return apiClient
			.rejectRejudge(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to reject rejudge');
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
