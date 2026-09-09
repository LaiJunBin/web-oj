<script lang="ts">
	import {
		JudgeWorkersEventDataSchema,
		useJudgeWorkersStorage
	} from '@/apps/admin/domain/judge-worker';
	import {
		Button,
		ButtonGroup,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';
	import { datetimeString } from '@/lib';
	import { useJudgeWorkerApp } from '@/apps/admin/application/judge-worker';
	import LogModel from '@/routes/(root)/dashboard/judge-workers/log-model.svelte';
	import { onMount } from 'svelte';
	import { streamJudgeWorkers } from '@/lib/api/stream';

	const judgeWorkerApp = useJudgeWorkerApp();
	const judgeWorkersStorage = useJudgeWorkersStorage();
	const judgeWorkers = $derived(judgeWorkersStorage.get()!);

	let openLogModal = $state(false);
	let selectedJudgeWorkerId = $state<string | null>(null);

	onMount(() => {
		const eventSource = streamJudgeWorkers();
		eventSource.addEventListener('message', (event) => {
			try {
				const { success, data, error } = JudgeWorkersEventDataSchema.safeParse(JSON.parse(event.data));
				if (!success) {
					console.error(error);
					return;
				}

				judgeWorkersStorage.set(data.data);
			} catch (e) {
				console.error('Failed to parse judge workers event', e);
			}
		});
		return () => {
			eventSource.close();
		};
	});

</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>評分工作者管理</span>
		<Button href="/dashboard/judge-workers/create" color="green">新增</Button>
	</h1>

	<Table striped hoverable shadow>
		<TableHead>
			<TableHeadCell>#</TableHeadCell>
			<TableHeadCell>狀態</TableHeadCell>
			<TableHeadCell>工作優先順序</TableHeadCell>
			<TableHeadCell>最後執行時間</TableHeadCell>
			<TableHeadCell>
				<span class="sr-only">Actions</span>
			</TableHeadCell>
		</TableHead>
		<TableBody tableBodyClass="divide-y">
			{#each judgeWorkers as judgeWorker, index}
				<TableBodyRow
					class="cursor-pointer"
					onclick={(e) => {
						const target = e.target as HTMLElement;
						if (target.tagName === 'TD') {
							openLogModal = true;
							selectedJudgeWorkerId = judgeWorker.id;
						}
					}}
				>
					<TableBodyCell>{index + 1}</TableBodyCell>
					<TableBodyCell>{judgeWorker.status}</TableBodyCell>
					<TableBodyCell>{judgeWorker.judgeOrder.join(', ')}</TableBodyCell>
					<TableBodyCell
						>{judgeWorker.lastExecutionAt
							? datetimeString(judgeWorker.lastExecutionAt)
							: '無'}</TableBodyCell
					>
					<TableBodyCell>
						<ButtonGroup>
							<Button
								pill
								color="blue"
								class="cursor-pointer"
								onclick={() => {
									openLogModal = true;
									selectedJudgeWorkerId = judgeWorker.id;
								}}>檢視日誌</Button
							>
							<Button
								pill
								color="red"
								class="cursor-pointer"
								disabled={judgeWorker.status === 'RUNNING'}
								onclick={() => {
									Swal.fire({
										title: '確定要刪除嗎？',
										text: `確定要刪除嗎？`,
										icon: 'warning',
										showCancelButton: true,
										confirmButtonText: '確定',
										cancelButtonText: '取消'
									}).then((result) => {
										if (!result.isConfirmed) {
											return;
										}

										judgeWorkerApp
											.deleteJudgeWorker(judgeWorker.id)
											.then(() => {
												Swal.fire({
													title: '刪除成功',
													icon: 'success',
													confirmButtonText: '確定'
												});
												judgeWorkerApp.fetchJudgeWorkers();
											})
											.catch((error) => {
												console.error(error);
												Swal.fire({
													title: '刪除失敗',
													icon: 'error',
													confirmButtonText: '確定'
												});
											});
									});
								}}>刪除</Button
							>
						</ButtonGroup>
					</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>
</div>

<LogModel id={selectedJudgeWorkerId} bind:open={openLogModal} />
