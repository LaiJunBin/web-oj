import { Application } from '@/apps/common/application/base';
import type { ContestProjectServiceInterface } from '@/apps/admin/application/service/contest-project';
import type {
	ContestProjectStorageInterface,
	JudgmentAnalyticsStorageInterface,
	RejudgeContestProjectPayload,
	UserScoresStorageInterface
} from '@/apps/admin/domain/contest-project';
import { container } from '@/apps/admin/container';

export class ContestProjectApplication extends Application {
	constructor(
		private readonly contestProjectService: ContestProjectServiceInterface,

		private readonly contestProjectStorage: ContestProjectStorageInterface,
		private readonly userScoresStorage: UserScoresStorageInterface,
		private readonly judgmentAnalyticsStorage: JudgmentAnalyticsStorageInterface
	) {
		super();
	}

	fetchContestProject(id: number) {
		return this.contestProjectService
			.fetchContestProject(id)
			.then((contestProject) => {
				this.contestProjectStorage.set(contestProject);
			})
			.catch((error) => {
				this.contestProjectStorage.setError(error);
			});
	}

	fetchUserScores(contestProjectID: number) {
		return this.contestProjectService
			.fetchUserScores(contestProjectID)
			.then((userScores) => {
				this.userScoresStorage.set(userScores);
			})
			.catch((error) => {
				this.userScoresStorage.setError(error);
			});
	}

	fetchJudgmentAnalytics(contestProjectID: number) {
		return this.contestProjectService
			.fetchJudgmentAnalytics(contestProjectID)
			.then((analytics) => {
				this.judgmentAnalyticsStorage.set(analytics);
			})
			.catch((error) => {
				this.judgmentAnalyticsStorage.setError(error);
			});
	}

	exportContestProjectScoreDetail(contestProjectID: number, contestProjectName: string) {
		return this.contestProjectService.exportContestProjectScoreDetail(contestProjectID, contestProjectName);
	}

	rejudgeContestProject(id: number, payload: RejudgeContestProjectPayload) {
		return this.contestProjectService.rejudgeContestProject(id, payload);
	}
}

export const useContestProjectApp = () => container.get(ContestProjectApplication);
