import type { Clarification, ClarificationDetail } from '@apps/contest/domain/clarification';

export interface ClarificationServiceInterface {
	fetchClarifications(projectId: number): Promise<Clarification[]>;
	fetchClarification(id: number): Promise<ClarificationDetail>;
	sendClarification(projectId: number, subject: string, content: string): Promise<number>;
	replyClarification(clarificationId: number, content: string): Promise<number>;
}
