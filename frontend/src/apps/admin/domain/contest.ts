import { ContestsApiSchemas } from '@/api/gen/Contests';
import type { Storage } from '@/apps/common/domain/storage';
import { container } from '@/apps/admin/container';
import type { z } from 'zod';

export type Contest = z.infer<typeof ContestsApiSchemas.Admin_ContestSpec_Contest>;
export type ContestDetail = z.infer<typeof ContestsApiSchemas.Admin_ContestSpec_ContestDetail>;

export type CreateContestPayload = z.infer<
	typeof ContestsApiSchemas.Admin_ContestSpec_CreateContestRequest
>;
export type UpdateContestPayload = z.infer<
	typeof ContestsApiSchemas.Admin_ContestSpec_UpdateContestRequest
>;

export interface ContestsStorageInterface extends Storage<Contest[]> {}
export const useContestsStorage = () =>
	container.get('ContestsStorage') as ContestsStorageInterface;

export interface ContestStorageInterface extends Storage<ContestDetail> {}
export const useContestStorage = () => container.get('ContestStorage') as ContestStorageInterface;

export type ContestParticipants = {
	userIds: number[];
	groupIds: number[];
};

export interface ContestParticipantsStorageInterface extends Storage<ContestParticipants> {}
export const useContestParticipantsStorage = () =>
	container.get('ContestParticipantsStorage') as ContestParticipantsStorageInterface;
