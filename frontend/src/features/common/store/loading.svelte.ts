import { untrack } from 'svelte';

import { type LoadingStorageInterface } from '@/apps/common/domain/loading';
export class LoadingStorage implements LoadingStorageInterface {
	private depth = $state(0);

	increase(): void {
		untrack(() => {
			this.depth += 1;
		});
	}

	decrease(): void {
		untrack(() => {
			this.depth -= 1;
		});
	}

	get isLoading(): boolean {
		return this.depth > 0;
	}
}
