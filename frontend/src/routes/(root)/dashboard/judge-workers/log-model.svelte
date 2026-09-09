<script lang="ts">
	import {
		JudgeWorkerLogAddedEventDataSchema,
		type JudgeWorkerLog
	} from '@/apps/admin/domain/judge-worker';
	import { datetimeString } from '@/lib';
	import { streamJudgeWorkerLog } from '@/lib/api/stream';
	import { goto } from '$app/navigation';
	import { Modal } from 'flowbite-svelte';
	import { tick } from 'svelte';

	let { id, open = $bindable() } = $props();

	let pre: HTMLElement | null = $state(null);
	let follow = $state(true);
	let logs: JudgeWorkerLog[] = $state([]);
	let eventSource: EventSource | null = null;
	$effect(() => {
		logs = [];
		if (id && open) {
			eventSource = streamJudgeWorkerLog(id);
			eventSource.addEventListener('open', (event) => {
			});
			eventSource.addEventListener('message', (event) => {
				const { success, data, error } = JudgeWorkerLogAddedEventDataSchema.safeParse(
					JSON.parse(event.data)
				);
				if (!success) {
					console.error(error);
					return;
				}

				logs.push({
					log: data.log,
					timestamp: data.timestamp
				});
				if (follow) {
					tick().then(() => {
						pre?.scrollTo(0, pre?.scrollHeight);
					});
				}
			});
			eventSource.addEventListener('error', (event) => {
				console.error(event);
				const status = (event as MessageEvent & { status?: number }).status;
				if (status === 401 || status === 403) {
					eventSource?.close();
					goto('/login');
				}
			});
			return () => {
				eventSource?.close();
			};
		}
	});
</script>

<Modal bind:open outsideclose>
	<div class="flex h-full flex-col p-4">
		<h1
			class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
		>
			<span>評分工作者</span>
		</h1>
		{#if logs.length === 0}
			<div class="text-center text-lg font-bold">No logs</div>
		{:else}
			<code
				class="mt-4 flex-1 overflow-auto rounded-lg bg-gray-100 p-4 dark:bg-gray-700"
				bind:this={pre}
				onscroll={(e) => {
					if (!pre) return;
					follow = pre.scrollHeight - pre.scrollTop === pre.clientHeight;
				}}
			>
				{#each logs as { log, timestamp }, i}
					<span class="text-gray-500 dark:text-gray-400">{datetimeString(timestamp)}</span>
					<div class="whitespace-pre text-gray-900 dark:text-white">{log}</div>
				{/each}
			</code>
		{/if}
	</div>
</Modal>
