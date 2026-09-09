<script lang="ts">
	import { Badge, Card } from 'flowbite-svelte';
	import { datetimeString } from '$lib';
	import { useProjectStorage } from '@/apps/contest/domain/project';
	import { onMount } from 'svelte';
	import { useProjectApp } from '@/apps/contest/application/project';

	const projectApp = useProjectApp();
	const projectStorage = useProjectStorage();
	const project = $derived(projectStorage.get());
	const submissions = $derived(project?.submissions ?? []);
	const statusColorMap: Record<string, 'dark' | 'red' | 'yellow' | 'green' | 'blue'> = {
		Created: 'dark',
		Queued: 'dark',
		Judging: 'blue',
		Judged: 'green',
		Cancelled: 'yellow',
		Error: 'red'
	};

	onMount(() => {
		if (!project) return;
		projectApp.withLoading().fetchProject(project.id);
	});
</script>

{#if !project}
	<p>Loading...</p>
{:else}
	<div class="p-4">
		<h1
			class="mb-4 flex items-center space-x-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
		>
			<span>{project.name} 提交紀錄 ({submissions.length})</span>
		</h1>

		{#if submissions.length === 0}
			<p>尚未有任何提交紀錄</p>
		{:else}
			{#each submissions as submission, i}
				<Card class="mb-4 max-w-none" href="/submissions/{submission.id}">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2">
							<span class="text-gray-600 dark:text-gray-400">#{i + 1}</span>
							<Badge color={submission.isPartial ? 'pink' : 'indigo'}>
								{submission.isPartial ? '部分提交' : '完整提交'}</Badge
							>
							<Badge color={statusColorMap[submission.status]}>{submission.status}</Badge>
							<Badge color="blue">
								得分 ({submission.score} / {submission.totalScore})
							</Badge>
						</div>
						<div class="mr-4 ml-auto text-gray-600 dark:text-gray-400">
							提交時間: {datetimeString(submission.createdAt)}
						</div>
					</div>
					{#if submission.staffNote}
						<!-- 有工作人員附註就在卡片下緣顯示一行,樣式比照詳情頁的「工作人員附註」(藍),
						     清楚但不搶版面。與 reason 一樣:進頁抓取、不做即時更新。 -->
						<div
							class="mt-2 border-t border-blue-100 pt-2 text-sm text-blue-800 dark:border-blue-900/60 dark:text-blue-300"
						>
							<span class="font-bold">工作人員附註:</span>{submission.staffNote}
						</div>
					{/if}
				</Card>
			{/each}
		{/if}
	</div>
{/if}
