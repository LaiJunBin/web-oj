import type {
	ContestProject,
	JudgmentAnalytics,
	RejudgeContestProjectPayload,
	UserScore
} from '@apps/admin/domain/contest-project';

export interface ContestProjectServiceInterface {
	fetchContestProject(id: number): Promise<ContestProject>;
	fetchUserScores(contestProjectID: number): Promise<UserScore[]>;
	fetchJudgmentAnalytics(contestProjectID: number): Promise<JudgmentAnalytics>;
	exportContestProjectScoreDetail(contestProjectID: number, contestProjectName: string): Promise<void>;
	rejudgeContestProject(id: number, payload: RejudgeContestProjectPayload): Promise<void>;
}
