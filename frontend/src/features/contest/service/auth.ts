import { apiClient } from '@/api';
import type { AuthServiceInterface } from '@/apps/contest/application/service/auth';
import { LoginFailedError, User } from '@/apps/contest/domain/user';

export class AuthService implements AuthServiceInterface {
	getProfile(): Promise<User> {
		return apiClient.getProfile().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to get profile');
			}
			return new User(data.name, data.username, data.roles);
		});
	}

	login(username: string, password: string): Promise<void> {
		return apiClient
			.login({ username, password })
			.then((response) => {
				if (!response.success) {
					throw new LoginFailedError(undefined);
				}

				return void 0;
			})
			.catch((error) => {
				// 這裡是唯一看得到 HTTP 狀態的地方。往上只丟一句「帳密錯誤」的話,
				// 被頻率限制擋下(429)與連不上伺服器都會被講成密碼打錯,選手於是
				// 一直重試 —— 而重試正好讓 429 更難恢復。
				if (error instanceof LoginFailedError) {
					throw error;
				}
				throw new LoginFailedError(
					(error as { response?: { status?: number } })?.response?.status
				);
			});
	}

	logout(): Promise<void> {
		return apiClient.logout(undefined).then((response) => {
			if (!response.success) {
				throw new Error('Failed to logout');
			}

			return void 0;
		});
	}
}
