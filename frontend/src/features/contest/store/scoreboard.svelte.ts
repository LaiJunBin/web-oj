import type { Scoreboard } from '@/apps/contest/domain/scoreboard';
import type { ScoreboardStorageInterface } from '@/apps/contest/domain/scoreboard';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class ScoreboardStorage
	extends BrowserStorage<Scoreboard>
	implements ScoreboardStorageInterface {}
