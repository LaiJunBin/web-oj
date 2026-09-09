import type { User, UserStorageInterface } from '@/apps/contest/domain/user';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class UserStorage extends BrowserStorage<User> implements UserStorageInterface {}
