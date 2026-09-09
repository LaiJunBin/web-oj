import { apiClient } from '@/api';
import type { JudgeWorkerServiceInterface } from '@/apps/admin/application/service/judge-worker';
import type {
	CreateJudgeWorkerPayload,
	JudgeWorker,
	SystemDefaults
} from '@/apps/admin/domain/judge-worker';

export class JudgeWorkerService implements JudgeWorkerServiceInterface {
	fetchJudgeWorkers(): Promise<JudgeWorker[]> {
		return apiClient.listJudgeWorker().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch judge workers');
			}

			return data;
		});
	}

	fetchJudgePriorityStrategies(): Promise<string[]> {
		return apiClient.listJudgePriorityStrategies().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch judge priority strategies');
			}

			return data;
		});
	}

	fetchSystemDefaults(): Promise<SystemDefaults> {
		return apiClient.getSystemDefaults().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch system defaults');
			}

			return data;
		});
	}

	createJudgeWorker(payload: CreateJudgeWorkerPayload): Promise<void> {
		return apiClient
			.addJudgeWorker(payload)
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to create judge worker');
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

	deleteJudgeWorker(id: string): Promise<void> {
		return apiClient
			.deleteJudgeWorker(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete judge worker');
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

	startJudgeWorker(id: string): Promise<void> {
		return apiClient
			.startJudgeWorker(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to start judge worker');
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

	stopJudgeWorker(id: string): Promise<void> {
		return apiClient
			.stopJudgeWorker(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to stop judge worker');
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
