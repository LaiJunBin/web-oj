import { apiClient } from '@/api';
import type { ScoreboardServiceInterface } from '@/apps/contest/application/service/scoreboard';
import type { UserProfile } from '@/apps/contest/application/user';
import type { Scoreboard } from '@/apps/contest/domain/scoreboard';

export class ScoreboardService implements ScoreboardServiceInterface {
	constructor(private readonly userProfile: UserProfile) {}
	fetchScoreboard(projectId: number): Promise<Scoreboard> {
		if (this.userProfile.isContestant) {
			return apiClient
				.getUserProjectScoreboard({
					params: {
						id: projectId
					}
				})
				.then((response) => {
					const { success, data } = response;
					if (!success) {
						throw new Error('Failed to fetch scoreboard');
					}

					return data;
				});
		}

		return apiClient
			.getPublicProjectScoreboard({
				params: {
					id: projectId
				}
			})
			.then((response) => {
				const { success, data } = response;
				if (!success) {
					throw new Error('Failed to fetch scoreboard');
				}

				return data;
			});
	}
}
