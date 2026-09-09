import { Application } from '@/apps/common/application/base';
import type { UserStorageInterface } from '@apps/contest/domain/user';
import type { AuthServiceInterface } from './service/auth';
import { container } from '@/apps/contest/container';

export class UserProfile {
	constructor(private userStorage: UserStorageInterface) {}

	get username(): string {
		return this.userStorage.get()?.username ?? '';
	}

	get name(): string {
		return this.userStorage.get()?.name ?? '';
	}

	get isLoggedIn(): boolean {
		return this.userStorage.get() !== null;
	}

	get isContestant(): boolean {
		return this.isLoggedIn && this.userStorage.get()!.isContestant;
	}

	get isAdmin(): boolean {
		return this.isLoggedIn && this.userStorage.get()!.isAdmin;
	}
}

export class UserApplication extends Application {
	constructor(
		private readonly userStorage: UserStorageInterface,
		private readonly authService: AuthServiceInterface
	) {
		super();
	}

	async fetchProfile() {
		return this.authService
			.getProfile()
			.then((user) => {
				this.userStorage.set(user);
			})
			.catch((error) => {
				this.userStorage.setError(error);
			});
	}

	async login(username: string, password: string) {
		// 原樣往上丟。這裡曾經一律換成 Error('Invalid credentials'),於是「被登入
		// 頻率限制擋下」和「連不到伺服器」在畫面上都變成「請檢查帳號密碼」——選手
		// 照著提示重試,反而讓情況更糟。fetchProfile 自己吞錯誤,不會走到這裡。
		await this.authService.login(username, password);
		await this.fetchProfile();
	}

	async logout() {
		try {
			await this.authService.logout();
		} catch (error) {
			throw new Error('Failed to logout');
		} finally {
			await this.userStorage.clear();
		}
	}
}

export const useUserProfile = () => container.get(UserProfile);
export const useUserApp = () => container.get(UserApplication);
