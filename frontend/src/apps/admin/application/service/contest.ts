import type {
	Contest,
	ContestDetail,
	ContestParticipants,
	CreateContestPayload,
	UpdateContestPayload
} from '@apps/admin/domain/contest';

export interface ContestServiceInterface {
	fetchContests(): Promise<Contest[]>;
	fetchContestDetail(contestId: number): Promise<ContestDetail>;
	createContest(payload: CreateContestPayload): Promise<number>;
	updateContest(id: number, payload: UpdateContestPayload): Promise<void>;
	deleteContest(id: number): Promise<void>;
	fetchContestParticipants(id: number): Promise<ContestParticipants>;
	addUsersToContest(id: number, userIds: number[]): Promise<void>;
	removeUsersFromContest(id: number, userIds: number[]): Promise<void>;
	addGroupsToContest(id: number, groupIds: number[]): Promise<void>;
	removeGroupsFromContest(id: number, groupIds: number[]): Promise<void>;
	exportContest(id: number): Promise<void>;
	importContest(file: File): Promise<{ contestId: number; name: string; createdProjects: number; unresolvedProjects: string[] }>;
}
