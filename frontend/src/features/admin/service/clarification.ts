import { apiClient } from '@/api';
import type { ClarificationServiceInterface } from '@/apps/admin/application/service/clarification';
import type {
	CreateClarificationPayload,
	Clarification,
	ClarificationDetail,
	ReplyClarificationPayload
} from '@/apps/admin/domain/clarification';

export class ClarificationService implements ClarificationServiceInterface {
	fetchClarifications(contestProjectId: number): Promise<Clarification[]> {
		return apiClient
			.listClarifications({
				params: {
					id: contestProjectId
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
	fetchClarificationDetail(clarificationId: number): Promise<ClarificationDetail> {
		return apiClient
			.getClarification({
				params: {
					id: clarificationId
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
	sendClarification(
		contestProjectId: number,
		payload: CreateClarificationPayload
	): Promise<number> {
		return apiClient
			.sendClarification(payload, {
				params: {
					id: contestProjectId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to create clarification');
				}
				return data;
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	replyClarification(id: number, payload: ReplyClarificationPayload): Promise<number> {
		return apiClient
			.replyClarificationToUser(payload, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to reply clarification');
				}
				return data;
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	markClarificationAsAnswered(id: number): Promise<void> {
		return apiClient
			.markClarificationAsAnswered(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to mark clarification as answered');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
	markClarificationAsUnanswered(id: number): Promise<void> {
		return apiClient
			.markClarificationAsUnanswered(undefined, {
				params: {
					id
				}
			})
			.then((response) => {
				const { success } = response;
				if (!success) {
					throw new Error('Failed to mark clarification as unanswered');
				}
			})
			.catch((error) => {
				if ('response' in error) {
					const { data } = error.response;
					throw new Error(data.message);
				}

				throw error;
			});
	}
}
