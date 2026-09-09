import type { ClarificationsApiSchemas } from '@/api/gen/Clarifications';
import type { Storage } from '@/apps/common/domain/storage';
import type { z } from 'zod';
import { container } from '@/apps/admin/container';

export type Clarification = z.infer<
	typeof ClarificationsApiSchemas.Admin_ProjectSpec_ClarificationSpec_ClarificationInfo
>;
export type ClarificationDetail = z.infer<
	typeof ClarificationsApiSchemas.Admin_ProjectSpec_ClarificationSpec_Clarification
>;

export type CreateClarificationPayload = z.infer<
	typeof ClarificationsApiSchemas.Admin_ProjectSpec_ClarificationSpec_CreateClarificationRequest
>;

export type ReplyClarificationPayload = z.infer<
	typeof ClarificationsApiSchemas.Admin_ProjectSpec_ClarificationSpec_ReplyClarificationRequest
>;

export interface ClarificationsStorageInterface extends Storage<Clarification[]> {}
export const useClarificationsStorage = () =>
	container.get('ClarificationsStorage') as ClarificationsStorageInterface;

export interface ClarificationStorageInterface extends Storage<ClarificationDetail> {}
export const useClarificationStorage = () =>
	container.get('ClarificationStorage') as ClarificationStorageInterface;
