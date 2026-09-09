import type {
	ContestProject,
	ContestProjectStorageInterface,
	JudgmentAnalytics,
	JudgmentAnalyticsStorageInterface,
	UserScore,
	UserScoresStorageInterface
} from '@/apps/admin/domain/contest-project';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class ContestProjectStorage
	extends BrowserStorage<ContestProject>
	implements ContestProjectStorageInterface {}

export class UserScoresStorage
	extends BrowserStorage<UserScore[]>
	implements UserScoresStorageInterface {}

export class JudgmentAnalyticsStorage
	extends BrowserStorage<JudgmentAnalytics>
	implements JudgmentAnalyticsStorageInterface {}
