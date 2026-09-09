import { UserApiSchemas } from '@/api/gen/User';
import z from 'zod';
import type { Storage } from '@apps/common/domain/storage';
import { container } from '@/apps/contest/container';

export type Clarification = z.infer<typeof UserApiSchemas.User_UserSpec_Clarification>;
export type ClarificationDetail = z.infer<typeof UserApiSchemas.User_UserSpec_ClarificationDetail>;

export interface ClarificationsStorageInterface extends Storage<Clarification[]> {}
export const useClarificationsStorage = () =>
	container.get('ClarificationsStorage') as ClarificationsStorageInterface;

export interface ClarificationStorageInterface extends Storage<ClarificationDetail> {}
export const useClarificationStorage = () =>
	container.get('ClarificationStorage') as ClarificationStorageInterface;
