<script lang="ts">
	import { page } from '$app/state';
	import {
		useRejudgeComparisonStorage,
		type GradingItemOutcome,
		type TestcaseOutcome
	} from '@/apps/admin/domain/rejudge';
	import { Accordion, AccordionItem, Alert, Badge, Button, Tooltip } from 'flowbite-svelte';
	import { ArrowRightOutline, InfoCircleSolid } from 'flowbite-svelte-icons';

	const comparison = $derived(useRejudgeComparisonStorage().get()!);
	const rejudgeId = page.params.id;

	const statusColorMap: Record<string, string> = {
		Pending: 'dark',
		Running: 'blue',
		Passed: 'green',
		Failed: 'red',
		Error: 'red',
		Timeout: 'indigo',
		Cancelled: 'yellow',
		Created: 'dark',
		Queued: 'dark',
		Judging: 'blue',
		Judged: 'green',
		Rejudged: 'purple',
		NotSubmitted: 'dark'
	};
	const statusColor = (s?: string) => (statusColorMap[s ?? ''] ?? 'dark') as never;

	// Pair before/after by id (never by index — a side may be missing when the
	// grading setup changed, which is a normal reason for a rejudge).
	function pairById<T extends { id: number }>(before: T[] | undefined, after: T[] | undefined) {
		const b = new Map((before ?? []).map((x) => [x.id, x]));
		const a = new Map((after ?? []).map((x) => [x.id, x]));
		const ids = [...new Set([...(after ?? []).map((x) => x.id), ...(before ?? []).map((x) => x.id)])];
		return ids.map((id) => ({ id, before: b.get(id), after: a.get(id) }));
	}

	type Change = 'added' | 'removed' | 'changed' | 'unchanged' | 'pending';
	const changeTag: Record<Change, { label: string; color: string }> = {
		added: { label: '新增', color: 'green' },
		removed: { label: '移除', color: 'red' },
		changed: { label: '有變動', color: 'yellow' },
		unchanged: { label: '無變動', color: 'dark' },
		pending: { label: '評分中', color: 'blue' }
	};
	// The rejudge judgment may still be running; while an item/testcase is not yet
	// in a final state we must not claim a definitive change.
	const inProgressStatus = (s?: string) =>
		['Created', 'Queued', 'Judging', 'Pending', 'Running'].includes(s ?? '');
	const outcomeChanged = (
		before: { status: string; score: number } | undefined,
		after: { status: string; score: number } | undefined
	): Change => {
		if (!before) return 'added';
		if (!after) return 'removed';
		if (inProgressStatus(after.status)) return 'pending';
		return before.status !== after.status || before.score !== after.score ? 'changed' : 'unchanged';
	};

	const gradingItemPairs = $derived(
		pairById<GradingItemOutcome>(comparison?.before?.gradingItems, comparison?.after.gradingItems)
	);
</script>

{#snippet scorePair(
	before: { score: number; totalScore: number } | null | undefined,
	after: { score: number; totalScore: number } | null | undefined,
	afterPending: boolean
)}
	<span class="inline-flex items-center gap-1 whitespace-nowrap">
		{#if before}<span class="text-gray-500 dark:text-gray-400">{before.score} / {before.totalScore}</span>{/if}
		{#if before && (after || afterPending)}<ArrowRightOutline class="h-4 w-4 text-gray-400" />{/if}
		{#if afterPending}
			<span class="text-blue-500 dark:text-blue-400">評分中</span>
		{:else if after}
			<span class="font-semibold">{after.score} / {after.totalScore}</span>
		{/if}
	</span>
{/snippet}

{#snippet statusPair(before: string | null | undefined, after: string | null | undefined)}
	<span class="inline-flex items-center gap-1">
		{#if before}<Badge color={statusColor(before)}>{before}</Badge>{/if}
		{#if before && after}<ArrowRightOutline class="h-4 w-4 text-gray-400" />{/if}
		{#if after}<Badge color={statusColor(after)}>{after}</Badge>{/if}
	</span>
{/snippet}

{#if !comparison}
	<div class="p-4">Loading...</div>
{:else}
	<div class="p-4">
		<h1 class="mb-4 flex items-center gap-3 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<Button
				href={`/dashboard/rejudgings/${rejudgeId}`}
				color="alternative"
				class="cursor-pointer p-2"
			>
				←
			</Button>
			<span>評分前後比對</span>
		</h1>

		<!-- 總覽 -->
		<div class="mb-4 grid grid-cols-1 gap-4 md:grid-cols-3">
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<h3 class="text-sm text-gray-500 dark:text-gray-400">提交</h3>
				<a
					href={`/dashboard/submissions/${comparison.id}`}
					class="text-primary-600 dark:text-primary-400 hover:underline"
				>
					查看完整提交詳情
				</a>
			</div>
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<h3 class="text-sm text-gray-500 dark:text-gray-400">狀態(前 → 後)</h3>
				<div class="mt-1">
					{@render statusPair(comparison.before?.status, comparison.after.status)}
				</div>
			</div>
			<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
				<h3 class="text-sm text-gray-500 dark:text-gray-400">總分(前 → 後)</h3>
				<div class="mt-1">
					{@render scorePair(
						comparison.before,
						comparison.after,
						inProgressStatus(comparison.after.status)
					)}
				</div>
			</div>
		</div>

		{#if !comparison.before}
			<Alert color="yellow" class="mb-4 dark:bg-yellow-800 dark:text-yellow-200">
				<InfoCircleSolid slot="icon" class="h-5 w-5" />
				此提交在本次重新評分前<strong>沒有前次採計評分</strong>,以下僅顯示重新評分後的結果。
			</Alert>
		{/if}

		{#if inProgressStatus(comparison.after.status)}
			<Alert color="yellow" class="mb-4 dark:bg-yellow-800 dark:text-yellow-200">
				<InfoCircleSolid slot="icon" class="h-5 w-5" />
				重新評分<strong>尚未完成</strong>,以下標示為「評分中」的項目結果可能仍會變動。
			</Alert>
		{/if}

		<Alert color="blue" class="mb-4 dark:bg-blue-900 dark:text-blue-100">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			分數皆為「<strong>實得 / 滿分</strong>」。評分項目 / 測資以 id 對應;僅存在單側者標示為<strong>新增 / 移除</strong>。
		</Alert>

		<!-- 各評分項目 -->
		<Accordion multiple>
			{#each gradingItemPairs as pair (pair.id)}
				{@const before = pair.before}
				{@const after = pair.after}
				{@const change = outcomeChanged(before, after)}
				{@const renamed = !!before && !!after && before.label !== after.label}
				{@const title = after?.label || after?.name || before?.label || before?.name || `#${pair.id}`}
				{@const testcasePairs = pairById<TestcaseOutcome>(before?.testcases, after?.testcases)}
				<AccordionItem open={change !== 'unchanged'}>
					<span slot="header" class="flex flex-wrap items-center gap-2">
						<Badge color={changeTag[change].color as never}>{changeTag[change].label}</Badge>
						<span class="font-semibold">{title}</span>
						{#if renamed}
							<Badge color="purple">已改名</Badge>
						{/if}
						<span class="ml-2">{@render statusPair(before?.status, after?.status)}</span>
						<span class="ml-2">{@render scorePair(before, after, inProgressStatus(after?.status))}</span>
						{#if (after ?? before)?.scoringStrategy}
							<Badge color="dark">{(after ?? before)?.scoringStrategy}</Badge>
						{/if}
					</span>
					<section>
						<div class="grid grid-cols-[1fr_auto_auto] items-center gap-x-4 gap-y-2">
							<div class="text-sm font-bold text-gray-500 dark:text-gray-400">測試案例</div>
							<div class="text-sm font-bold text-gray-500 dark:text-gray-400">狀態(前 → 後)</div>
							<div class="text-sm font-bold text-gray-500 dark:text-gray-400">分數(前 → 後)</div>
							{#each testcasePairs as tc (tc.id)}
								{@const tcChange = outcomeChanged(tc.before, tc.after)}
								<div class="flex items-center gap-2">
									<Badge color={changeTag[tcChange].color as never}>{changeTag[tcChange].label}</Badge>
									<span>{tc.after?.name || tc.before?.name || `#${tc.id}`}</span>
									<InfoCircleSolid class="h-4 w-4 text-gray-400" />
									<Tooltip class="max-w-xs text-left">
										parser 原始輸出 (parsedScore):
										前 {tc.before?.parsedScore ?? '—'} → 後 {tc.after?.parsedScore ?? '—'}
									</Tooltip>
								</div>
								<div>{@render statusPair(tc.before?.status, tc.after?.status)}</div>
								<div>{@render scorePair(tc.before, tc.after, inProgressStatus(tc.after?.status))}</div>
							{:else}
								<div class="col-span-3 text-gray-500 dark:text-gray-400">無測試案例</div>
							{/each}
						</div>
					</section>
				</AccordionItem>
			{:else}
				<AccordionItem open>
					<span slot="header">無評分項目</span>
				</AccordionItem>
			{/each}
		</Accordion>
	</div>
{/if}
