import type {
	Rejudge,
	RejudgeDetail,
	SubmissionComparisonDetail,
	RejudgesStorageInterface,
	RejudgeStorageInterface,
	RejudgeComparisonStorageInterface
} from '@/apps/admin/domain/rejudge';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class RejudgesStorage
	extends BrowserStorage<Rejudge[]>
	implements RejudgesStorageInterface {}

export class RejudgeStorage
	extends BrowserStorage<RejudgeDetail>
	implements RejudgeStorageInterface {}

export class RejudgeComparisonStorage
	extends BrowserStorage<SubmissionComparisonDetail>
	implements RejudgeComparisonStorageInterface {}
