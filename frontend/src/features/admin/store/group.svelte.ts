import type {
	Group,
	GroupDetail,
	GroupsStorageInterface,
	GroupStorageInterface
} from '@/apps/admin/domain/group';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class GroupsStorage extends BrowserStorage<Group[]> implements GroupsStorageInterface {}

export class GroupStorage extends BrowserStorage<GroupDetail> implements GroupStorageInterface {}
