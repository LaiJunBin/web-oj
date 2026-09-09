import type { UsersApiSchemas } from '@/api/gen/Users';
import type { Storage } from '@/apps/common/domain/storage';
import type { z } from 'zod';
import { container } from '@/apps/admin/container';
import type { SubmissionsApiSchemas } from '@/api/gen/Submissions';

export type User = z.infer<typeof UsersApiSchemas.Admin_UserSpec_User>;

type UserExtraFields = {
	submissions: z.infer<typeof SubmissionsApiSchemas.Admin_SubmissionSpec_Submission>[];
};

export type UserDetail = User & UserExtraFields;

export type CreateUserPayload = z.infer<typeof UsersApiSchemas.Admin_UserSpec_CreateUserRequest>;

export type UpdateUserPayload = z.infer<typeof UsersApiSchemas.Admin_UserSpec_UpdateUserRequest>;

export interface UsersStorageInterface extends Storage<User[]> {}
export const useUsersStorage = () => container.get('UsersStorage') as UsersStorageInterface;

export interface UserStorageInterface extends Storage<UserDetail> {}
export const useUserStorage = () => container.get('UserStorage') as UserStorageInterface;
