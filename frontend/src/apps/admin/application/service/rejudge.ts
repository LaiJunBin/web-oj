import type {
	Rejudge,
	RejudgeDetail,
	SubmissionComparisonDetail
} from '@/apps/admin/domain/rejudge';

export interface RejudgeServiceInterface {
	fetchRejudges(): Promise<Rejudge[]>;
	fetchRejudgesByContestProject(contestProjectId: number): Promise<Rejudge[]>;
	fetchRejudgeDetail(id: string): Promise<RejudgeDetail>;
	fetchSubmissionComparison(
		rejudgeId: string,
		submissionId: string
	): Promise<SubmissionComparisonDetail>;
	cancelRejudge(id: string): Promise<void>;
	approveRejudge(id: string): Promise<void>;
	rejectRejudge(id: string): Promise<void>;
}
