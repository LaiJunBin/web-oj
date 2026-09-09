import { z } from 'zod';
import { PublicContestApiSchemas } from '@/api/gen/PublicContest';
import { UserContestApiSchemas } from '@/api/gen/UserContest';
import type { Storage } from '@apps/common/domain/storage';
import { container } from '@/apps/contest/container';

export type PublicContest = z.infer<typeof PublicContestApiSchemas.Common_ContestSpec_Contest>;
export type PublicContestDetail = z.infer<
	typeof PublicContestApiSchemas.Common_ContestSpec_ContestDetail
>;

export type UserContest = z.infer<typeof UserContestApiSchemas.Common_ContestSpec_Contest>;
export type UserContestDetail = z.infer<
	typeof UserContestApiSchemas.Common_ContestSpec_ContestDetail
>;
export type Contest = PublicContest | UserContest;
export type ContestDetail = PublicContestDetail | UserContestDetail;
export type ScheduleProject = z.infer<
	typeof PublicContestApiSchemas.Common_ContestSpec_ContestProjectSchedule
>;
export type ContestSchedule = z.infer<
	typeof PublicContestApiSchemas.Common_ContestSpec_ContestSchedule
>;

export interface ContestsStorageInterface extends Storage<Contest[]> {}
export const useContestsStorage = () =>
	container.get('ContestsStorage') as ContestsStorageInterface;

export interface ContestStorageInterface extends Storage<ContestDetail> {}
export const useContestStorage = () => container.get('ContestStorage') as ContestStorageInterface;

export interface ContestScheduleStorageInterface extends Storage<ContestSchedule> {}
export const useContestScheduleStorage = () =>
	container.get('ContestScheduleStorage') as ContestScheduleStorageInterface;
