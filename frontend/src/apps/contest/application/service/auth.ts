import type { User } from '@apps/contest/domain/user';

export interface AuthServiceInterface {
	getProfile(): Promise<User>;
	login(username: string, password: string): Promise<void>;
	logout(): Promise<void>;
}
