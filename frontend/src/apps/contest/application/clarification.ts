import {
	type ClarificationsStorageInterface,
	type ClarificationStorageInterface
} from '@apps/contest/domain/clarification';
import { Application } from '@apps/common/application/base';
import type { ClarificationServiceInterface } from './service/clarification';
import { container } from '@/apps/contest/container';

export class ClarificationApplication extends Application {
	constructor(
		private readonly clarificationService: ClarificationServiceInterface,

		private readonly clarificationStorage: ClarificationStorageInterface,
		private readonly clarificationsStorage: ClarificationsStorageInterface
	) {
		super();
	}

	fetchClarifications(projectId: number) {
		return this.clarificationService
			.fetchClarifications(projectId)
			.then((clarifications) => {
				this.clarificationsStorage.set(clarifications);
			})
			.catch((error) => {
				this.clarificationsStorage.setError(error);
			});
	}

	fetchClarification(id: number) {
		return this.clarificationService
			.fetchClarification(id)
			.then((clarification) => {
				this.clarificationStorage.set(clarification);
			})
			.catch((error) => {
				this.clarificationStorage.setError(error);
			});
	}

	sendClarification(projectId: number, subject: string, content: string) {
		return this.clarificationService.sendClarification(projectId, subject, content);
	}

	replyClarification(clarificationId: number, content: string) {
		return this.clarificationService.replyClarification(clarificationId, content);
	}
}

export const useClarificationApp = () => container.get(ClarificationApplication);
