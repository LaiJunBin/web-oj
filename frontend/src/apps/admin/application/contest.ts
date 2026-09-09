import { Application } from '@/apps/common/application/base';
import type { ContestServiceInterface } from '@/apps/admin/application/service/contest';
import type {
	ContestParticipantsStorageInterface,
	ContestsStorageInterface,
	ContestStorageInterface,
	CreateContestPayload,
	UpdateContestPayload
} from '@/apps/admin/domain/contest';
import { container } from '@/apps/admin/container';

export class ContestApplication extends Application {
	constructor(
		private readonly contestService: ContestServiceInterface,

		private readonly contestsStorage: ContestsStorageInterface,
		private readonly contestStorage: ContestStorageInterface,
		private readonly contestParticipantsStorage: ContestParticipantsStorageInterface
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

	createContest(payload: CreateContestPayload) {
		return this.contestService.createContest(payload);
	}

	updateContest(id: number, payload: UpdateContestPayload) {
		return this.contestService.updateContest(id, payload);
	}

	deleteContest(id: number) {
		return this.contestService.deleteContest(id);
	}

	fetchContestParticipants(id: number) {
		return this.contestService
			.fetchContestParticipants(id)
			.then((participants) => {
				this.contestParticipantsStorage.set(participants);
			})
			.catch((error) => {
				this.contestParticipantsStorage.setError(error);
			});
	}

	addUsersToContest(id: number, userIds: number[]) {
		return this.contestService.addUsersToContest(id, userIds);
	}

	removeUsersFromContest(id: number, userIds: number[]) {
		return this.contestService.removeUsersFromContest(id, userIds);
	}

	addGroupsToContest(id: number, groupIds: number[]) {
		return this.contestService.addGroupsToContest(id, groupIds);
	}

	removeGroupsFromContest(id: number, groupIds: number[]) {
		return this.contestService.removeGroupsFromContest(id, groupIds);
	}

	exportContest(id: number) {
		return this.contestService.exportContest(id);
	}

	importContest(file: File) {
		return this.contestService.importContest(file);
	}
}

export const useContestApp = () => container.get(ContestApplication);
