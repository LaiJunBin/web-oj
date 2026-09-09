<script lang="ts">
	import { goto } from '$app/navigation';
	import { useRejudgeApp } from '@/apps/admin/application/rejudge';
	import { useRejudgesStorage, type Rejudge } from '@/apps/admin/domain/rejudge';
	import { adminCurrentProject } from '$lib/adminCurrentProject.svelte';
	import StateMessage from '$lib/components/StateMessage/StateMessage.svelte';
	import { datetimeString } from '@/lib';
	import {
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';

	const rejudgeApp = useRejudgeApp();
	const rejudges = $derived(useRejudgesStorage().get());

	// 與其他競賽過程頁一致:以側欄選定的競賽專案為範圍。
	$effect(() => {
		const projectId = adminCurrentProject.projectId;
		if (projectId == null) return;
		rejudgeApp.withLoading().fetchRejudgesByContestProject(projectId);
	});

	const rejudgeRowClass = (rejudge: Rejudge) => {
		switch (rejudge.status) {
			case 'Approved':
				return 'bg-green-100 hover:bg-green-200 dark:bg-green-700 dark:hover:bg-green-800';
			case 'Rejected':
				return 'bg-red-100 hover:bg-red-200 dark:bg-red-800 dark:hover:bg-red-900';
			case 'Finished':
				return 'bg-amber-100 hover:bg-amber-200 dark:bg-amber-700 dark:hover:bg-amber-800';
			default:
				return 'hover:bg-gray-100 dark:hover:bg-gray-700';
		}
	};

	const doneRejudges = $derived(
		rejudges?.filter(
			(rejudge) =>
				rejudge.status === 'Approved' ||
				rejudge.status === 'Rejected' ||
				rejudge.status === 'Canceled'
		) || []
	);
	const remainingRejudges = $derived(
		rejudges?.filter(
			(rejudge) =>
				rejudge.status !== 'Approved' &&
				rejudge.status !== 'Rejected' &&
				rejudge.status !== 'Canceled'
		) || []
	);
</script>

{#if adminCurrentProject.projectId == null}
	<StateMessage
		title="尚未選擇競賽專案"
		description="請從左側的「競賽過程」選擇比賽與專案。"
	/>
{:else if !rejudges}
	<div>Loading...</div>
{:else}
	<div class="flex flex-col p-4">
		<h2 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			重評紀錄({remainingRejudges.length})
		</h2>

		<Table>
			<TableHead>
				<TableHeadCell>編號</TableHeadCell>
				<TableHeadCell>重新評分原因</TableHeadCell>
				<TableHeadCell>受影響的提交數量</TableHeadCell>
				<TableHeadCell>進度</TableHeadCell>
				<TableHeadCell>狀態</TableHeadCell>
				<TableHeadCell>時間</TableHeadCell>
			</TableHead>
			<TableBody>
				{#each remainingRejudges as rejudge, index}
					<TableBodyRow
						class="cursor-pointer {rejudgeRowClass(rejudge)}"
						onclick={() => {
							goto(`/dashboard/rejudgings/${rejudge.id}`);
						}}
					>
						<TableBodyCell>{remainingRejudges.length - index}</TableBodyCell>
						<TableBodyCell>
							{rejudge.reason}
						</TableBodyCell>
						<TableBodyCell>
							{rejudge.effectedJudgments}
						</TableBodyCell>
						<TableBodyCell>
							{rejudge.completedJudgments} / {rejudge.effectedJudgments}
						</TableBodyCell>
						<TableBodyCell>
							{rejudge.status}
						</TableBodyCell>
						<TableBodyCell>{datetimeString(rejudge.createdAt)}</TableBodyCell>
					</TableBodyRow>
				{:else}
					<TableBodyRow>
						<TableBodyCell colspan={6} class="text-center">沒有重評紀錄</TableBodyCell>
					</TableBodyRow>
				{/each}
			</TableBody>
		</Table>

		<hr class="my-4" />

		<h2 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			已完成的重評紀錄({doneRejudges.length})
		</h2>

		<Table>
			<TableHead>
				<TableHeadCell>編號</TableHeadCell>
				<TableHeadCell>重新評分原因</TableHeadCell>
				<TableHeadCell>受影響的提交數量</TableHeadCell>
				<TableHeadCell>進度</TableHeadCell>
				<TableHeadCell>狀態</TableHeadCell>
				<TableHeadCell>時間</TableHeadCell>
			</TableHead>
			<TableBody>
				{#each doneRejudges as rejudge, index}
					<TableBodyRow
						class="cursor-pointer {rejudgeRowClass(rejudge)}"
						onclick={() => {
							goto(`/dashboard/rejudgings/${rejudge.id}`);
						}}
					>
						<TableBodyCell>{doneRejudges.length - index}</TableBodyCell>
						<TableBodyCell>
							{rejudge.reason}
						</TableBodyCell>
						<TableBodyCell>
							{rejudge.effectedJudgments}
						</TableBodyCell>
						<TableBodyCell>
							{rejudge.completedJudgments} / {rejudge.effectedJudgments}
						</TableBodyCell>
						<TableBodyCell>
							{rejudge.status}
						</TableBodyCell>
						<TableBodyCell>{datetimeString(rejudge.createdAt)}</TableBodyCell>
					</TableBodyRow>
				{:else}
					<TableBodyRow>
						<TableBodyCell colspan={6} class="text-center">沒有已完成的重評紀錄</TableBodyCell>
					</TableBodyRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
