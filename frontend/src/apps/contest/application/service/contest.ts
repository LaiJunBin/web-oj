import type { Contest, ContestDetail, ContestSchedule } from '@apps/contest/domain/contest';

export interface ContestServiceInterface {
	fetchContests(): Promise<Contest[]>;
	fetchContestDetail(contestId: number): Promise<ContestDetail>;
	fetchContestSchedule(contestId: number): Promise<ContestSchedule>;
}
