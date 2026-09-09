import type {
	User,
	UserDetail,
	UsersStorageInterface,
	UserStorageInterface
} from '@/apps/admin/domain/user';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class UsersStorage extends BrowserStorage<User[]> implements UsersStorageInterface {}

export class UserStorage extends BrowserStorage<UserDetail> implements UserStorageInterface {}
