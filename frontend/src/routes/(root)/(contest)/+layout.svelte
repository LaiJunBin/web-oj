<script lang="ts">
	import {
		Sidebar,
		SidebarGroup,
		SidebarItem,
		SidebarWrapper,
		SidebarCta,
		Tooltip
	} from 'flowbite-svelte';
	import {
		StarSolid,
		MailBoxSolid,
		FlagSolid,
		ArchiveSolid,
		EditOutline,
		UserCircleSolid,
		RocketSolid,
		ClockSolid
	} from 'flowbite-svelte-icons';
	import { page } from '$app/state';
	import { datetimeString } from '$lib';
	import { useUserStorage } from '@/apps/contest/domain/user';
	import { useProjectStorage } from '@/apps/contest/domain/project';
	import { useContestsStorage, useContestStorage } from '@/apps/contest/domain/contest';
	import StateMessage from '$lib/components/StateMessage/StateMessage.svelte';

	let spanClass = 'flex-1 ms-3 whitespace-nowrap';

	const userStorage = useUserStorage();
	const projectStorage = useProjectStorage();
	const contestsStorage = useContestsStorage();
	const contestStorage = useContestStorage();

	const user = $derived(userStorage.get());
	const project = $derived(projectStorage.get());
	const contests = $derived(contestsStorage.get());
	const contest = $derived(contestStorage.get());

	// Kiosk (venue display) mode hides the sidebar so the page fills the screen.
	const kiosk = $derived(page.url.searchParams.has('kiosk'));

	// Every page under this layout needs a project. On a fresh deployment whose
	// database was never seeded (no contests / no projects), or when the backend
	// is unreachable, the bootstrap chain in the root layout never produces one —
	// so distinguish "still loading" from "there is genuinely nothing" / "a fetch
	// failed", and show a real message instead of an infinite "Loading...".
	const bootstrapState: 'ready' | 'loading' | 'empty' | 'error' = $derived.by(() => {
		if (project) return 'ready';
		if (contestsStorage.error || contestStorage.error || projectStorage.error) return 'error';
		if (contests === null) return 'loading';
		if (contests.length === 0) return 'empty';
		if (contest && (contest.projects?.length ?? 0) === 0) return 'empty';
		return 'loading';
	});

	const { children } = $props();
</script>

<section class="flex h-0 flex-grow overflow-y-auto">
	{#if bootstrapState === 'empty' || bootstrapState === 'error'}
		<StateMessage
			{kiosk}
			tone={bootstrapState}
			title={bootstrapState === 'error' ? '無法載入比賽資料' : '目前沒有可參加的比賽'}
			description={bootstrapState === 'error'
				? '請稍後再試，或洽詢主辦單位。'
				: '請稍後再查看，或洽詢主辦單位。'}
		/>
	{:else}
	{#if !kiosk}
		<Sidebar
			activeUrl={page.url.pathname}
			class="sticky top-0 flex min-w-64 flex-col border-r-2 border-gray-100 dark:border-gray-700"
		>
		<SidebarWrapper class="flex flex-grow flex-col">
			<SidebarGroup class="flex flex-grow flex-col">
				<SidebarItem label="專案概述" href="/overview">
					<svelte:fragment slot="icon">
						<StarSolid
							class="h-6 w-6 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					</svelte:fragment>
				</SidebarItem>
				<SidebarItem label="評分項目" href="/grading-items">
					<svelte:fragment slot="icon">
						<FlagSolid
							class="h-6 w-6 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					</svelte:fragment>
				</SidebarItem>
				<SidebarItem label="提問與回應" {spanClass} href="/clarifications">
					<svelte:fragment slot="icon">
						<MailBoxSolid
							class="h-6 w-6 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					</svelte:fragment>
				</SidebarItem>
				<SidebarItem label="提交紀錄" href="/submissions">
					<svelte:fragment slot="icon">
						<ArchiveSolid
							class="h-6 w-6 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					</svelte:fragment>
				</SidebarItem>
				<SidebarItem label="記分板" href="/scoreboard">
					<svelte:fragment slot="icon">
						<EditOutline
							class="h-6 w-6 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
					</svelte:fragment>
				</SidebarItem>
				<SidebarCta divClass="" spanClass="" class="mt-auto bg-gray-100 dark:bg-gray-700">
					<div class="grid gap-x-3 gap-y-4" style="grid-template-columns: max-content 1fr">
						<UserCircleSolid
							class="h-6 w-6 flex-shrink-0 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
						<div class="text-gray-900 dark:text-gray-200">
							{user?.username || '尚未登入'}
						</div>
						{#if user !== null}
							<Tooltip>
								{user?.name || '尚未登入'}
							</Tooltip>
						{/if}
						<RocketSolid
							class="h-6 w-6 flex-shrink-0 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
						<div class="text-gray-900 dark:text-gray-200">
							(
							{project?.score.current ?? 0}/{project?.score.total ?? 0}
							分 )
						</div>
						{#if project !== null}
							<Tooltip>(分數/總分)</Tooltip>
						{/if}
						<ClockSolid
							class="h-6 w-6 flex-shrink-0 text-gray-500 transition duration-75 group-hover:text-gray-900 dark:text-gray-400 dark:group-hover:text-white"
						/>
						<div class="text-gray-900 dark:text-gray-200">
							{datetimeString(project?.submissions?.[0]?.createdAt) || '尚未提交'}
						</div>
						{#if project !== null}
							<Tooltip>最後提交時間</Tooltip>
						{/if}
					</div>
				</SidebarCta>
			</SidebarGroup>
		</SidebarWrapper>
		</Sidebar>
	{/if}

	<section class="w-0 flex-grow overflow-auto">
		{@render children()}
	</section>
	{/if}
</section>
