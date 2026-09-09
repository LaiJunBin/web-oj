import { apiClient } from '@/api';
import type { LifecycleServiceInterface } from '@/apps/admin/application/service/lifecycle';
import type { CreateLifecyclePayload, UpdateLifecyclePayload } from '@/apps/admin/domain/lifecycle';

export class LifecycleService implements LifecycleServiceInterface {
	createLifecycle(projectId: number, payload: CreateLifecyclePayload): Promise<void> {
		return apiClient
			.createLifecycle(payload, {
				params: {
					id: projectId
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to create lifecycle');
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
	updateLifecycle(id: number, payload: UpdateLifecyclePayload): Promise<void> {
		return apiClient
			.updateLifecycle(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to update lifecycle');
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
	deleteLifecycle(id: number): Promise<void> {
		return apiClient
			.deleteLifecycle(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete lifecycle');
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
