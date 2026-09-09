import type {
	Contest,
	ContestDetail,
	ContestParticipants,
	ContestParticipantsStorageInterface,
	ContestsStorageInterface,
	ContestStorageInterface
} from '@/apps/admin/domain/contest';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class ContestsStorage
	extends BrowserStorage<Contest[]>
	implements ContestsStorageInterface {}

export class ContestStorage
	extends BrowserStorage<ContestDetail>
	implements ContestStorageInterface {}

export class ContestParticipantsStorage
	extends BrowserStorage<ContestParticipants>
	implements ContestParticipantsStorageInterface {}
