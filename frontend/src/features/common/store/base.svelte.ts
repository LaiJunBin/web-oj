import { browser } from '$app/environment';
import { type Storage as IStorage } from '@/apps/common/domain/storage';

export class BrowserStorage<T> implements IStorage<T> {
	public error: Error | null = $state(null);
	protected data: T | null = $state(null);

	set(data: T): void {
		if (!browser) {
			return;
		}
		this.data = data;
		this.error = null;
	}

	get(): T | null {
		return this.data;
	}

	clear(): void {
		if (!browser) {
			return;
		}
		this.data = null;
		this.error = null;
	}

	setError(e: Error): void {
		this.data = null;
		this.error = e;
	}
}
