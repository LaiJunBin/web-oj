import { container } from '@/apps/contest/container';
import type { ScoreboardStorageInterface } from '@apps/contest/domain/scoreboard';
import { Application } from '@apps/common/application/base';
import type { ScoreboardServiceInterface } from './service/scoreboard';

export class ScoreboardApplication extends Application {
	constructor(
		private readonly scoreboardService: ScoreboardServiceInterface,

		private readonly scoreboardStorage: ScoreboardStorageInterface
	) {
		super();
	}

	fetchScoreboard(projectId: number) {
		return this.scoreboardService
			.fetchScoreboard(projectId)
			.then((scoreboard) => {
				this.scoreboardStorage.set(scoreboard);
			})
			.catch((error) => {
				this.scoreboardStorage.setError(error);
			});
	}

	/**
	 * Refreshes the board over REST after a stream reconnect, or periodically
	 * while the stream is down. This is a best-effort fallback, so failure
	 * changes nothing — not even `error`. `setError` clears the data, and the
	 * one time this runs is precisely when the board is holding the last good
	 * scores for a venue screen whose backend just went away; wiping them turns
	 * "stale but visible" into "Loading...". The stream notice already tells the
	 * viewer the connection is bad; the store does not need to say it again.
	 */
	resyncScoreboard(projectId: number) {
		return this.scoreboardService
			.fetchScoreboard(projectId)
			.then((scoreboard) => {
				this.scoreboardStorage.set(scoreboard);
			})
			.catch((error) => {
				console.warn('[scoreboard] resync failed; keeping what is on screen', error);
			});
	}
}

export const useScoreboardApp = () => container.get(ScoreboardApplication);
