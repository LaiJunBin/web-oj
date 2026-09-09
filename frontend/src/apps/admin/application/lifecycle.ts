import { Application } from '@/apps/common/application/base';
import type { LifecycleServiceInterface } from '@/apps/admin/application/service/lifecycle';
import type { CreateLifecyclePayload, UpdateLifecyclePayload } from '@/apps/admin/domain/lifecycle';
import { container } from '@/apps/admin/container';

export class LifecycleApplication extends Application {
	constructor(private readonly lifecycleService: LifecycleServiceInterface) {
		super();
	}

	createLifecycle(projectId: number, payload: CreateLifecyclePayload) {
		return this.lifecycleService.createLifecycle(projectId, payload);
	}

	updateLifecycle(id: number, payload: UpdateLifecyclePayload) {
		return this.lifecycleService.updateLifecycle(id, payload);
	}

	deleteLifecycle(id: number) {
		return this.lifecycleService.deleteLifecycle(id);
	}
}

export const useLifecycleApp = () => container.get(LifecycleApplication);
