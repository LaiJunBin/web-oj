import type {
	Lifecycle,
	LifecycleType,
	LifecycleStorageInterface,
	LifecycleTypesStorageInterface
} from '@/apps/admin/domain/lifecycle';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class LifecycleStorage
	extends BrowserStorage<Lifecycle>
	implements LifecycleStorageInterface {}

export class LifecycleTypesStorage
	extends BrowserStorage<LifecycleType[]>
	implements LifecycleTypesStorageInterface {}
