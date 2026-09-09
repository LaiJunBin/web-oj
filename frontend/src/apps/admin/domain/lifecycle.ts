import { LifecyclesApiSchemas } from '@/api/gen/Lifecycles';
import type { Storage } from '@/apps/common/domain/storage';
import type { z } from 'zod';
import { container } from '@/apps/admin/container';

export type Lifecycle = z.infer<
	typeof LifecyclesApiSchemas.Admin_ProjectSpec_LifecycleSpec_LifecycleDetail
>;

export const LifecycleTypeEnum =
	LifecyclesApiSchemas.Admin_ProjectSpec_LifecycleSpec_LifecycleType.Enum;
export type LifecycleType = z.infer<
	typeof LifecyclesApiSchemas.Admin_ProjectSpec_LifecycleSpec_LifecycleType
>;

export const LifecycleRunsOnEnum = LifecyclesApiSchemas.Shared_RunsOn.Enum;
export type LifecycleRunsOn = z.infer<typeof LifecyclesApiSchemas.Shared_RunsOn>;

export type CreateLifecyclePayload = z.infer<
	typeof LifecyclesApiSchemas.Admin_ProjectSpec_LifecycleSpec_CreateLifecycleRequest
>;

export type UpdateLifecyclePayload = z.infer<
	typeof LifecyclesApiSchemas.Admin_ProjectSpec_LifecycleSpec_UpdateLifecycleRequest
>;

export interface LifecycleTypesStorageInterface extends Storage<LifecycleType[]> {}
export const useLifecycleTypesStorage = () =>
	container.get('LifecycleTypesStorage') as LifecycleTypesStorageInterface;

export interface LifecycleStorageInterface extends Storage<Lifecycle> {}
export const useLifecycleStorage = () =>
	container.get('LifecycleStorage') as LifecycleStorageInterface;
