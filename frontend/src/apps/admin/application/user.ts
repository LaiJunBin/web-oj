import { Application } from '@/apps/common/application/base';
import type { UserServiceInterface } from '@/apps/admin/application/service/user';
import type {
	CreateUserPayload,
	UpdateUserPayload,
	UsersStorageInterface,
	UserStorageInterface
} from '@/apps/admin/domain/user';
import { container } from '@/apps/admin/container';

export class UserApplication extends Application {
	constructor(
		private readonly userService: UserServiceInterface,

		private readonly usersStorage: UsersStorageInterface,
		private readonly userStorage: UserStorageInterface
	) {
		super();
	}

	fetchUsers() {
		return this.userService
			.fetchUsers()
			.then((users) => {
				this.usersStorage.set(users);
			})
			.catch((error) => {
				this.usersStorage.setError(error);
			});
	}

	fetchUser(id: number) {
		return this.userService
			.fetchUser(id)
			.then((user) => {
				this.userStorage.set(user);
			})
			.catch((error) => {
				this.userStorage.setError(error);
			});
	}

	createUser(payload: CreateUserPayload) {
		return this.userService.createUser(payload);
	}

	updateUser(id: number, payload: UpdateUserPayload) {
		return this.userService.updateUser(id, payload);
	}

	deleteUser(id: number) {
		return this.userService.deleteUser(id);
	}

	importUsers(file: File) {
		return this.userService.importUsers(file);
	}
}

export const useUserApp = () => container.get(UserApplication);
