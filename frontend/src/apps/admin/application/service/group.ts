import type {
	CreateGroupPayload,
	Group,
	GroupDetail,
	UpdateGroupPayload
} from '@/apps/admin/domain/group';

export interface GroupServiceInterface {
	fetchGroups(): Promise<Group[]>;
	fetchGroup(id: number): Promise<GroupDetail>;
	createGroup(payload: CreateGroupPayload): Promise<number>;
	updateGroup(id: number, payload: UpdateGroupPayload): Promise<void>;
	deleteGroup(id: number): Promise<void>;
	addUsersToGroup(id: number, userIds: number[]): Promise<void>;
	removeUsersFromGroup(id: number, userIds: number[]): Promise<void>;
}
