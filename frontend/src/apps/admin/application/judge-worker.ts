import { Application } from '@/apps/common/application/base';
import type { JudgeWorkerServiceInterface } from '@/apps/admin/application/service/judge-worker';
import type {
	CreateJudgeWorkerPayload,
	JudgePriorityStrategiesStorageInterface,
	JudgeWorkersStorageInterface,
	SystemDefaultsStorageInterface
} from '@/apps/admin/domain/judge-worker';
import { container } from '@/apps/admin/container';

export class JudgeWorkerApplication extends Application {
	constructor(
		private readonly judgeWorkerService: JudgeWorkerServiceInterface,

		private readonly judgeWorkersStorage: JudgeWorkersStorageInterface,
		private readonly judgePriorityStrategiesStorage: JudgePriorityStrategiesStorageInterface,
		private readonly systemDefaultsStorage: SystemDefaultsStorageInterface
	) {
		super();
	}

	fetchJudgeWorkers() {
		return this.judgeWorkerService
			.fetchJudgeWorkers()
			.then((judgeWorkers) => {
				this.judgeWorkersStorage.set(judgeWorkers);
			})
			.catch((error) => {
				this.judgeWorkersStorage.setError(error);
			});
	}

	/**
	 * Refreshes the worker list after a stream reconnect, or periodically while
	 * the stream is down. Best-effort: on failure nothing changes, not even
	 * `error`. `setError` would clear the list, and an empty list renders as
	 * 「尚無評分工作者」 — telling an operator mid-outage that there are no
	 * workers, when what is true is that the last known workers are still the
	 * best information on the screen.
	 */
	resyncJudgeWorkers() {
		return this.judgeWorkerService
			.fetchJudgeWorkers()
			.then((judgeWorkers) => {
				this.judgeWorkersStorage.set(judgeWorkers);
			})
			.catch((error) => {
				console.warn('[judge-workers] resync failed; keeping what is on screen', error);
			});
	}

	fetchJudgePriorityStrategies() {
		return this.judgeWorkerService
			.fetchJudgePriorityStrategies()
			.then((strategies) => {
				this.judgePriorityStrategiesStorage.set(strategies);
			})
			.catch((error) => {
				this.judgePriorityStrategiesStorage.setError(error);
			});
	}

	fetchSystemDefaults() {
		return this.judgeWorkerService
			.fetchSystemDefaults()
			.then((defaults) => {
				this.systemDefaultsStorage.set(defaults);
			})
			.catch((error) => {
				this.systemDefaultsStorage.setError(error);
			});
	}

	createJudgeWorker(payload: CreateJudgeWorkerPayload) {
		return this.judgeWorkerService.createJudgeWorker(payload);
	}

	deleteJudgeWorker(id: string) {
		return this.judgeWorkerService.deleteJudgeWorker(id);
	}

	startJudgeWorker(id: string) {
		return this.judgeWorkerService.startJudgeWorker(id);
	}

	stopJudgeWorker(id: string) {
		return this.judgeWorkerService.stopJudgeWorker(id);
	}
}

export const useJudgeWorkerApp = () => container.get(JudgeWorkerApplication);
