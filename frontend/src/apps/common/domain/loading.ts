import { container } from '@/apps/common/container';

export interface LoadingStorageInterface {
	increase(): void;
	decrease(): void;
	isLoading: boolean;
}
export const useLoadingStorage = () => container.get('LoadingStorage') as LoadingStorageInterface;
