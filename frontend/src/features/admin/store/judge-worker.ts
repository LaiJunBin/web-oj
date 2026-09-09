import type {
	JudgePriorityStrategiesStorageInterface,
	JudgeWorker,
	JudgeWorkersStorageInterface,
	SystemDefaults,
	SystemDefaultsStorageInterface
} from '@/apps/admin/domain/judge-worker';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class JudgeWorkersStorage
	extends BrowserStorage<JudgeWorker[]>
	implements JudgeWorkersStorageInterface {}

export class JudgePriorityStrategiesStorage
	extends BrowserStorage<string[]>
	implements JudgePriorityStrategiesStorageInterface {}

export class SystemDefaultsStorage
	extends BrowserStorage<SystemDefaults>
	implements SystemDefaultsStorageInterface {}
