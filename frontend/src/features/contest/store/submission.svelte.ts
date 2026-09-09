import type {
	Submission,
	SubmissionDetail,
	SubmissionsStorageInterface,
	SubmissionStorageInterface
} from '@/apps/contest/domain/submission';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class SubmissionsStorage
	extends BrowserStorage<Submission[]>
	implements SubmissionsStorageInterface {}
export class SubmissionStorage
	extends BrowserStorage<SubmissionDetail>
	implements SubmissionStorageInterface {}
