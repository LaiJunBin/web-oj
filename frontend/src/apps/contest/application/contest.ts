import { container } from '@/apps/contest/container';
import type {
	ContestsStorageInterface,
	ContestStorageInterface,
	ContestScheduleStorageInterface
} from '@apps/contest/domain/contest';
import { Application } from '@apps/common/application/base';
import type { ContestServiceInterface } from './service/contest';

export class ContestApplication extends Application {
	constructor(
		private readonly contestService: ContestServiceInterface,

		private readonly contestsStorage: ContestsStorageInterface,
		private readonly contestStorage: ContestStorageInterface,
		private readonly contestScheduleStorage: ContestScheduleStorageInterface
	) {
		super();
	}

	fetchContests() {
		return this.contestService
			.fetchContests()
			.then((contests) => {
				this.contestsStorage.set(contests);
			})
			.catch((error) => {
				this.contestsStorage.setError(error);
			});
	}

	fetchContest(id: number) {
		return this.contestService
			.fetchContestDetail(id)
			.then((contest) => {
				this.contestStorage.set(contest);
			})
			.catch((error) => {
				this.contestStorage.setError(error);
			});
	}

	fetchContestSchedule(id: number) {
		return this.contestService
			.fetchContestSchedule(id)
			.then((schedule) => {
				this.contestScheduleStorage.set(schedule);
			})
			.catch((error) => {
				this.contestScheduleStorage.setError(error);
			});
	}

	getCurrentAvailableContests() {
		return (
			this.contestsStorage
				.get()
				?.filter((contest) => contest.id !== this.contestStorage.get()?.id) || []
		);
	}
}

export const useContestApp = () => container.get(ContestApplication);
