<script lang="ts">
	import { onMount } from 'svelte';
	import type { LoadingStorageInterface } from '@/apps/common/domain/loading';

	let loadingStorage: LoadingStorageInterface | null = $state(null);
	let isLoading = $derived.by(() => (loadingStorage === null ? true : loadingStorage.isLoading));

	onMount(async () => {
		const { useLoadingStorage } = await import('@/apps/common/domain/loading');
		loadingStorage = useLoadingStorage();
	});
</script>

<div
	class="absolute inset-0 z-50 flex items-center justify-center bg-gray-200/50 dark:bg-gray-900/50"
	hidden={!isLoading}
>
	<div
		class="h-32 w-32 animate-spin rounded-full border-t-2 border-b-2 border-gray-900 dark:border-gray-100"
	></div>
</div>
