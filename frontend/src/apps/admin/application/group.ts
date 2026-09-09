import { Application } from '@/apps/common/application/base';
import type { GroupServiceInterface } from '@/apps/admin/application/service/group';
import type {
	CreateGroupPayload,
	GroupsStorageInterface,
	GroupStorageInterface,
	UpdateGroupPayload
} from '@/apps/admin/domain/group';
import { container } from '@/apps/admin/container';

export class GroupApplication extends Application {
	constructor(
		private readonly groupService: GroupServiceInterface,

		private readonly groupsStorage: GroupsStorageInterface,
		private readonly groupStorage: GroupStorageInterface
	) {
		super();
	}

	fetchGroups() {
		return this.groupService
			.fetchGroups()
			.then((groups) => {
				this.groupsStorage.set(groups);
			})
			.catch((error) => {
				this.groupsStorage.setError(error);
			});
	}

	fetchGroup(id: number) {
		return this.groupService
			.fetchGroup(id)
			.then((group) => {
				this.groupStorage.set(group);
			})
			.catch((error) => {
				this.groupStorage.setError(error);
			});
	}

	createGroup(payload: CreateGroupPayload) {
		return this.groupService.createGroup(payload);
	}

	updateGroup(id: number, payload: UpdateGroupPayload) {
		return this.groupService.updateGroup(id, payload);
	}

	deleteGroup(id: number) {
		return this.groupService.deleteGroup(id);
	}

	addUsersToGroup(id: number, userIds: number[]) {
		return this.groupService.addUsersToGroup(id, userIds);
	}

	removeUsersFromGroup(id: number, userIds: number[]) {
		return this.groupService.removeUsersFromGroup(id, userIds);
	}
}

export const useGroupApp = () => container.get(GroupApplication);
