<script lang="ts">
	import { page } from '$app/state';
	import { useJudgmentAnalyticsStorage } from '@/apps/admin/domain/contest-project';
	import { Badge, Button, Card } from 'flowbite-svelte';

	const analyticsStorage = useJudgmentAnalyticsStorage();
	const analytics = $derived(analyticsStorage.get());

	const statusColorMap: Record<string, 'dark' | 'red' | 'yellow' | 'green' | 'blue' | 'indigo'> = {
		Judged: 'green',
		Judging: 'blue',
		Queued: 'dark',
		Created: 'dark',
		Cancelled: 'yellow',
		Error: 'red'
	};
</script>

{#if !analytics}
	<div>Loading...</div>
{:else}
	<div class="flex flex-col p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/dashboard/contests/{page.params.id}/projects/{page.params.pid}"
					color="alternative"
					class="flex items-center gap-2 rounded-sm border border-gray-300 p-1"
				>
					<svg
						class="h-6 w-6 text-gray-800 dark:text-white"
						aria-hidden="true"
						xmlns="http://www.w3.org/2000/svg"
						width="24"
						height="24"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke="currentColor"
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M5 12h14M5 12l4-4m-4 4 4 4"
						/>
					</svg>
				</Button>
				<span>評分分析</span>
			</div>
		</h1>

		<div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-2">
			<Card>
				<h2 class="mb-1 text-sm font-medium text-gray-500">總提交數</h2>
				<p class="text-3xl font-bold">{analytics.totalSubmissions}</p>
			</Card>
			<Card>
				<h2 class="mb-1 text-sm font-medium text-gray-500">總評分數</h2>
				<p class="text-3xl font-bold">{analytics.totalJudgments}</p>
			</Card>
		</div>

		<div class="mb-6">
			<h2 class="mb-3 text-lg font-bold">評分狀態分佈</h2>
			{#if analytics.statusCounts.length === 0}
				<p class="text-gray-500">無資料</p>
			{:else}
				<div class="flex flex-wrap gap-3">
					{#each analytics.statusCounts as item}
						<div class="flex items-center gap-2">
							<Badge color={statusColorMap[item.status] ?? 'dark'}>{item.status}</Badge>
							<span class="font-medium">{item.count}</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<Card>
				<h2 class="mb-1 text-sm font-medium text-gray-500">平均評分時間</h2>
				<p class="text-2xl font-bold">
					{analytics.avgJudgingDurationMs !== null ? `${analytics.avgJudgingDurationMs} ms` : '—'}
				</p>
			</Card>
			<Card>
				<h2 class="mb-1 text-sm font-medium text-gray-500">平均排隊等待時間</h2>
				<p class="text-2xl font-bold">
					{analytics.avgQueueWaitMs !== null ? `${analytics.avgQueueWaitMs} ms` : '—'}
				</p>
			</Card>
		</div>
	</div>
{/if}
