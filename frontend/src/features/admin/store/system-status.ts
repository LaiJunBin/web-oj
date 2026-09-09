import type {
	JudgeQueueStatus,
	JudgeQueueStorageInterface,
	JudgmentError,
	JudgmentErrorsStorageInterface,
	SubmissionRatePoint,
	SubmissionRateStorageInterface
} from '@/apps/admin/domain/system-status';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class JudgeQueueStorage
	extends BrowserStorage<JudgeQueueStatus>
	implements JudgeQueueStorageInterface {}

export class JudgmentErrorsStorage
	extends BrowserStorage<JudgmentError[]>
	implements JudgmentErrorsStorageInterface {}

export class SubmissionRateStorage
	extends BrowserStorage<SubmissionRatePoint[]>
	implements SubmissionRateStorageInterface {}
