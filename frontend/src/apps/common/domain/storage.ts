export interface Storage<T> {
	set(data: T): void;
	get(): T | null;
	clear(): void;

	error: Error | null;
	setError(error: Error): void;
}
