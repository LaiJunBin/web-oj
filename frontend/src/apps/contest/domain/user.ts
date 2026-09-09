import { container } from '@/apps/contest/container';
import type { Storage } from '@apps/common/domain/storage';

type Roles = string[];

export class User {
	name: string;
	username: string;
	roles: Roles;
	constructor(name: string, username: string, roles: Roles) {
		this.name = name;
		this.username = username;
		this.roles = roles;
	}

	get isContestant(): boolean {
		return this.roles.includes('contestant');
	}

	get isAdmin(): boolean {
		return this.roles.includes('admin');
	}
}

/**
 * 登入失敗的原因決定選手該做什麼:401 是真的打錯帳密,429 是被登入頻率限制擋下
 * (「再試一次」才是對的動作,不是改密碼),連不上則根本不在他身上。HTTP 狀態碼
 * 只有 service 層看得到,所以把它帶進領域錯誤,而不是在應用層壓成一句「登入失敗」。
 */
export class LoginFailedError extends Error {
	constructor(readonly status: number | undefined) {
		super(`login failed${status === undefined ? '' : ` with status ${status}`}`);
		this.name = 'LoginFailedError';
	}
}

export interface UserStorageInterface extends Storage<User> {}
export const useUserStorage = () => container.get('UserStorage') as UserStorageInterface;
