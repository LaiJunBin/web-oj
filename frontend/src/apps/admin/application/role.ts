import { Application } from '@/apps/common/application/base';
import type { RoleServiceInterface } from '@/apps/admin/application/service/role';
import type { RolesStorageInterface } from '@/apps/admin/domain/role';
import { container } from '@/apps/admin/container';

export class RoleApplication extends Application {
	constructor(
		private readonly roleService: RoleServiceInterface,

		private readonly rolesStorage: RolesStorageInterface
	) {
		super();
	}

	fetchRoles() {
		return this.roleService
			.fetchRoles()
			.then((roles) => {
				this.rolesStorage.set(roles);
			})
			.catch((error) => {
				this.rolesStorage.setError(error);
			});
	}
}

export const useRoleApp = () => container.get(RoleApplication);
