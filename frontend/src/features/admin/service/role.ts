import { apiClient } from '@/api';
import type { RoleServiceInterface } from '@/apps/admin/application/service/role';
import type { Role } from '@/apps/admin/domain/role';

export class RoleService implements RoleServiceInterface {
	fetchRoles(): Promise<Role[]> {
		return apiClient.listRole().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch roles');
			}

			return data;
		});
	}
}
