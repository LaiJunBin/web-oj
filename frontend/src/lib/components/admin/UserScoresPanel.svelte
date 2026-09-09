<script lang="ts">
	import { goto } from '$app/navigation';
	import { useContestProjectApp } from '@/apps/admin/application/contest-project';
	import { useUserScoresStorage } from '@/apps/admin/domain/contest-project';
	import {
		Button,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	// Admin view of a contest project's per-user scores (with CSV export). Shared
	// between the 競賽過程 page and the 競賽管理 project page.
	const { projectId, projectName }: { projectId: number; projectName: string } = $props();

	const contestProjectApp = useContestProjectApp();
	const userScores = $derived(useUserScoresStorage().get() || []);
	$effect(() => {
		contestProjectApp.withLoading().fetchUserScores(projectId);
	});
</script>

<div class="mb-4 flex items-center gap-2">
	<span class="font-bold">當前專案:</span>
	<span>{projectName}</span>
	<Button
		color="green"
		class="ml-auto cursor-pointer"
		onclick={() => {
			contestProjectApp
				.withLoading()
				.exportContestProjectScoreDetail(projectId, projectName)
				.catch((error) => {
					Swal.fire({
						icon: 'error',
						title: '匯出失敗',
						text: error.message
					});
				});
		}}
	>
		匯出分數明細
	</Button>
</div>

<Table>
	<TableHead>
		<TableHeadCell>#</TableHeadCell>
		<TableHeadCell>使用者</TableHeadCell>
		<TableHeadCell>分數</TableHeadCell>
	</TableHead>
	<TableBody>
		{#each userScores as userScore, index}
			<TableBodyRow
				class={userScore.submissionId
					? 'cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700'
					: ''}
				onclick={(e) => {
					const target = e.target as HTMLElement;
					if (target.tagName === 'TD' && userScore.submissionId) {
						goto(`/dashboard/submissions/${userScore.submissionId}`);
					}
				}}
			>
				<TableBodyCell>{index + 1}</TableBodyCell>
				<TableBodyCell>
					{userScore.user.name}({userScore.user.username})
				</TableBodyCell>
				<TableBodyCell>{userScore.score}</TableBodyCell>
			</TableBodyRow>
		{:else}
			<TableBodyRow>
				<TableBodyCell colspan={5} class="text-center">沒有使用者</TableBodyCell>
			</TableBodyRow>
		{/each}
	</TableBody>
</Table>
