import type {
	ProjectTestcaseGradingItem,
	ProjectTestcaseGradingItemsStorageInterface,
	Testcase,
	TestcasesStorageInterface,
	TestcaseStorageInterface
} from '@/apps/admin/domain/testcase';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class TestcasesStorage
	extends BrowserStorage<Testcase[]>
	implements TestcasesStorageInterface {}

export class TestcaseStorage extends BrowserStorage<Testcase> implements TestcaseStorageInterface {}

export class ProjectTestcaseGradingItemsStorage
	extends BrowserStorage<ProjectTestcaseGradingItem[]>
	implements ProjectTestcaseGradingItemsStorageInterface {}
