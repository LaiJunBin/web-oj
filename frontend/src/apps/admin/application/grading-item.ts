import { Application } from '@/apps/common/application/base';
import type { GradingItemServiceInterface } from '@/apps/admin/application/service/grading-item';
import type {
	CreateGradingItemPayload,
	GradingItemsStorageInterface,
	GradingItemStorageInterface,
	UpdateGradingItemPayload
} from '@/apps/admin/domain/grading-item';
import { container } from '@/apps/admin/container';

export class GradingItemApplication extends Application {
	constructor(
		private readonly gradingItemService: GradingItemServiceInterface,

		private readonly gradingItemsStorage: GradingItemsStorageInterface,
		private readonly gradingItemStorage: GradingItemStorageInterface
	) {
		super();
	}

	fetchGradingItems(projectId: number) {
		return this.gradingItemService
			.fetchGradingItems(projectId)
			.then((gradingItems) => {
				this.gradingItemsStorage.set(gradingItems);
			})
			.catch((error) => {
				this.gradingItemsStorage.setError(error);
			});
	}

	fetchGradingItem(id: number) {
		return this.gradingItemService
			.fetchGradingItemDetail(id)
			.then((gradingItem) => {
				this.gradingItemStorage.set(gradingItem);
			})
			.catch((error) => {
				this.gradingItemStorage.setError(error);
			});
	}

	createGradingItem(projectId: number, payload: CreateGradingItemPayload) {
		return this.gradingItemService.createGradingItem(projectId, payload);
	}

	updateGradingItem(id: number, payload: UpdateGradingItemPayload) {
		return this.gradingItemService.updateGradingItem(id, payload);
	}

	deleteGradingItem(id: number) {
		return this.gradingItemService.deleteGradingItem(id);
	}
}

export const useGradingItemApp = () => container.get(GradingItemApplication);
