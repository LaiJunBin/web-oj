<script lang="ts">
	import { flip } from 'svelte/animate';
	import { Alert, Tooltip } from 'flowbite-svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { InfoCircleSolid } from 'flowbite-svelte-icons';
	import { useProjectStorage } from '@/apps/contest/domain/project';
	import { useScoreboardApp } from '@/apps/contest/application/scoreboard';
	import { useScoreboardStorage } from '@/apps/contest/domain/scoreboard';
	import { useTimeStorage } from '@/apps/contest/domain/time';
	import StateMessage from '$lib/components/StateMessage/StateMessage.svelte';

	const scoreboardApp = useScoreboardApp();
	const scoreboardStorage = useScoreboardStorage();
	const scoreboard = $derived(scoreboardStorage.get());

	const projectStorage = useProjectStorage();
	const project = $derived(projectStorage.get());

	const timeStorage = useTimeStorage();
	const time = $derived(timeStorage.get());

	// Kiosk (venue display) mode: the layouts hide the navbar/sidebar; here we add
	// a big-screen header with the competition clock and an optional free-text line
	// (?text=...), used for a headline such as the current problem, plus an optional
	// always-visible footer notice (?footer=...) for a standing announcement.
	const kiosk = $derived(page.url.searchParams.has('kiosk'));
	const displayText = $derived(page.url.searchParams.get('text'));
	const footerText = $derived(page.url.searchParams.get('footer'));

	let isDragging = false;

	$effect(() => {
		if (!project) {
			return;
		}

		scoreboardApp.withLoading().fetchScoreboard(project.id);
	});
</script>

<svelte:body onmouseup={() => (isDragging = false)} />
{#if !scoreboard}
	<p>Loading...</p>
{:else}
	<main class={['mx-auto flex h-full flex-col p-4', !kiosk && 'max-w-6xl']}>
		{#if kiosk}
			<div class="mb-6 shrink-0 text-center">
				<h1 class="text-6xl leading-none font-extrabold tracking-tight text-gray-900 dark:text-white">
					記分板
				</h1>
				{#if displayText}
					<p class="mt-3 text-3xl font-semibold text-gray-600 dark:text-gray-300">
						{displayText}
					</p>
				{/if}
				<div
					class="mt-4 flex items-baseline justify-center gap-3 text-gray-900 dark:text-white"
				>
					<span class="text-xl text-gray-500 dark:text-gray-400">競賽時間</span>
					<span class="text-5xl font-bold tabular-nums">
						{#if time === null || time.isContestEnd()}
							已結束
						{:else if time.isInWaitContestStart()}
							{time.time} 後開放
						{:else}
							{time.time}
						{/if}
					</span>
				</div>
			</div>
		{:else}
			<Alert color="yellow" class="mb-4 shrink-0 dark:bg-yellow-800 dark:text-yellow-200">
				<InfoCircleSolid slot="icon" class="h-5 w-5" />
				僅採計完整提交的裁判結果
			</Alert>

			<h1
				class="shrink-0 text-center text-4xl leading-none font-extrabold tracking-tight text-gray-900 dark:text-white"
			>
				記分板
			</h1>
		{/if}

		{#if scoreboard.heads.length === 0}
			<StateMessage
				{kiosk}
				title="尚未設定評分項目"
				description="待主辦方設定評分項目後，記分板將顯示於此。"
			/>
		{:else if scoreboard.rows.length === 0}
			<!-- rows 是「全體已報名參賽者」（後端會為未作答者補上 NotSubmitted 的一列），
			     所以空的唯一原因是這個專案還沒有任何參賽者報名——不是「還沒有成績」。 -->
			<StateMessage
				{kiosk}
				title="尚無參賽者"
				description="此專案目前沒有已報名的參賽者，記分板會在參賽者加入後顯示。"
			/>
		{:else if scoreboard.compact}
			<!-- 精簡樣式：每位選手一條橫條，每個評分項一段漸層色塊；
			     適合評分項很多、逐項表格在大螢幕塞不下的考場 -->
			<div class="mt-4 min-h-0 flex-1 space-y-3 overflow-auto">
				{#each scoreboard.rows as row (row.username)}
					<div animate:flip class="flex items-center gap-3">
						<div
							class="w-44 shrink-0 truncate text-right text-gray-900 dark:text-white"
							title={row.name}
						>
							{row.name}
							<span class="text-sm text-gray-500 dark:text-gray-400">({row.username})</span>
						</div>
						<div class="flex h-7 flex-1 overflow-hidden rounded border dark:border-gray-600">
							{#each row.statuses as status, i}
								<div
									class={{
										'h-full flex-1': true,
										'bg-gradient-to-b from-green-200 to-green-400': status === 'Passed',
										'bg-gradient-to-b from-red-200 to-red-400': status === 'Failed',
										'bg-gray-100 dark:bg-gray-700': status !== 'Passed' && status !== 'Failed'
									}}
									title={`${scoreboard.heads[i]?.name}: ${status}`}
								/>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<!-- 只有表格所在的這層可捲動;上方標題與通知固定不動 -->
			<div class="mt-4 min-h-0 flex-1 overflow-auto">
				<section
					class="mx-auto grid w-fit cursor-grab gap-2 select-none active:cursor-grabbing"
					style={`grid-template-columns: repeat(${
						scoreboard.heads.length + 1
					}, minmax(6rem, min-content))`}
					onmousedown={(e) => {
						isDragging = true;
					}}
					onmousemove={(e) => {
						let el = e.currentTarget;
						if (isDragging) {
							while (el && el.tagName !== 'BODY') {
								el.scrollLeft -= e.movementX;
								el.scrollTop -= e.movementY;
								el = el.parentElement as HTMLElement;
							}
						}
					}}
				>
					<div />
					{#each scoreboard.heads as head, i}
						<button
							class="sticky top-0 flex cursor-pointer items-center justify-center border bg-gray-100 text-center whitespace-pre-wrap dark:bg-gray-700"
							onclick={() => {
								goto('/grading-items', { state: { id: head.id } });
							}}
						>
							{i + 1}
						</button>
						<Tooltip class="fixed">
							{head.name}
						</Tooltip>
					{/each}

					{#each scoreboard.rows as row (row.username)}
						<div
							animate:flip
							class="col-span-full grid justify-center gap-2"
							style={`grid-template-columns: repeat(${
								scoreboard.heads.length + 1
							}, minmax(6rem, min-content))`}
						>
							<div
								class="sticky left-0 flex flex-col items-center justify-center border bg-gray-100 whitespace-pre-wrap dark:bg-gray-700"
							>
								{row.name}
								<span>({row.username})</span>
							</div>
							{#each row.statuses as status}
								<div
									class={{
										'aspect-square border': true,
										'bg-green-300': status === 'Passed',
										'bg-red-300': status === 'Failed'
									}}
								/>
							{/each}
						</div>
					{/each}
				</section>
			</div>
		{/if}

		{#if footerText}
			<!-- 常駐頁尾公告：位於可捲動內容之後、以 shrink-0 固定於畫面底部，
			     選手再多也不會被捲走 -->
			<Alert
				color="yellow"
				class="mt-4 shrink-0 justify-center text-center text-2xl font-medium dark:bg-yellow-800 dark:text-yellow-200"
			>
				<InfoCircleSolid slot="icon" class="h-6 w-6" />
				{footerText}
			</Alert>
		{/if}
	</main>
{/if}
