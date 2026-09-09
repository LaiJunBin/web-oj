<script lang="ts">
	import { onMount } from 'svelte';
	import {
		JudgeWorkersEventDataSchema,
		useJudgeWorkersStorage
	} from '@/apps/admin/domain/judge-worker';
	import { useJudgeWorkerApp } from '@/apps/admin/application/judge-worker';
	import { streamJudgeWorkers } from '@/lib/api/stream';
	import { useContestsStorage } from '@/apps/admin/domain/contest';
	import { useContestProjectApp } from '@/apps/admin/application/contest-project';
	import { useJudgmentAnalyticsStorage } from '@/apps/admin/domain/contest-project';
	import {
		JudgeQueueEventDataSchema,
		JudgmentErrorsEventDataSchema,
		SubmissionRateEventDataSchema,
		useJudgeQueueStorage,
		useJudgmentErrorsStorage,
		useSubmissionRateStorage
	} from '@/apps/admin/domain/system-status';
	import { streamSystemStatus } from '@/lib/api/stream';
	import { createManagedStream, type StreamFailure } from '@/lib/api/managed-stream';
	import { adminCurrentProject } from '$lib/adminCurrentProject.svelte';
	import { datetimeString } from '@/lib';
	import { Badge, Button } from 'flowbite-svelte';

	// 系統狀態頁:比賽期間的指揮台。一眼看出評分系統是否活著、佇列是否堵塞、
	// 有沒有 Error 爆量 — 取代原本與競賽管理專案頁重複的 landing。
	const judgeWorkerApp = useJudgeWorkerApp();
	const judgeWorkersStorage = useJudgeWorkersStorage();
	const judgeWorkers = $derived(judgeWorkersStorage.get() ?? []);
	const runningWorkers = $derived(
		judgeWorkers.filter((w) => w.status === 'RUNNING' || w.status === 'IDLE')
	);
	const erroredWorkers = $derived(judgeWorkers.filter((w) => w.status === 'ERROR'));

	const contestProjectApp = useContestProjectApp();
	const analyticsStorage = useJudgmentAnalyticsStorage();
	const analytics = $derived(analyticsStorage.get());
	const errorJudgments = $derived(
		analytics?.statusCounts.find((s) => s.status === 'Error')?.count ?? 0
	);

	let now = $state(Date.now());
	const contestList = $derived(useContestsStorage().get());
	const ongoingContests = $derived(
		(contestList ?? []).filter(
			(c) =>
				new Date(c.activateTime).getTime() <= now && now <= new Date(c.deactivateTime).getTime()
		)
	);

	const judgeQueueStorage = useJudgeQueueStorage();
	const judgeQueue = $derived(judgeQueueStorage.get());
	const totalQueued = $derived(
		judgeQueue
			? judgeQueue.queueDepths.judge +
					judgeQueue.queueDepths.judgePartial +
					judgeQueue.queueDepths.rejudge
			: 0
	);
	// 「等很久」不等於「堵塞」:單筆評分可能要跑數分鐘,提交數大於 worker 數時,
	// 排隊等待數分鐘是健康的飽和。真正的異常是「有排隊、有 worker 在運作,
	// 卻沒有任何評分在執行」= 派工卡死。條件需持續 15 秒(快照每 2 秒一筆)
	// 才亮紅,以跳過上一筆結束、下一筆尚未被接手的正常空窗。
	const stallCondition = $derived(
		judgeQueue != null &&
			totalQueued > 0 &&
			judgeQueue.activeJudgments === 0 &&
			runningWorkers.length > 0
	);
	let stallSince = $state<number | null>(null);
	$effect(() => {
		if (stallCondition) {
			if (stallSince === null) stallSince = Date.now();
		} else {
			stallSince = null;
		}
	});
	const queueStalled = $derived(
		stallCondition && stallSince !== null && Date.now() - stallSince >= 15_000
	);

	const judgmentErrorsStorage = useJudgmentErrorsStorage();
	const submissionRateStorage = useSubmissionRateStorage();
	const judgmentErrors = $derived(judgmentErrorsStorage.get() ?? []);
	const submissionRate = $derived(submissionRateStorage.get() ?? []);
	const recentSubmissionTotal = $derived(submissionRate.reduce((sum, p) => sum + p.count, 0));
	const maxRateCount = $derived(Math.max(1, ...submissionRate.map((p) => p.count)));

	// 這頁的每個數字都靠 SSE 推送。連線壞掉時最危險的不是沒資料,而是舊資料
	// 看起來像現況 —— 所以連線退化要說出來。兩條連線各自記錄,否則其中一條
	// 恢復會把另一條仍在退化的事實一起抹掉。
	let workersFailure = $state<StreamFailure | null>(null);
	let statusFailure = $state<StreamFailure | null>(null);
	const streamFailure = $derived(workersFailure ?? statusFailure);

	onMount(() => {
		// workers stream 是事件驅動(只在變化時推),需要一次初始載入;
		// system-status 三條是 ticker snapshot 且連線當下立即首推,不需初始 GET。
		judgeWorkerApp.fetchJudgeWorkers();
		const timer = setInterval(() => (now = Date.now()), 30_000);

		const parsed = <T,>(
			eventName: string,
			schema: { safeParse: (v: unknown) => { success: boolean; data?: { data: T }; error?: unknown } },
			apply: (data: T) => void
		) => {
			return (event: MessageEvent) => {
				try {
					const result = schema.safeParse(JSON.parse(event.data));
					if (!result.success || !result.data) {
						console.error(result.error);
						return;
					}
					apply(result.data.data);
				} catch (e) {
					console.error(`Failed to parse ${eventName} event`, e);
				}
			};
		};

		// 兩條連線都交給 createManagedStream 監督:上游失敗時代理只能先關掉連線,
		// 沒有人接手的話瀏覽器會每秒重連一次,而指揮台上不會有任何跡象。
		const stopWorkers = createManagedStream({
			connect: streamJudgeWorkers,
			on: {
				message: parsed('judge workers', JudgeWorkersEventDataSchema, (data) =>
					judgeWorkersStorage.set(data)
				)
			},
			// 這條 stream 只在 worker 狀態變化時推,重連不會補送 —— 靠 REST 拉回來。
			// 用 resync 版:失敗時保留畫面上的清單,而不是清空後顯示「尚無評分工作者」。
			onResync: () => judgeWorkerApp.resyncJudgeWorkers(),
			onDegraded: (failure) => (workersFailure = failure),
			onRecovered: () => (workersFailure = null)
		});

		// 三條系統狀態 stream 由 BFF 聚合成單一連線,以事件名稱分流。
		const stopStatus = createManagedStream({
			connect: streamSystemStatus,
			on: {
				'judge-queue': parsed('judge-queue', JudgeQueueEventDataSchema, (data) =>
					judgeQueueStorage.set(data)
				),
				'judgment-errors': parsed('judgment-errors', JudgmentErrorsEventDataSchema, (data) =>
					judgmentErrorsStorage.set(data)
				),
				'submission-rate': parsed('submission-rate', SubmissionRateEventDataSchema, (data) =>
					submissionRateStorage.set(data)
				)
			},
			onDegraded: (failure) => (statusFailure = failure),
			onRecovered: () => (statusFailure = null)
		});

		return () => {
			clearInterval(timer);
			stopWorkers();
			stopStatus();
		};
	});

	$effect(() => {
		const projectId = adminCurrentProject.projectId;
		if (projectId == null) return;
		contestProjectApp.fetchJudgmentAnalytics(projectId);
	});

	const workerStatusColorMap: Record<string, 'dark' | 'red' | 'yellow' | 'green' | 'blue'> = {
		INIT: 'dark',
		IDLE: 'green',
		STARTING: 'blue',
		RUNNING: 'green',
		STOPPED: 'yellow',
		ERROR: 'red'
	};

	const formatMs = (ms: number | null | undefined) =>
		ms == null ? '—' : ms >= 1000 ? `${(ms / 1000).toFixed(1)} 秒` : `${Math.round(ms)} ms`;

	const formatSeconds = (s: number | null | undefined) => {
		if (s == null) return '—';
		const total = Math.floor(s);
		return total >= 60 ? `${Math.floor(total / 60)} 分 ${total % 60} 秒` : `${total} 秒`;
	};
</script>

<div class="flex flex-col gap-4 p-4">
	<h1
		class="border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		系統狀態
	</h1>

	{#if streamFailure}
		<div
			role="status"
			class="rounded-lg border border-amber-300 bg-amber-50 p-3 font-medium text-amber-800 dark:border-amber-700 dark:bg-amber-900 dark:text-amber-200"
		>
			{#if streamFailure.terminal}
				即時連線已中斷({streamFailure.status}),以下數字停在最後一次更新。請重新登入或重新整理頁面。
			{:else}
				即時連線異常,以下數字可能不是現況 —— 持續重新連線中。若長時間未恢復,請檢查後端服務。
			{/if}
		</div>
	{/if}

	{#if queueStalled}
		<div
			class="rounded-lg border border-red-300 bg-red-50 p-3 font-medium text-red-800 dark:border-red-700 dark:bg-red-900 dark:text-red-200"
		>
			評分派工疑似卡住:佇列有 {totalQueued} 筆等待,但沒有任何評分在執行中。請檢查評分工作者。
		</div>
	{/if}
	{#if judgeWorkers.length > 0 && runningWorkers.length === 0}
		<div
			class="rounded-lg border border-red-300 bg-red-50 p-3 font-medium text-red-800 dark:border-red-700 dark:bg-red-900 dark:text-red-200"
		>
			所有評分工作者皆未運作，提交將無法評分！請立即檢查評分工作者。
		</div>
	{:else if erroredWorkers.length > 0}
		<div
			class="rounded-lg border border-red-300 bg-red-50 p-3 font-medium text-red-800 dark:border-red-700 dark:bg-red-900 dark:text-red-200"
		>
			有 {erroredWorkers.length} 個評分工作者處於 ERROR 狀態，請前往檢查。
		</div>
	{/if}

	<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
		<!-- 評分工作者(SSE 即時) -->
		<section class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
			<h2 class="mb-3 flex items-center gap-2 text-lg font-bold text-gray-900 dark:text-white">
				評分工作者
				<Badge color={runningWorkers.length > 0 ? 'green' : 'red'}>
					{runningWorkers.length} / {judgeWorkers.length} 運作中
				</Badge>
				<Button size="xs" color="light" class="ml-auto" href="/dashboard/judge-workers">
					管理
				</Button>
			</h2>
			{#if judgeWorkers.length === 0}
				<p class="text-gray-500 dark:text-gray-400">尚無評分工作者。</p>
			{:else}
				<ul class="flex flex-col gap-2">
					{#each judgeWorkers as worker (worker.id)}
						<li class="flex items-center gap-2 text-sm text-gray-900 dark:text-gray-200">
							<Badge color={workerStatusColorMap[worker.status] ?? 'dark'}>{worker.status}</Badge>
							<span class="truncate" title={worker.id}>{worker.id}</span>
							<span class="ml-auto shrink-0 text-gray-500 dark:text-gray-400">
								最後執行:{datetimeString(worker.lastExecutionAt) || '從未'}
							</span>
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<!-- 評分佇列(SSE 即時) -->
		<section class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
			<h2 class="mb-3 flex items-center gap-2 text-lg font-bold text-gray-900 dark:text-white">
				評分佇列
				<Badge color={queueStalled ? 'red' : totalQueued > 0 ? 'blue' : 'green'}>
					{totalQueued} 筆排隊中
				</Badge>
			</h2>
			{#if !judgeQueue}
				<p class="text-gray-500 dark:text-gray-400">Loading...</p>
			{:else}
				<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">完整評分</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-white">
							{judgeQueue.queueDepths.judge}
						</p>
					</div>
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">部分評分</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-white">
							{judgeQueue.queueDepths.judgePartial}
						</p>
					</div>
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">重評</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-white">
							{judgeQueue.queueDepths.rejudge}
						</p>
					</div>
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">執行中</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-white">
							{judgeQueue.activeJudgments}
						</p>
					</div>
				</div>
				<!-- 中性資訊:等待長短取決於評分時長與 worker 數,本身不代表異常 -->
				<p class="mt-3 text-sm text-gray-500 dark:text-gray-400">
					最久等待:
					<span>
						{judgeQueue.oldestQueuedSeconds == null
							? '無排隊'
							: formatSeconds(judgeQueue.oldestQueuedSeconds)}
					</span>
				</p>
				<div class="mt-3 border-t border-gray-200 pt-3 dark:border-gray-700">
					<p class="mb-1 text-sm text-gray-500 dark:text-gray-400">
						近 10 分鐘提交:{recentSubmissionTotal} 筆
					</p>
					<div class="flex h-10 items-end gap-1">
						{#each submissionRate as point (point.minute)}
							<div
								class="flex-1 rounded-t bg-blue-300 dark:bg-blue-600"
								style={`height: ${Math.max(6, (point.count / maxRateCount) * 100)}%`}
								title={`${datetimeString(point.minute)}:${point.count} 筆`}
							></div>
						{:else}
							<p class="text-sm text-gray-400">尚無資料</p>
						{/each}
					</div>
				</div>
			{/if}
		</section>

		<!-- 開放中的比賽 -->
		<section class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
			<h2 class="mb-3 flex items-center gap-2 text-lg font-bold text-gray-900 dark:text-white">
				開放中的比賽
				<Badge color={ongoingContests.length > 0 ? 'green' : 'dark'}>
					{ongoingContests.length}
				</Badge>
				<Button size="xs" color="light" class="ml-auto" href="/dashboard/contests">管理</Button>
			</h2>
			{#if ongoingContests.length === 0}
				<p class="text-gray-500 dark:text-gray-400">目前沒有開放中的比賽。</p>
			{:else}
				<ul class="flex flex-col gap-2">
					{#each ongoingContests as contest (contest.id)}
						<li class="flex items-center gap-2 text-sm text-gray-900 dark:text-gray-200">
							<a href="/dashboard/contests/{contest.id}" class="border-b">{contest.name}</a>
							<span class="ml-auto shrink-0 text-gray-500 dark:text-gray-400">
								{contest.totalProjects} 個專案,開放至 {datetimeString(contest.deactivateTime)}
							</span>
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<!-- 最近的評分錯誤 -->
		<section class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
			<h2 class="mb-3 flex items-center gap-2 text-lg font-bold text-gray-900 dark:text-white">
				最近的評分錯誤
				<Badge color={judgmentErrors.length > 0 ? 'red' : 'green'}>
					{judgmentErrors.length}
				</Badge>
			</h2>
			{#if judgmentErrors.length === 0}
				<p class="text-gray-500 dark:text-gray-400">目前沒有評分錯誤。</p>
			{:else}
				<ul class="flex flex-col gap-2">
					{#each judgmentErrors as err (err.judgmentId)}
						<li class="min-w-0 text-sm">
							<!-- 錯誤訊息可能是整段 stack:斷行 + 截兩行,完整內容進 judgment 頁看 -->
							<a
								href={`/dashboard/submissions/${err.submissionId}/judgments/${err.judgmentId}`}
								class="line-clamp-2 break-all text-red-600 hover:underline dark:text-red-400"
								title={err.errorMessage ?? undefined}
							>
								{datetimeString(err.createdAt)} — {err.errorMessage || '(無錯誤訊息)'}
							</a>
						</li>
					{/each}
				</ul>
			{/if}
		</section>

		<!-- 選定專案的評分管線 -->
		<section class="rounded-lg border border-gray-200 p-4 lg:col-span-2 dark:border-gray-700">
			<h2 class="mb-3 flex items-center gap-2 text-lg font-bold text-gray-900 dark:text-white">
				評分管線
				{#if adminCurrentProject.projectName}
					<span class="font-normal text-gray-500 dark:text-gray-400">
						— {adminCurrentProject.projectName}
					</span>
				{/if}
				{#if adminCurrentProject.projectId != null}
					<Button
						size="xs"
						color="light"
						class="ml-auto"
						onclick={() => contestProjectApp.fetchJudgmentAnalytics(adminCurrentProject.projectId!)}
					>
						重新整理
					</Button>
				{/if}
			</h2>
			{#if adminCurrentProject.projectId == null}
				<p class="text-gray-500 dark:text-gray-400">
					尚未選擇競賽專案 — 請從左側的「競賽過程」選擇,以檢視其評分管線狀態。
				</p>
			{:else if !analytics}
				<p class="text-gray-500 dark:text-gray-400">Loading...</p>
			{:else}
				<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">提交總數</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-white">
							{analytics.totalSubmissions}
						</p>
					</div>
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">評分總數</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-white">
							{analytics.totalJudgments}
						</p>
					</div>
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">平均評分時長</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-white">
							{formatMs(analytics.avgJudgingDurationMs)}
						</p>
					</div>
					<div>
						<p class="text-sm text-gray-500 dark:text-gray-400">平均排隊等待</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-white">
							{formatMs(analytics.avgQueueWaitMs)}
						</p>
					</div>
				</div>
				<div class="mt-4 flex flex-wrap items-center gap-2">
					<span class="text-sm text-gray-500 dark:text-gray-400">評分狀態分布:</span>
					{#each analytics.statusCounts as { status, count } (status)}
						<Badge color={status === 'Error' ? 'red' : status === 'Judging' ? 'blue' : 'dark'}>
							{status}: {count}
						</Badge>
					{:else}
						<span class="text-sm text-gray-500 dark:text-gray-400">尚無評分</span>
					{/each}
					{#if errorJudgments > 0}
						<span class="text-sm font-medium text-red-600 dark:text-red-400">
							⚠ 有 {errorJudgments} 筆評分發生錯誤
						</span>
					{/if}
				</div>
			{/if}
		</section>
	</div>

	<!-- 快速入口 -->
	<section class="flex flex-wrap gap-2">
		<Button color="light" href="/dashboard/submissions">提交紀錄</Button>
		<Button color="light" href="/dashboard/clarifications">公告與選手提問</Button>
		<Button color="light" href="/dashboard/user-scores">分數統計</Button>
		<Button color="light" href="/dashboard/rejudgings">重評紀錄</Button>
	</section>
</div>
