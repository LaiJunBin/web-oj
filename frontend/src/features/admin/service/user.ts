import { apiClient } from '@/api';
import type { UserServiceInterface } from '@/apps/admin/application/service/user';
import type {
	CreateUserPayload,
	UpdateUserPayload,
	User,
	UserDetail
} from '@/apps/admin/domain/user';

export class UserService implements UserServiceInterface {
	fetchUsers(): Promise<User[]> {
		return apiClient.listUser().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch users');
			}

			return data;
		});
	}
	fetchUser(userId: number): Promise<UserDetail> {
		return Promise.all([
			apiClient
				.getUser({
					params: {
						id: userId
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch user');
					}

					return data;
				}),
			apiClient
				.getSubmissionsByUser({
					params: {
						id: userId
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch user submissions');
					}

					return data;
				})
		]).then(([user, submissions]) => {
			return {
				...user,
				submissions
			};
		});
	}

	createUser(payload: CreateUserPayload): Promise<number> {
		return apiClient.createUser(payload).then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to create user');
			}

			return data;
		});
	}

	updateUser(id: number, payload: UpdateUserPayload): Promise<void> {
		return apiClient
			.updateUser(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to update user');
				}

				return;
			});
	}

	deleteUser(id: number): Promise<void> {
		return apiClient
			.deleteUser(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete user');
				}
			});
	}

	importUsers(file: File): Promise<number> {
		return apiClient
			.importUsers({
				users: file
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to import users');
				}

				return data;
			})
			.catch((error) => {
				if (error.response) {
					const { data } = error.response;
					const { details } = data;
					if (details) {
						throw details.join('\n');
					}
					throw data.message;
				}

				throw error;
			});
	}
}
