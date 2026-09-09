import type { Role } from '@/apps/admin/domain/role';

export interface RoleServiceInterface {
	fetchRoles(): Promise<Role[]>;
}
