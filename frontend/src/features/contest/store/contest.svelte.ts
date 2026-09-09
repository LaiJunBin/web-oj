import type {
	Contest,
	ContestDetail,
	ContestSchedule,
	ContestScheduleStorageInterface,
	ContestsStorageInterface,
	ContestStorageInterface
} from '@/apps/contest/domain/contest';
import { BrowserStorage } from '@features/common/store/base.svelte';

export class ContestsStorage
	extends BrowserStorage<Contest[]>
	implements ContestsStorageInterface {}

export class ContestStorage
	extends BrowserStorage<ContestDetail>
	implements ContestStorageInterface
{
	set(data: ContestDetail): void {
		// 換到另一場比賽:記住的 project 屬於舊的那場,安靜丟掉就好。同一場比賽
		// (重新整理、重抓):留著,由根 layout 對照題目清單決定——它丟掉時會說一聲,
		// 使用者才不會看到題目無聲地換了。
		const previousContestId = parseInt(sessionStorage.getItem('currentContestId')!, 10);
		if (previousContestId && previousContestId !== data.id) {
			sessionStorage.removeItem('currentProjectId');
		}

		super.set(data);
		sessionStorage.setItem('currentContestId', data.id.toString());
	}

	clear(): void {
		super.clear();
		sessionStorage.removeItem('currentContestId');
		sessionStorage.removeItem('currentProjectId');
	}
}

export class ContestScheduleStorage
	extends BrowserStorage<ContestSchedule>
	implements ContestScheduleStorageInterface {}
