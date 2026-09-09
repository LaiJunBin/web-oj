import { container } from '@/apps/contest/container';
import type { Storage } from '@apps/common/domain/storage';

export class Time {
	type: string | null = null;
	time: string | null = null;

	constructor(type: string, time: string) {
		this.type = type;
		this.time = time.replace(/(\d+)([dhms])/g, (_, num, unit) => {
			return num.padStart(2, '0') + unit;
		});
	}

	isInWaitContestStart() {
		return this.type === 'wait_start_submit' && this.time !== '00s';
	}

	isInSubmitPeriod() {
		return this.type === 'submit' && this.time !== '00s';
	}

	isContestEnd() {
		return this.type === 'submit' && this.time === '00s';
	}
}

export interface TimeStorageInterface extends Storage<Time> {}
export const useTimeStorage = () => container.get('TimeStorage') as TimeStorageInterface;
