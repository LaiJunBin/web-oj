import type { CreateLifecyclePayload, UpdateLifecyclePayload } from '@/apps/admin/domain/lifecycle';

export interface LifecycleServiceInterface {
	createLifecycle(projectId: number, payload: CreateLifecyclePayload): Promise<void>;
	updateLifecycle(id: number, payload: UpdateLifecyclePayload): Promise<void>;
	deleteLifecycle(id: number): Promise<void>;
}
