import { UserApplication } from '@/apps/contest/application/user';
import { LoginFailedError, User, type UserStorageInterface } from '@/apps/contest/domain/user';
import type { AuthServiceInterface } from '@/apps/contest/application/service/auth';
import { describe, expect, test } from 'vitest';

class FakeUserStorage implements UserStorageInterface {
	private value: User | null = null;
	error: Error | null = null;
	set(data: User) {
		this.value = data;
	}
	get() {
		return this.value;
	}
	clear() {
		this.value = null;
	}
	setError(error: Error) {
		this.error = error;
	}
}

function appRejectingWith(error: unknown) {
	const service: AuthServiceInterface = {
		login: () => Promise.reject(error),
		logout: () => Promise.resolve(),
		getProfile: () => Promise.resolve(new User('n', 'u', []))
	};
	return new UserApplication(new FakeUserStorage(), service);
}

// 這條守的是一個曾經真的壞掉的性質:login 一律 catch 後改丟
// Error('Invalid credentials'),於是被登入頻率限制擋下(429)與連不到伺服器
// 在畫面上都變成「請檢查帳號密碼」。選手照著提示重試,而重試正是最糟的動作。
// 型別檢查抓不到這種「錯誤被換掉」,只能用測試釘住。
describe('UserApplication.login', () => {
	test('把服務層的失敗原樣往上丟,不壓成單一種錯誤', async () => {
		const original = new LoginFailedError(429);

		await expect(appRejectingWith(original).login('demo01', '1234')).rejects.toBe(original);
	});

	test('狀態碼在傳遞過程中不會遺失', async () => {
		for (const status of [401, 429, undefined]) {
			const caught = await appRejectingWith(new LoginFailedError(status))
				.login('demo01', '1234')
				.then(() => null)
				.catch((error: unknown) => error);

			expect(caught).toBeInstanceOf(LoginFailedError);
			expect((caught as LoginFailedError).status).toBe(status);
		}
	});
});
