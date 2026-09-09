import type { Storage } from '@/apps/common/domain/storage';
import { container } from '@/apps/admin/container';

export type Group = {
	id: number;
	name: string;
	description: string;
};

export type GroupDetail = Group & {
	userIds: number[];
};

export type CreateGroupPayload = {
	name: string;
	description?: string;
};

export type UpdateGroupPayload = {
	name?: string;
	description?: string;
};

export interface GroupsStorageInterface extends Storage<Group[]> {}
export const useGroupsStorage = () => container.get('GroupsStorage') as GroupsStorageInterface;

export interface GroupStorageInterface extends Storage<GroupDetail> {}
export const useGroupStorage = () => container.get('GroupStorage') as GroupStorageInterface;
