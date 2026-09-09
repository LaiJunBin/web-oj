import { apiClient } from '@/api';
import type { SystemStatusServiceInterface } from '@/apps/admin/application/service/system-status';
import type {
	JudgeQueueStatus,
	JudgmentError,
	SubmissionRatePoint
} from '@/apps/admin/domain/system-status';

export class SystemStatusService implements SystemStatusServiceInterface {
	fetchJudgeQueue(): Promise<JudgeQueueStatus> {
		return apiClient.getJudgeQueue().then((response) => {
			if (!response.success) {
				throw new Error('Failed to fetch judge queue');
			}
			return response.data;
		});
	}

	fetchJudgmentErrors(limit: number): Promise<JudgmentError[]> {
		return apiClient.listJudgmentErrors({ queries: { limit } }).then((response) => {
			if (!response.success) {
				throw new Error('Failed to fetch judgment errors');
			}
			return response.data;
		});
	}

	fetchSubmissionRate(minutes: number, contestProjectId?: number): Promise<SubmissionRatePoint[]> {
		return apiClient
			.getSubmissionRate({ queries: { minutes, contestProjectId } })
			.then((response) => {
				if (!response.success) {
					throw new Error('Failed to fetch submission rate');
				}
				return response.data;
			});
	}
}
