import type { Scoreboard } from '@apps/contest/domain/scoreboard';

export interface ScoreboardServiceInterface {
	fetchScoreboard(projectId: number): Promise<Scoreboard>;
}
