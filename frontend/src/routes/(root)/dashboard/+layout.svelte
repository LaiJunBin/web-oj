<script lang="ts">
	import { useContestsStorage, useContestStorage } from '@/apps/admin/domain/contest';
	import { useContestApp } from '@/apps/admin/application/contest';
	import { adminCurrentProject } from '$lib/adminCurrentProject.svelte';
	import { choose, rememberedId } from '$lib/contestSelection';
	import {
		Sidebar,
		SidebarDropdownItem,
		SidebarDropdownWrapper,
		SidebarGroup,
		SidebarItem,
		SidebarWrapper
	} from 'flowbite-svelte';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { BookSolid, CogSolid, FlagSolid, StarSolid } from 'flowbite-svelte-icons';

	const { children } = $props();

	const contestApp = useContestApp();
	const contestsStorage = useContestsStorage();
	const contestStorage = useContestStorage();

	// Ticks every second: drives both the "currently open" filter below and the
	// phase countdown in the content area.
	let now = $state(Date.now());

	// The picker is a shortcut to what is being run right now, not a browser of
	// every contest ever held (that's 競賽管理) — so it only offers contests and
	// projects inside their activate window.
	const within = (activate: string, deactivate: string) =>
		new Date(activate).getTime() <= now && now <= new Date(deactivate).getTime();

	const contestList = $derived(contestsStorage.get());
	const ongoingContests = $derived(
		(contestList ?? []).filter((c) => within(c.activateTime, c.deactivateTime))
	);
	const contestDetail = $derived(contestStorage.get());
	// Only offer the selected contest's own projects — the admin contest detail
	// store is shared with the management pages, which may have left a different
	// contest in it.
	const projectItems = $derived(
		contestDetail && contestDetail.id === adminCurrentProject.contestId
			? contestDetail.projects.filter((p) => within(p.activateTime, p.deactivateTime))
			: []
	);

	onMount(() => {
		contestApp.withLoading().fetchContests();
		const timer = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(timer);
	});

	// Load the selected contest's detail so its projects become selectable.
	$effect(() => {
		const contestId = adminCurrentProject.contestId;
		if (contestId == null) return;
		contestApp.fetchContest(contestId);
	});

	// Seed the picker's default once, following the navbar's current contest so an
	// admin who walked into the dashboard from a contest lands on that one instead
	// of an arbitrary first. One-way read of the contestant-side remembered ids
	// (contestSelection) — the dashboard never writes back to them. Runs only while
	// nothing is selected, so the admin's own remembered selection (restored in
	// adminCurrentProject's constructor) and any explicit pick win. choose() gives
	// exactly the wanted order: the navbar contest when it is in the admin's ongoing
	// list, else the first ongoing one.
	let seededContest = false;
	$effect(() => {
		if (seededContest || adminCurrentProject.contestId !== null || contestList === null) return;
		seededContest = true;
		const { id } = choose(
			rememberedId('currentContestId'),
			ongoingContests.map((c) => c.id)
		);
		const c = ongoingContests.find((x) => x.id === id);
		if (c) adminCurrentProject.setContest(c.id, c.name);
	});

	// Within the navbar's contest, follow its remembered project too — but only when
	// that project is actually offered here. If it is not (a different contest, or a
	// project not in the activate window), leave it to the single-project auto-select
	// or the admin's pick; never force an arbitrary first project on a multi-project
	// contest. Same one-way read.
	let seededProject = false;
	$effect(() => {
		if (seededProject || adminCurrentProject.projectId !== null) return;
		if (!contestDetail || contestDetail.id !== adminCurrentProject.contestId) return;
		if (adminCurrentProject.contestId !== rememberedId('currentContestId')) return;
		const remembered = rememberedId('currentProjectId');
		if (remembered === null) return;
		seededProject = true;
		const p = projectItems.find((x) => x.id === remembered);
		if (p) adminCurrentProject.setProject(p.id, p.name, p.startTime, p.endTime);
	});

	// Drop a selection that is no longer in its activate window (e.g. yesterday's
	// exam persisted in sessionStorage, or the window closing while the page is
	// open) so the dashboard never silently operates on a closed project.
	$effect(() => {
		if (contestList === null) return;
		const cid = adminCurrentProject.contestId;
		if (cid == null || ongoingContests.some((c) => c.id === cid)) return;
		adminCurrentProject.setContest(null, null);
		adminCurrentProject.setProject(null, null);
	});
	$effect(() => {
		if (!contestDetail || contestDetail.id !== adminCurrentProject.contestId) return;
		const pid = adminCurrentProject.projectId;
		if (pid == null || projectItems.some((p) => p.id === pid)) return;
		adminCurrentProject.setProject(null, null);
	});

	$effect(() => {
		if (adminCurrentProject.projectId != null) return;
		if (!contestDetail || contestDetail.id !== adminCurrentProject.contestId) return;
		if (projectItems.length !== 1) return;
		const p = projectItems[0];
		adminCurrentProject.setProject(p.id, p.name, p.startTime, p.endTime);
	});

	// Backfill the contest window onto a selection that lacks it (seeded from the
	// contestant side, or persisted before the times were stored).
	$effect(() => {
		if (adminCurrentProject.projectId == null || adminCurrentProject.projectStartTime) return;
		const p = projectItems.find((p) => p.id === adminCurrentProject.projectId);
		if (p) adminCurrentProject.setProject(p.id, p.name, p.startTime, p.endTime);
	});

	function onContestChange(value: string) {
		const id = value ? Number(value) : null;
		const name = ongoingContests.find((c) => c.id === id)?.name ?? null;
		adminCurrentProject.setContest(id, name);
		adminCurrentProject.setProject(null, null);
	}

	function onProjectChange(value: string) {
		const id = value ? Number(value) : null;
		const p = projectItems.find((p) => p.id === id);
		adminCurrentProject.setProject(id, p?.name ?? null, p?.startTime ?? null, p?.endTime ?? null);
	}

	// Phase of the selected project relative to its start/end (the activate
	// window is already guaranteed by the filter above).
	const phase = $derived.by(() => {
		if (adminCurrentProject.projectId == null) return null;
		const start = adminCurrentProject.projectStartTime;
		const end = adminCurrentProject.projectEndTime;
		if (!start || !end) return null;
		const s = new Date(start).getTime();
		const e = new Date(end).getTime();
		if (now < s) return { kind: 'before' as const, ms: s - now };
		if (now <= e) return { kind: 'running' as const, ms: e - now };
		return { kind: 'ended' as const, ms: 0 };
	});

	function formatDuration(ms: number) {
		const total = Math.max(0, Math.floor(ms / 1000));
		const d = Math.floor(total / 86400);
		const h = Math.floor((total % 86400) / 3600)
			.toString()
			.padStart(2, '0');
		const m = Math.floor((total % 3600) / 60)
			.toString()
			.padStart(2, '0');
		const s = (total % 60).toString().padStart(2, '0');
		return d > 0 ? `${d} 天 ${h}:${m}:${s}` : `${h}:${m}:${s}`;
	}
</script>

<section class="flex h-0 flex-grow overflow-y-auto">
	<Sidebar
		activeUrl={page.url.pathname}
		class="sticky top-0 flex min-w-64 flex-col border-r-2 border-gray-100 dark:border-gray-700"
	>
		<SidebarWrapper class="flex flex-grow flex-col">
			<SidebarGroup class="flex flex-grow flex-col">
				<SidebarItem label="系統狀態" href="/dashboard">
					<svelte:fragment slot="icon">
						<StarSolid
							class="h-6 w-6 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					</svelte:fragment>
				</SidebarItem>
				<SidebarDropdownWrapper isOpen label="賽前設定" class="cursor-pointer">
					<svelte:fragment slot="icon">
						<CogSolid
							class="h-6 w-6 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					</svelte:fragment>
					<SidebarDropdownItem
						label="使用者管理"
						href="/dashboard/users"
						active={page.url.pathname.startsWith('/dashboard/users')}
					/>
					<SidebarDropdownItem
						label="群組管理"
						href="/dashboard/groups"
						active={page.url.pathname.startsWith('/dashboard/groups')}
					/>
					<SidebarDropdownItem
						label="評分環境管理"
						href="/dashboard/images"
						active={page.url.pathname.startsWith('/dashboard/images')}
					/>
					<SidebarDropdownItem
						label="專案管理"
						href="/dashboard/projects"
						active={page.url.pathname.startsWith('/dashboard/projects')}
					/>
					<SidebarDropdownItem
						label="競賽管理"
						href="/dashboard/contests"
						active={page.url.pathname.startsWith('/dashboard/contests')}
					/>
					<SidebarDropdownItem
						label="評分工作者管理"
						href="/dashboard/judge-workers"
						active={page.url.pathname.startsWith('/dashboard/judge-workers')}
					/>
					<SidebarDropdownItem
						label="匯入完整競賽"
						href="/dashboard/bundles/import"
						active={page.url.pathname.startsWith('/dashboard/bundles')}
					/>
				</SidebarDropdownWrapper>
				<SidebarDropdownWrapper isOpen label="競賽過程" class="cursor-pointer">
					<svelte:fragment slot="icon">
						<FlagSolid
							class="h-6 w-6 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					</svelte:fragment>
					<!-- 這組頁面都以「目前開放中的競賽專案」為對象;選單只列 activate
					     期間內的比賽/專案(歷史比賽請走競賽管理)。 -->
					<div class="mb-2 ps-9 pe-2">
						<select
							class="mb-2 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2 text-sm text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							value={adminCurrentProject.contestId ?? ''}
							onchange={(e) => onContestChange(e.currentTarget.value)}
						>
							<option value="" disabled>選擇比賽</option>
							{#each ongoingContests as c}
								<option value={c.id}>{c.name}</option>
							{:else}
								<option value="" disabled>目前沒有進行中的比賽</option>
							{/each}
						</select>
						<select
							class="block w-full rounded-lg border border-gray-300 bg-gray-50 p-2 text-sm text-gray-900 disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							value={adminCurrentProject.projectId ?? ''}
							disabled={adminCurrentProject.contestId == null || projectItems.length === 0}
							onchange={(e) => onProjectChange(e.currentTarget.value)}
						>
							<option value="" disabled>選擇專案</option>
							{#each projectItems as p}
								<option value={p.id}>{p.name}</option>
							{/each}
						</select>
					</div>
					<SidebarDropdownItem
						label="公告與選手提問"
						href="/dashboard/clarifications"
						active={page.url.pathname.startsWith('/dashboard/clarifications')}
					/>
					<SidebarDropdownItem
						label="提交紀錄"
						href="/dashboard/submissions"
						active={page.url.pathname.startsWith('/dashboard/submissions')}
					/>
					<SidebarDropdownItem
						label="重評紀錄"
						href="/dashboard/rejudgings"
						active={page.url.pathname.startsWith('/dashboard/rejudgings')}
					/>
					<SidebarDropdownItem
						label="分數統計"
						href="/dashboard/user-scores"
						active={page.url.pathname.startsWith('/dashboard/user-scores')}
					/>
					<SidebarDropdownItem
						label="記分板展示"
						href="/dashboard/scoreboard-display"
						active={page.url.pathname.startsWith('/dashboard/scoreboard-display')}
					/>
				</SidebarDropdownWrapper>
				<SidebarItem label="參考文件" href="/dashboard/reference">
					<svelte:fragment slot="icon">
						<BookSolid
							class="h-6 w-6 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					</svelte:fragment>
				</SidebarItem>
			</SidebarGroup>
		</SidebarWrapper>
	</Sidebar>

	<section class="flex w-0 flex-grow flex-col">
		{#if phase}
			<!-- 選定專案的階段列:讓管理員隨時知道現在是賽前/進行中/已結束 -->
			<div
				class={[
					'shrink-0 px-4 py-2 text-center text-sm font-medium',
					phase.kind === 'before' && 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
					phase.kind === 'running' && 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
					phase.kind === 'ended' && 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
				]}
			>
				{adminCurrentProject.projectName} —
				{#if phase.kind === 'before'}
					尚未開始，距開始 {formatDuration(phase.ms)}
				{:else if phase.kind === 'running'}
					競賽進行中，剩餘 {formatDuration(phase.ms)}
				{:else}
					競賽已結束
				{/if}
			</div>
		{/if}
		<div class="min-h-0 flex-1 overflow-auto">
			{@render children()}
		</div>
	</section>
</section>
