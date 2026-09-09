import { Application } from '@/apps/common/application/base';
import type { RejudgeServiceInterface } from '@/apps/admin/application/service/rejudge';
import type {
	RejudgesStorageInterface,
	RejudgeStorageInterface,
	RejudgeComparisonStorageInterface
} from '@/apps/admin/domain/rejudge';
import { container } from '@/apps/admin/container';

export class RejudgeApplication extends Application {
	constructor(
		private readonly rejudgeService: RejudgeServiceInterface,

		private readonly rejudgesStorage: RejudgesStorageInterface,
		private readonly rejudgeStorage: RejudgeStorageInterface,
		private readonly rejudgeComparisonStorage: RejudgeComparisonStorageInterface
	) {
		super();
	}

	fetchRejudges() {
		return this.rejudgeService
			.fetchRejudges()
			.then((rejudges) => {
				this.rejudgesStorage.set(rejudges);
			})
			.catch((error) => {
				this.rejudgesStorage.setError(error);
			});
	}

	fetchRejudgesByContestProject(contestProjectId: number) {
		return this.rejudgeService
			.fetchRejudgesByContestProject(contestProjectId)
			.then((rejudges) => {
				this.rejudgesStorage.set(rejudges);
			})
			.catch((error) => {
				this.rejudgesStorage.setError(error);
			});
	}

	fetchRejudge(id: string) {
		return this.rejudgeService
			.fetchRejudgeDetail(id)
			.then((rejudge) => {
				this.rejudgeStorage.set(rejudge);
			})
			.catch((error) => {
				this.rejudgeStorage.setError(error);
			});
	}

	fetchSubmissionComparison(rejudgeId: string, submissionId: string) {
		return this.rejudgeService
			.fetchSubmissionComparison(rejudgeId, submissionId)
			.then((comparison) => {
				this.rejudgeComparisonStorage.set(comparison);
			})
			.catch((error) => {
				this.rejudgeComparisonStorage.setError(error);
			});
	}

	cancelRejudge(id: string) {
		return this.rejudgeService.cancelRejudge(id);
	}

	approveRejudge(id: string) {
		return this.rejudgeService.approveRejudge(id);
	}

	rejectRejudge(id: string) {
		return this.rejudgeService.rejectRejudge(id);
	}
}

export const useRejudgeApp = () => container.get(RejudgeApplication);
