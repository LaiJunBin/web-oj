<script lang="ts">
	import Swal from 'sweetalert2';
	import { useContestsStorage, useContestStorage } from '@/apps/contest/domain/contest';
	import Navbar from '$lib/components/Navbar/Navbar.svelte';
	import { useUserProfile } from '@/apps/contest/application/user.js';
	import { useUserStorage } from '@/apps/contest/domain/user.js';
	import { useSettingStorage } from '@/apps/contest/domain/setting.js';
	import { onMount } from 'svelte';
	import { useProjectApp } from '@/apps/contest/application/project.js';
	import { useContestApp } from '@/apps/contest/application/contest.js';
	import { useProjectStorage } from '@/apps/contest/domain/project.js';
	import { Time, useTimeStorage } from '@/apps/contest/domain/time.js';
	import { streamPublicContest, streamUserContest } from '$lib/api/stream.js';
	import { createManagedStream, type StreamFailure } from '$lib/api/managed-stream.js';
	import { useScoreboardStorage } from '@/apps/contest/domain/scoreboard.js';
	import { useScoreboardApp } from '@/apps/contest/application/scoreboard.js';
	import { afterLoad, choose, forget, rememberedId } from '$lib/contestSelection.js';
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import StateMessage from '$lib/components/StateMessage/StateMessage.svelte';

	let { children } = $props();

	// Kiosk (venue display) mode hides the top navbar and removes the width
	// constraint so a page like the scoreboard can fill a big screen.
	const kiosk = $derived(page.url.searchParams.has('kiosk'));

	const userProfile = useUserProfile();
	const userStorage = useUserStorage();
	const settingStorage = useSettingStorage();

	// Public display surfaces (the login page itself, the scoreboard, and anything
	// opened in kiosk mode) must never be bounced to /login on a 401 — a big screen
	// at the venue front should keep showing the board, not a login form. Read the
	// live location so this stays correct regardless of when the 401 arrives.
	function isPublicDisplay(): boolean {
		const { pathname, search } = window.location;
		return (
			pathname === '/login' ||
			pathname === '/scoreboard' ||
			new URLSearchParams(search).has('kiosk')
		);
	}

	const contestApp = useContestApp();
	const contestsStorage = useContestsStorage();
	const contestStorage = useContestStorage();
	const projectApp = useProjectApp();
	const projectStorage = useProjectStorage();
	const timeStorage = useTimeStorage();
	const scoreboardStorage = useScoreboardStorage();
	const scoreboardApp = useScoreboardApp();

	// Pinned kiosk:?kiosk&contestId=&projectId= 把大螢幕鎖定在特定場次
	// (由後台「記分板展示」產生連結)。存在任一 id 參數即進入 pin 模式,
	// 完全跳過「sessionStorage / 第一場」的推斷。
	const kioskPin = $derived.by(() => {
		if (!kiosk) return null;
		const rawContest = page.url.searchParams.get('contestId');
		const rawProject = page.url.searchParams.get('projectId');
		if (rawContest === null && rawProject === null) return null;
		const contestId = Number(rawContest);
		const projectId = Number(rawProject);
		const valid =
			rawContest !== null &&
			rawProject !== null &&
			Number.isInteger(contestId) &&
			contestId > 0 &&
			Number.isInteger(projectId) &&
			projectId > 0;
		return { contestId, projectId, valid };
	});

	// Pin 模式的失敗「必須大聲暴露」:指定的場次不存在時直接顯示錯誤,
	// 絕不 fallback 到別場——否則場地人員會以為畫面上的就是指定的那場。
	const kioskPinError = $derived.by(() => {
		if (!kioskPin) return null;
		if (!kioskPin.valid) {
			return '網址參數不完整或格式錯誤:需要同時提供 contestId 與 projectId。';
		}
		const contests = contestsStorage.get();
		if (contests && !contests.some((c) => c.id === kioskPin.contestId)) {
			return '找不到網址指定的競賽:可能不存在、未公開,或已下線。';
		}
		const contest = contestStorage.get();
		if (
			contest &&
			contest.id === kioskPin.contestId &&
			!contest.projects?.some((p) => p.id === kioskPin.projectId)
		) {
			return '此競賽中找不到網址指定的專案。';
		}
		if (contestStorage.error || projectStorage.error) {
			return '無法載入指定的競賽專案,請稍後再試。';
		}
		return null;
	});

	onMount(() => {
		contestApp.withLoading().fetchContests();
	});

	// React once to a dead session (the api layer dispatches this on a terminal
	// 401 — see src/api/index.ts). A burst of parallel calls can fire several
	// events at once, so guard against showing multiple toasts / redirects.
	onMount(() => {
		let handling = false;
		const onUnauthorized = () => {
			if (handling || isPublicDisplay()) {
				return;
			}
			handling = true;
			userStorage.clear();
			Swal.fire({
				toast: true,
				position: 'top-end',
				icon: 'warning',
				title: '登入已過期，請重新登入',
				showConfirmButton: false,
				timer: 3000
			});
			goto('/login').finally(() => {
				handling = false;
			});
		};
		window.addEventListener('auth:unauthorized', onUnauthorized);
		return () => window.removeEventListener('auth:unauthorized', onUnauthorized);
	});

	// What to tell the viewer when the live connection is not delivering. Null
	// while healthy. A venue big screen that quietly freezes on its last frame is
	// worse than one that says it has stopped, so this is shown on kiosk too.
	let streamNotice = $state<string | null>(null);

	const UNAUTHORIZED_NOTICE = '即時連線已中斷(未授權),畫面停在最後一次更新';

	/**
	 * A terminal failure is usually not a fault at all — 404 is how a contest
	 * stream reports its own window closing, or a public board reports being
	 * un-published. Promising a reconnection there would be a lie, and a scary
	 * one on a venue screen; but 401/403 is a refused session, which is a
	 * different thing again and must not be described as the contest ending.
	 */
	function noticeFor(failure: StreamFailure): string {
		if (!failure.terminal) return '即時連線異常,畫面可能不是最新的 —— 持續重新連線中';
		return failure.status === 404
			? '這場比賽已結束或已下架,畫面停在最後一次更新'
			: UNAUTHORIZED_NOTICE;
	}

	// Extract only the primitive fields that should trigger SSE reconnection.
	// Reading projectStorage.get() directly in $effect re-runs on every project object
	// update (e.g. each fetchProject call), creating duplicate EventSource connections.
	// Using $derived on primitives mirrors the pattern used in submissions/+page.svelte.
	const project = $derived(projectStorage.get());
	const projectId = $derived(project?.id);
	const timeLeft = $derived(project?.timeLeft);

	// Effect 1：依據當前 project 及登入狀態建立 SSE 連線
	// userProfile.isContestant 是故意的 reactive dependency：
	// 登入/登出時需切換 public ↔ user 兩條不同的後端路徑
	$effect(() => {
		if (!projectId || timeLeft === 0) {
			timeStorage.clear();
			// Clear the notice here too: without this, a contest that ends while a
			// notice is showing leaves it on screen for good — there is no stream
			// left to recover and take it down.
			streamNotice = null;
			return;
		}

		// Kiosk is a public venue display: use the public stream even when a
		// contestant happens to be logged in on that machine. It carries the same
		// scoreboard data without the user stream's clarification traffic (and its
		// notifications / invalidateAll), and avoids tying the big screen to — or
		// exposing — a contestant's session.
		const asContestant = userProfile.isContestant && !kiosk;
		const streamProjectId = projectId;

		const listeners: Record<string, (event: MessageEvent) => void> = {
			time: (event) => {
				try {
					const json = JSON.parse(event.data);
					timeStorage.set(new Time(json.time_type, json.time_left));
				} catch (e) {
					console.error('Failed to parse time event', e);
				}
			},
			scoreboard: (event) => {
				try {
					const json = JSON.parse(event.data);
					scoreboardStorage.set(json);
				} catch (e) {
					console.error('Failed to parse scoreboard event', e);
				}
			}
		};
		if (asContestant) {
			listeners.clarification = (event) => {
				try {
					const json = JSON.parse(event.data);
					if (
						typeof Notification !== 'undefined' &&
						Notification.permission === 'granted' &&
						settingStorage.get()?.enableNotification
					) {
						new Notification(json.subject, { body: json.content });
					}
					invalidateAll();
				} catch (e) {
					console.error('Failed to parse clarification event', e);
				}
			};
		}

		// A fresh connection starts with a clean slate — otherwise a notice from
		// the previous project would linger over a board that is now fine.
		streamNotice = null;

		// Supervised, because this connection multiplexes two or three backend
		// streams: if one of them fails the proxy can only report it and close,
		// and left to itself the browser would reconnect once a second forever
		// while the board sat frozen with nothing on screen to say why.
		return createManagedStream({
			connect: () =>
				asContestant ? streamUserContest(streamProjectId) : streamPublicContest(streamProjectId),
			on: listeners,
			// The scoreboard stream pushes only when a score changes, so a reconnect
			// leaves the board on whatever it was showing. Pull it back over REST —
			// deliberately without withLoading(), which would throw the full-screen
			// spinner over a venue display every time the network hiccuped, and via
			// the resync variant, whose failure leaves the board alone rather than
			// clearing it to Loading... while the backend is down.
			onResync: () => scoreboardApp.resyncScoreboard(streamProjectId),
			onDegraded: (failure) => (streamNotice = noticeFor(failure)),
			onRecovered: () => (streamNotice = null),
			onError: (event, stop) => {
				const status = (event as MessageEvent & { status?: number }).status;
				if (status === 401 || status === 403) {
					stop();
					// Same exemption as XHR 401s: don't bounce a public display page.
					// But it must not just freeze either — without a word on screen,
					// a big screen at the venue front goes on showing a dead board.
					if (isPublicDisplay()) {
						streamNotice = UNAUTHORIZED_NOTICE;
					} else {
						goto('/login');
					}
				}
			}
		});
	});

	// Pin 模式 Effect A/B:只在指定的 id 確認存在後才 fetch;不存在時
	// 什麼都不做(kioskPinError 會顯示錯誤畫面),不清 sessionStorage、不 fallback。
	$effect(() => {
		if (!kioskPin?.valid) return;
		const contests = contestsStorage.get();
		if (!contests) return;
		if (!contests.some((c) => c.id === kioskPin.contestId)) return;
		contestApp.withLoading().fetchContest(kioskPin.contestId);
	});
	$effect(() => {
		if (!kioskPin?.valid) return;
		const contest = contestStorage.get();
		if (!contest || contest.id !== kioskPin.contestId) return;
		if (!contest.projects?.some((p) => p.id === kioskPin.projectId)) return;
		projectApp.withLoading().fetchProject(kioskPin.projectId);
	});

	// 記住的選擇已經不在清單上時會切到預設項目——但要說一聲。靜默換到另一場
	// 比賽,使用者只會看到標題變了、一頭霧水。樣式比照登入過期的 toast。
	function announceSwitch(what: '比賽' | '專案') {
		Swal.fire({
			toast: true,
			position: 'top-end',
			icon: 'info',
			title: `先前選擇的${what}已不存在或無法存取,已為你切換`,
			showConfirmButton: false,
			timer: 4000
		});
	}

	// Effect 2:依據 contests 列表決定並 fetch 當前 contest。
	// 記住的 id 先對清單:清單是剛抓的、記憶不是——不在清單上就直接丟掉,不必先打
	// 一次註定失敗的請求。清單說在、明細卻 404(中間被刪了),才走 afterLoad 的
	// 「忘掉、重抓清單」;5xx 則保留記憶,讓錯誤畫面的「請稍後再試」是真話。
	// 這裡不同步讀 contestStorage.error,失敗不會再觸發自己,也就沒有迴圈。
	$effect(() => {
		if (kioskPin) return; // pin 模式(含參數無效)一律不做預設推斷
		const contests = contestsStorage.get();
		if (!contests) return;

		const choice = choose(
			rememberedId('currentContestId'),
			contests.map((c) => c.id)
		);
		if (choice.forgotten) {
			forget('currentContestId');
			if (choice.id !== undefined) announceSwitch('比賽');
		}
		if (choice.id === undefined) return;

		const { id, remembered } = choice;
		contestApp
			.withLoading()
			.fetchContest(id)
			.then(() => {
				if (afterLoad(contestStorage.error, remembered) === 'forget') {
					forget('currentContestId');
					contestApp.fetchContests();
				}
			});
	});

	// Effect 3:依據當前 contest 決定並 fetch 當前 project,規則同上。
	// contest 明細本身就帶題目清單,所以同樣先對清單再抓。
	$effect(() => {
		if (kioskPin) return;
		const contest = contestStorage.get();
		if (!contest) return;

		const choice = choose(
			rememberedId('currentProjectId'),
			(contest.projects ?? []).map((p) => p.id)
		);
		if (choice.forgotten) {
			forget('currentProjectId');
			if (choice.id !== undefined) announceSwitch('專案');
		}
		if (choice.id === undefined) return;

		const { id, remembered } = choice;
		projectApp
			.withLoading()
			.fetchProject(id)
			.then(() => {
				if (afterLoad(projectStorage.error, remembered) === 'forget') {
					forget('currentProjectId');
					contestApp.fetchContest(contest.id);
				}
			});
	});
</script>

<div class="flex min-h-screen flex-col dark:bg-gray-800">
	<div class={['flex flex-grow flex-col', !kiosk && 'container mx-auto']}>
		{#if !kiosk}
			<Navbar />
		{/if}
		{#if kioskPinError}
			<StateMessage
				kiosk
				tone="error"
				title="無法顯示指定的記分板"
				description={`${kioskPinError} 請回後台「記分板展示」重新產生連結。`}
			/>
		{:else}
			{@render children()}
		{/if}
	</div>
	{#if streamNotice}
		<!-- Deliberately shown on kiosk too: a venue big screen is exactly where a
		     silently stale scoreboard does the most damage. -->
		<div
			role="status"
			class={[
				'fixed right-4 z-50 flex items-center gap-2 rounded-lg bg-amber-100 px-4 py-2 font-medium text-amber-900 shadow-lg ring-1 ring-amber-300 dark:bg-amber-900 dark:text-amber-100 dark:ring-amber-700',
				// Kiosk keeps a full-width footer band at the bottom, whose text the
				// organiser writes and can make long. Sit above it rather than on it.
				// Kiosk pins a full-width footer band at the bottom whose text the
				// organiser writes; it wraps, so no fixed bottom offset can clear it.
				// The top-right corner is free by construction — the kiosk header's
				// first line is the literal word 記分板, centred.
				kiosk ? 'top-4 text-xl' : 'bottom-4 text-sm'
			]}
		>
			<span class="relative flex h-2.5 w-2.5">
				<span
					class="absolute inline-flex h-full w-full animate-ping rounded-full bg-amber-500 opacity-75"
				></span>
				<span class="relative inline-flex h-2.5 w-2.5 rounded-full bg-amber-600"></span>
			</span>
			{streamNotice}
		</div>
	{/if}
</div>
