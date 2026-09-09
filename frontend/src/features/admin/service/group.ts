import { apiClient } from '@/api';
import type { GroupServiceInterface } from '@/apps/admin/application/service/group';
import type {
	CreateGroupPayload,
	Group,
	GroupDetail,
	UpdateGroupPayload
} from '@/apps/admin/domain/group';

export class GroupService implements GroupServiceInterface {
	fetchGroups(): Promise<Group[]> {
		return apiClient.listGroup().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch groups');
			}

			return data;
		});
	}

	fetchGroup(id: number): Promise<GroupDetail> {
		return apiClient
			.getGroup({
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch group');
				}

				return data;
			});
	}

	createGroup(payload: CreateGroupPayload): Promise<number> {
		return apiClient.createGroup(payload).then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to create group');
			}

			return data;
		});
	}

	updateGroup(id: number, payload: UpdateGroupPayload): Promise<void> {
		return apiClient
			.updateGroup(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to update group');
				}
			});
	}

	deleteGroup(id: number): Promise<void> {
		return apiClient
			.deleteGroup(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to delete group');
				}
			});
	}

	addUsersToGroup(id: number, userIds: number[]): Promise<void> {
		return apiClient
			.addUsersToGroup(
				{ userIds },
				{
					params: {
						id
					}
				}
			)
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to add users to group');
				}
			});
	}

	removeUsersFromGroup(id: number, userIds: number[]): Promise<void> {
		return apiClient
			.removeUsersFromGroup(
				{ userIds },
				{
					params: {
						id
					}
				}
			)
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to remove users from group');
				}
			});
	}
}
