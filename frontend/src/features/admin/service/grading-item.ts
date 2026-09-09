import { apiClient } from '@/api';
import type { GradingItemServiceInterface } from '@/apps/admin/application/service/grading-item';
import type {
	CreateGradingItemPayload,
	GradingItem,
	GradingItemDetail,
	UpdateGradingItemPayload
} from '@/apps/admin/domain/grading-item';

export class GradingItemService implements GradingItemServiceInterface {
	fetchGradingItems(projectId: number): Promise<GradingItem[]> {
		return apiClient
			.listGradingItem({
				params: {
					id: projectId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch grading items');
				}

				return data;
			});
	}
	fetchGradingItemDetail(id: number): Promise<GradingItemDetail> {
		return apiClient
			.getGradingItem({
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch grading item');
				}

				return data;
			});
	}
	createGradingItem(projectId: number, payload: CreateGradingItemPayload): Promise<number> {
		return apiClient
			.createGradingItem(payload, {
				params: {
					id: projectId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to create grading item');
				}
				return data;
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	updateGradingItem(id: number, payload: UpdateGradingItemPayload): Promise<void> {
		return apiClient
			.updateGradingItem(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to update grading item');
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
	deleteGradingItem(id: number): Promise<void> {
		return apiClient
			.deleteGradingItem(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete grading item');
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
