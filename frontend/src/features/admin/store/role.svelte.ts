import type { Role, RolesStorageInterface } from '@/apps/admin/domain/role';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class RolesStorage extends BrowserStorage<Role[]> implements RolesStorageInterface {}
