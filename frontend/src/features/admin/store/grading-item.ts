import type {
	GradingItem,
	GradingItemDetail,
	GradingItemsStorageInterface,
	GradingItemStorageInterface
} from '@/apps/admin/domain/grading-item';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class GradingItemsStorage
	extends BrowserStorage<GradingItem[]>
	implements GradingItemsStorageInterface {}

export class GradingItemStorage
	extends BrowserStorage<GradingItemDetail>
	implements GradingItemStorageInterface {}
