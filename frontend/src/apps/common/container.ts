import { LoadingStorage } from '@/features/common/store/loading.svelte';
import { Container } from '@/lib/container';

export const container = new Container();

export function initCommonContainer() {
	container.set(LoadingStorage);
}
