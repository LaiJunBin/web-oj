import type {
	CreateUserPayload,
	UpdateUserPayload,
	User,
	UserDetail
} from '@/apps/admin/domain/user';

export interface UserServiceInterface {
	fetchUsers(): Promise<User[]>;
	fetchUser(id: number): Promise<UserDetail>;
	createUser(payload: CreateUserPayload): Promise<number>;
	updateUser(id: number, payload: UpdateUserPayload): Promise<void>;
	deleteUser(id: number): Promise<void>;
	importUsers(file: File): Promise<number>;
}
