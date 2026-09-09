import { Application } from '@/apps/common/application/base';
import type { SystemStatusServiceInterface } from '@/apps/admin/application/service/system-status';
import type {
	JudgeQueueStorageInterface,
	JudgmentErrorsStorageInterface,
	SubmissionRateStorageInterface
} from '@/apps/admin/domain/system-status';
import { container } from '@/apps/admin/container';

export class SystemStatusApplication extends Application {
	constructor(
		private readonly systemStatusService: SystemStatusServiceInterface,

		private readonly judgeQueueStorage: JudgeQueueStorageInterface,
		private readonly judgmentErrorsStorage: JudgmentErrorsStorageInterface,
		private readonly submissionRateStorage: SubmissionRateStorageInterface
	) {
		super();
	}

	fetchJudgeQueue() {
		return this.systemStatusService
			.fetchJudgeQueue()
			.then((status) => {
				this.judgeQueueStorage.set(status);
			})
			.catch((error) => {
				this.judgeQueueStorage.setError(error);
			});
	}

	fetchJudgmentErrors(limit: number) {
		return this.systemStatusService
			.fetchJudgmentErrors(limit)
			.then((errors) => {
				this.judgmentErrorsStorage.set(errors);
			})
			.catch((error) => {
				this.judgmentErrorsStorage.setError(error);
			});
	}

	fetchSubmissionRate(minutes: number, contestProjectId?: number) {
		return this.systemStatusService
			.fetchSubmissionRate(minutes, contestProjectId)
			.then((points) => {
				this.submissionRateStorage.set(points);
			})
			.catch((error) => {
				this.submissionRateStorage.setError(error);
			});
	}
}

export const useSystemStatusApp = () => container.get(SystemStatusApplication);
