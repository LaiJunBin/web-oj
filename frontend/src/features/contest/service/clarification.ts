import { apiClient } from '@/api';
import type { ClarificationServiceInterface } from '@/apps/contest/application/service/clarification';
import type { Clarification, ClarificationDetail } from '@/apps/contest/domain/clarification';

export class ClarificationService implements ClarificationServiceInterface {
	fetchClarifications(projectId: number): Promise<Clarification[]> {
		return apiClient
			.listUserProjectClarifications({
				params: {
					id: projectId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch clarifications');
				}

				return data;
			});
	}

	fetchClarification(id: number): Promise<ClarificationDetail> {
		return apiClient
			.getUserClarification({
				params: {
					clarificationId: id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch clarification');
				}

				return data;
			});
	}

	sendClarification(projectId: number, subject: string, content: string): Promise<number> {
		return apiClient
			.sendUserProjectClarificationToAdmin(
				{
					subject,
					content
				},
				{
					params: {
						id: projectId
					}
				}
			)
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to send clarification');
				}

				return data.id;
			});
	}

	replyClarification(clarificationId: number, content: string): Promise<number> {
		return apiClient
			.replyUserClarificationToAdmin(
				{
					content
				},
				{
					params: {
						clarificationId
					}
				}
			)
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to reply clarification');
				}

				return data.id;
			});
	}
}
