import { apiClient } from '@/api';
import type { ContestServiceInterface } from '@/apps/contest/application/service/contest';
import type { UserProfile } from '@/apps/contest/application/user';
import type { Contest, ContestDetail, ContestSchedule } from '@/apps/contest/domain/contest';

export class ContestService implements ContestServiceInterface {
	constructor(private readonly userProfile: UserProfile) {}
	fetchContests(): Promise<Contest[]> {
		if (this.userProfile.isContestant) {
			return apiClient.listUserContests().then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch contests');
				}

				return data;
			});
		}

		return apiClient.listPublicContests().then((response) => {
			const { success, data } = response;
			if (!success) {
				throw new Error('Failed to fetch contests');
			}

			return data;
		});
	}
	fetchContestDetail(contestId: number): Promise<ContestDetail> {
		if (this.userProfile.isContestant) {
			return apiClient
				.getUserContest({
					params: {
						contestId
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch contest');
					}

					return data;
				});
		}

		return apiClient
			.getPublicContest({
				params: {
					contestId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch contest');
				}

				return data;
			});
	}
	fetchContestSchedule(contestId: number): Promise<ContestSchedule> {
		return apiClient
			.getPublicContestSchedule({
				params: {
					contestId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch contest schedule');
				}

				return data;
			});
	}
}
