import { Application } from '@/apps/common/application/base';
import type { ClarificationServiceInterface } from '@/apps/admin/application/service/clarification';
import type {
	CreateClarificationPayload,
	ClarificationsStorageInterface,
	ClarificationStorageInterface,
	ReplyClarificationPayload
} from '@/apps/admin/domain/clarification';
import { container } from '@/apps/admin/container';

export class ClarificationApplication extends Application {
	constructor(
		private readonly clarificationService: ClarificationServiceInterface,

		private readonly clarificationsStorage: ClarificationsStorageInterface,
		private readonly clarificationStorage: ClarificationStorageInterface
	) {
		super();
	}

	fetchClarifications(contestProjectId: number) {
		return this.clarificationService
			.fetchClarifications(contestProjectId)
			.then((clarifications) => {
				this.clarificationsStorage.set(clarifications);
			})
			.catch((error) => {
				this.clarificationsStorage.setError(error);
			});
	}

	fetchClarification(id: number) {
		return this.clarificationService
			.fetchClarificationDetail(id)
			.then((clarification) => {
				this.clarificationStorage.set(clarification);
			})
			.catch((error) => {
				this.clarificationStorage.setError(error);
			});
	}

	sendClarification(contestProjectId: number, payload: CreateClarificationPayload) {
		return this.clarificationService.sendClarification(contestProjectId, payload);
	}

	replyClarification(id: number, payload: ReplyClarificationPayload) {
		return this.clarificationService.replyClarification(id, payload);
	}

	markClarificationAsAnswered(id: number) {
		return this.clarificationService.markClarificationAsAnswered(id);
	}

	markClarificationAsUnanswered(id: number) {
		return this.clarificationService.markClarificationAsUnanswered(id);
	}
}

export const useClarificationApp = () => container.get(ClarificationApplication);
