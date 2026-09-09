import type { RejudgeSubmissionPayload, Submission, SubmissionDetail } from '@/apps/admin/domain/submission';

export interface SubmissionServiceInterface {
	fetchSubmissions(contestProjectId: number): Promise<Submission[]>;
	fetchSubmissionDetail(id: string): Promise<SubmissionDetail>;
	rejudgeSubmission(id: string, payload: RejudgeSubmissionPayload): Promise<void>;
	ignoreSubmission(id: string): Promise<void>;
	unignoreSubmission(id: string): Promise<void>;
	setJudgmentStaffNote(judgmentId: string, staffNote: string): Promise<void>;
}
