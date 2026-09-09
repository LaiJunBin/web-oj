import type {
	Clarification,
	ClarificationDetail,
	ClarificationsStorageInterface,
	ClarificationStorageInterface
} from '@/apps/contest/domain/clarification';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class ClarificationsStorage
	extends BrowserStorage<Clarification[]>
	implements ClarificationsStorageInterface {}

export class ClarificationStorage
	extends BrowserStorage<ClarificationDetail>
	implements ClarificationStorageInterface {}
