import type {
	CreateClarificationPayload,
	Clarification,
	ClarificationDetail,
	ReplyClarificationPayload
} from '@/apps/admin/domain/clarification';

export interface ClarificationServiceInterface {
	fetchClarifications(contestProjectId: number): Promise<Clarification[]>;
	fetchClarificationDetail(id: number): Promise<ClarificationDetail>;
	sendClarification(contestProjectId: number, payload: CreateClarificationPayload): Promise<number>;
	replyClarification(id: number, payload: ReplyClarificationPayload): Promise<number>;
	markClarificationAsAnswered(id: number): Promise<void>;
	markClarificationAsUnanswered(id: number): Promise<void>;
}
