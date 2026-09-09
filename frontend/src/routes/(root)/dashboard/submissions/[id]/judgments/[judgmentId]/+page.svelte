<script lang="ts">
	import { page } from '$app/state';
	import { useSubmissionStorage } from '@/apps/admin/domain/submission';
	import StateMessage from '$lib/components/StateMessage/StateMessage.svelte';
	import LifecycleResult from '$lib/components/LifecycleResult/LifecycleResult.svelte';
	import {
		collectFailedSteps,
		describeStep,
		environmentLabel,
		firstLine,
		hasFailed,
		PHASE_LABEL
	} from '$lib/judgmentLifecycle';
	import { datetimeString } from '@/lib';
	import { Accordion, AccordionItem, Badge, Button, Input, Label, Modal, Textarea } from 'flowbite-svelte';
	import Swal from 'sweetalert2';
	import { adminCurrentProject } from '$lib/adminCurrentProject.svelte';
	import { useClarificationApp } from '@/apps/admin/application/clarification';
	import { useSubmissionApp } from '@/apps/admin/application/submission';
	import { notifyStudentTemplates } from '$lib/notifyStudentTemplates';

	// 單一 judgment 的完整檢視。原本是提交詳情頁上的 modal,但評分項目可能上百個,
	// modal 又深又窄;獨立頁面才有足夠寬度,也能 deep link。
	const { id, judgmentId } = page.params;
	const submission = $derived(useSubmissionStorage().get());
	const judgment = $derived(submission?.judgments.find((j) => j.id === judgmentId));

	const formatGradingItemName = (item: { label?: string; name: string }) =>
		item.label ? item.label + ': ' + item.name : item.name;
	const gradingItemNameMap = $derived(
		submission?.gradingItems.reduce(
			(acc, gradingItem) => {
				acc[gradingItem.id] = formatGradingItemName(gradingItem);
				return acc;
			},
			{} as Record<number, string>
		) ?? {}
	);
	const testcaseNameMap = $derived(
		submission?.gradingItems
			.map((g) => g.testcases)
			.flat()
			.reduce(
				(acc, testcase) => {
					acc[testcase.id] = testcase.name;
					return acc;
				},
				{} as Record<number, string>
			) ?? {}
	);

	// 判定層的四組步驟:專案環境 / 測試環境 × Before All / After All。
	const lifecycleGroups = $derived(
		(['project', 'test'] as const).flatMap((environment) =>
			(['beforeAll', 'afterAll'] as const).map((phase) => {
				const runs =
					judgment?.judgmentLifecycle[phase].filter(
						(run) => run.executionEnvironment === environment
					) ?? [];
				return {
					title: `${environmentLabel(environment)} ${PHASE_LABEL[phase]}`,
					runs,
					failed: runs.filter(hasFailed).length
				};
			})
		)
	);
	const lifecycleCount = $derived(lifecycleGroups.reduce((n, group) => n + group.runs.length, 0));

	// 一次真實的 before_all 失敗(mysql Access denied)只能靠翻資料才找到 stderr。
	// 失敗的步驟直接列在頁首:哪一步、exit 多少、stderr 第一行;展開就是完整輸出。
	const failedSteps = $derived(judgment ? collectFailedSteps(judgment) : []);
	const stepNames = $derived({
		gradingItem: (id: number) => gradingItemNameMap[id] || 'Unknown',
		testcase: (id: number) => testcaseNameMap[id] || 'Unknown'
	});

	// 「通知選手」:對這筆提交的選手本人發一則定向 clarification(receiverUserId,
	// 非公告)。送出的 contest-project id 取自側欄目前選定的專案——提交列表本來就以
	// 它為範圍抓取(見 dashboard/+layout 與 submissions/+page),所以從列表點進來時
	// 它正是這筆提交所屬的 contest-project;沒有選定時把鈕停用。
	const clarificationApp = useClarificationApp();
	const templates = $derived(
		submission ? notifyStudentTemplates(datetimeString(submission.createdAt)) : []
	);
	let openNotify = $state(false);
	let notifySubject = $state('');
	let notifyContent = $state('');
	let notifyBusy = $state(false);

	function applyTemplate(t: { subject: string; content: string }) {
		notifySubject = t.subject;
		notifyContent = t.content;
	}

	function openNotifyModal() {
		if (templates[0]) applyTemplate(templates[0]); // 先帶第一式,可再切換或改寫
		openNotify = true;
	}

	function submitNotify(e: SubmitEvent) {
		e.preventDefault();
		const contestProjectId = adminCurrentProject.projectId;
		if (contestProjectId == null || !submission) return;
		notifyBusy = true;
		clarificationApp
			.sendClarification(contestProjectId, {
				subject: notifySubject,
				content: notifyContent,
				receiverUserId: submission.user.id,
				sendToAll: false
			})
			.then(() => {
				openNotify = false;
				Swal.fire({ title: '已通知選手', icon: 'success', showConfirmButton: false, timer: 1500 });
			})
			.catch((error) => {
				console.error(error);
				Swal.fire({ icon: 'error', title: '通知失敗', text: error?.message });
			})
			.finally(() => {
				notifyBusy = false;
			});
	}

	// 給選手的附註:每個 judgment 一則(選手可見,與系統 reason 並存)。last-write-wins、
	// 可清空(送空字串)。judgment 是非同步載入的,所以用「已同步的 judgment id」把草稿
	// 從 judgment.staffNote 帶入一次,不覆蓋使用者正在打的字。
	const submissionApp = useSubmissionApp();
	let staffNoteDraft = $state('');
	let staffNoteSyncedFor = '';
	let staffNoteBusy = $state(false);
	$effect(() => {
		if (judgment && judgment.id !== staffNoteSyncedFor) {
			staffNoteSyncedFor = judgment.id;
			staffNoteDraft = judgment.staffNote ?? '';
		}
	});
	function saveStaffNote() {
		if (!judgment || id === undefined) return;
		staffNoteBusy = true;
		submissionApp
			.setJudgmentStaffNote(judgment.id, staffNoteDraft)
			.then(() => submissionApp.fetchSubmission(id)) // 回存後重抓,讓 store 的 judgment.staffNote 一致
			.then(() => {
				Swal.fire({ title: '附註已儲存', icon: 'success', showConfirmButton: false, timer: 1500 });
			})
			.catch((error) => {
				console.error(error);
				Swal.fire({ icon: 'error', title: '儲存失敗', text: error?.message });
			})
			.finally(() => {
				staffNoteBusy = false;
			});
	}

	const statusColorMap: Record<string, 'dark' | 'red' | 'yellow' | 'green' | 'blue'> = {
		Created: 'dark',
		Queued: 'dark',
		Judging: 'blue',
		Judged: 'green',
		Cancelled: 'yellow',
		Error: 'red'
	};
	const testcaseColorMap: Record<
		string,
		'dark' | 'red' | 'yellow' | 'green' | 'blue' | 'primary' | 'indigo'
	> = {
		Pending: 'dark',
		Running: 'blue',
		Passed: 'green',
		Failed: 'primary',
		Error: 'red',
		Timeout: 'indigo',
		Cancelled: 'yellow'
	};

	// 狀態篩選:評分項目可能上百個,考場上最常見的動作是快速聚焦在未通過的項目。
	type Filter = 'all' | 'failed' | 'passed' | 'running' | 'error';
	let statusFilter = $state<Filter>('all');
	let searchText = $state('');

	const filterOf = (status: string): Exclude<Filter, 'all'> | null => {
		if (status === 'Failed') return 'failed';
		if (status === 'Passed') return 'passed';
		if (status === 'Pending' || status === 'Running') return 'running';
		if (status === 'Error' || status === 'Timeout') return 'error';
		return null;
	};

	const filterChips: { key: Filter; label: string }[] = [
		{ key: 'all', label: '全部' },
		{ key: 'failed', label: '未通過' },
		{ key: 'passed', label: '通過' },
		{ key: 'running', label: '評分中' },
		{ key: 'error', label: '錯誤/逾時' }
	];
	const countOf = (key: Filter) =>
		key === 'all'
			? (judgment?.gradingItems.length ?? 0)
			: (judgment?.gradingItems.filter((gi) => filterOf(gi.status) === key).length ?? 0);

	const filteredGradingItems = $derived(
		(judgment?.gradingItems ?? [])
			.filter((gi) => statusFilter === 'all' || filterOf(gi.status) === statusFilter)
			.filter(
				(gi) =>
					searchText.trim() === '' ||
					(gradingItemNameMap[gi.id] ?? '').toLowerCase().includes(searchText.trim().toLowerCase())
			)
	);
</script>

{#if !submission}
	<p class="p-4">Loading...</p>
{:else if !judgment}
	<StateMessage
		title="找不到此評分紀錄"
		description="它可能已被移除,或網址有誤。"
		tone="error"
	/>
{:else}
	<div class="p-4">
		<h1 class="mb-4 flex items-center gap-3 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<Button href={`/dashboard/submissions/${id}`} color="alternative" class="cursor-pointer p-2">
				←
			</Button>
			<span>
				{datetimeString(judgment.createdAt)} 的評分
			</span>
			<Badge color={statusColorMap[judgment.status]}>
				{judgment.status}
			</Badge>
			<Badge color="dark">{judgment.judgeType}</Badge>
			<span class="ml-auto">
				<span class="text-lg font-bold">{judgment.score}</span> / {judgment.totalScore} 分
			</span>
		</h1>

		{#if judgment.errorMessage}
			<div
				class="mb-4 rounded-md border border-red-300 bg-red-50 p-3 text-sm text-red-800 dark:border-red-700 dark:bg-red-900 dark:text-red-200"
			>
				<span class="font-bold">錯誤原因：</span>
				<span class="break-all">{judgment.errorMessage}</span>
			</div>
		{/if}

		{#if failedSteps.length > 0}
			<div
				class="mb-4 rounded-md border border-red-300 bg-red-50 p-3 text-sm dark:border-red-700 dark:bg-red-950"
				role="alert"
			>
				<div class="mb-2 font-bold text-red-800 dark:text-red-200">
					生命週期步驟失敗({failedSteps.length})
				</div>
				<ul class="flex flex-col gap-2">
					{#each failedSteps as step}
						<li>
							<details>
								<summary class="cursor-pointer select-none">
									<Badge color="red" class="mr-2 font-bold">exit {step.run.result.exitCode}</Badge>
									<span class="font-bold">{describeStep(step.at, stepNames)}</span>
									{#if firstLine(step.run.result.stderr)}
										<code class="ml-2 break-all text-red-700 dark:text-red-300"
											>{firstLine(step.run.result.stderr)}</code
										>
									{/if}
								</summary>
								<div class="mt-2">
									<LifecycleResult run={step.run} title={describeStep(step.at, stepNames)} />
								</div>
							</details>
						</li>
					{/each}
				</ul>
			</div>
		{/if}

		<div class="mb-4 rounded-md border border-gray-200 p-3 dark:border-gray-700">
			<Label for="staff-note" class="font-bold">給選手的附註</Label>
			<p class="mt-1 mb-2 text-sm text-gray-500 dark:text-gray-400">
				選手在提交詳情看得到,與系統原因並存(不取代它)。留空即清除。寫給選手看,不要放指令、路徑或內部細節。
			</p>
			<Textarea id="staff-note" class="h-24" bind:value={staffNoteDraft} />
			<div class="mt-2 flex items-center gap-3">
				<Button
					size="sm"
					color="green"
					class="cursor-pointer"
					onclick={saveStaffNote}
					disabled={staffNoteBusy || staffNoteDraft === (judgment.staffNote ?? '')}
				>
					儲存附註
				</Button>
				{#if staffNoteDraft !== (judgment.staffNote ?? '')}
					<span class="text-sm text-amber-600 dark:text-amber-400">尚未儲存</span>
				{/if}
				<!-- 通知走 modal,和附註是兩件事:留空純通知(如「已重新評測」)也要能發,
				     所以不 gate 在附註非空;仍需選定競賽專案才知道發到哪個 contest-project。 -->
				<Button
					size="sm"
					color="light"
					class="ml-auto cursor-pointer"
					disabled={adminCurrentProject.projectId == null}
					title={adminCurrentProject.projectId == null ? '請先在側欄選擇競賽專案' : undefined}
					onclick={openNotifyModal}
				>
					通知選手
				</Button>
			</div>
		</div>

		<Accordion multiple>
			<AccordionItem>
				<span slot="header">執行環境資訊</span>
				<div class="flex flex-col gap-4">
					<div class="flex items-center gap-4">
						<span class="font-bold">執行環境 Tag：</span>
						<span>{judgment.project.imageTag}</span>
					</div>
					<div class="flex items-center gap-4">
						<span class="font-bold">掛載路徑：</span>
						<span>{judgment.project.mountPath}</span>
					</div>
					<Accordion>
						<AccordionItem>
							<span slot="header">環境變數</span>
							<code
								class="block max-h-60 min-h-20 overflow-auto rounded-md bg-gray-100 p-2 whitespace-pre text-gray-600"
							>
								{judgment.project.env}
							</code>
						</AccordionItem>
					</Accordion>
					<div class="flex items-center gap-4">
						<span class="font-bold">測試環境 Tag：</span>
						<span>{judgment.project.testImageTag}</span>
					</div>
					<Accordion>
						<AccordionItem>
							<span slot="header">測試環境變數</span>
							<code
								class="block max-h-60 min-h-20 overflow-auto rounded-md bg-gray-100 p-2 whitespace-pre text-gray-600"
							>
								{judgment.project.testEnv}
							</code>
						</AccordionItem>
					</Accordion>
				</div>

			</AccordionItem>
			<AccordionItem open={failedSteps.length > 0}>
				<span slot="header">
					生命週期執行結果 ({lifecycleCount})
					{#if failedSteps.length > 0}
						<Badge color="red" class="ml-2">{failedSteps.length} 失敗</Badge>
					{/if}
				</span>
				<Accordion multiple>
					{#each lifecycleGroups as group (group.title)}
						<AccordionItem open={group.failed > 0}>
							<span slot="header">
								{group.title} ({group.runs.length})
								{#if group.failed > 0}
									<Badge color="red" class="ml-2">{group.failed} 失敗</Badge>
								{/if}
							</span>
							{#if group.runs.length === 0}
								<p class="text-gray-500 dark:text-gray-400">沒有步驟。</p>
							{:else}
								<div class="flex flex-col gap-3">
									{#each group.runs as run, index}
										<LifecycleResult {run} title={`${group.title} #${index + 1}`} />
									{/each}
								</div>
							{/if}
						</AccordionItem>
					{/each}
				</Accordion>
			</AccordionItem>
		</Accordion>

		<h2 class="mt-6 mb-3 text-xl font-bold">評分項目</h2>
		<div class="mb-4 flex flex-wrap items-center gap-2">
			{#each filterChips as chip (chip.key)}
				<button
					class={[
						'cursor-pointer rounded-full border px-3 py-1 text-sm',
						statusFilter === chip.key
							? 'border-blue-600 bg-blue-600 text-white'
							: 'border-gray-300 bg-white text-gray-700 hover:bg-gray-100 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700'
					]}
					onclick={() => (statusFilter = chip.key)}
				>
					{chip.label} ({countOf(chip.key)})
				</button>
			{/each}
			<div class="ml-auto w-64">
				<Input bind:value={searchText} placeholder="搜尋評分項目名稱..." size="sm" />
			</div>
		</div>

		{#if filteredGradingItems.length === 0}
			<p class="text-gray-500 dark:text-gray-400">沒有符合條件的評分項目。</p>
		{:else}
			<Accordion multiple>
				{#each filteredGradingItems as gradingItem (gradingItem.id)}
					<!-- 預設收合:上百個評分項全展開時,連切換篩選都會卡頓
					     (收合的 AccordionItem 不渲染內容);點開單項時測資維持展開。 -->
					<AccordionItem>
						<span slot="header">
							<Badge color={testcaseColorMap[gradingItem.status]} class="mr-2"
								>{gradingItem.status}</Badge
							>
							<span class="ml-2"
								>{gradingItemNameMap[gradingItem.id] || 'Unknown'} - {gradingItem.score} / {gradingItem.totalScore}
								分</span
							>
						</span>
						<Accordion multiple>
							{#each gradingItem.testcases as testcase (testcase.id)}
								<AccordionItem open>
									<span slot="header">
										<Badge color={testcaseColorMap[testcase.status]} class="mr-2"
											>{testcase.status}</Badge
										>
										<span class="ml-2"
											>{testcaseNameMap[testcase.id] || 'Unknown'} - {testcase.score} / {testcase.totalScore}
											分</span
										>
									</span>
									<div class="flex items-center gap-4">
										<span class="font-bold">指令：</span>
										<span>{testcase.command}</span>
									</div>
									<div class="flex items-center gap-4">
										<span class="font-bold">結束狀態碼：</span>
										<span>{testcase.result.exitCode}</span>
									</div>
									<div class="flex items-center gap-4">
										<span class="font-bold">執行時間：</span>
										<span>{testcase.result.execTime} ms</span>
									</div>
									<div>
										<div class="font-bold">stdout：</div>
										<code
											class="block max-h-60 min-h-20 overflow-auto rounded-md bg-gray-100 p-2 whitespace-pre text-gray-600"
										>
											{testcase.result.stdout}
										</code>
									</div>
									<div>
										<div class="font-bold">stderr：</div>
										<code
											class="block max-h-60 min-h-20 overflow-auto rounded-md bg-gray-100 p-2 whitespace-pre text-gray-600"
										>
											{testcase.result.stderr}
										</code>
									</div>
									<!-- 測資自己的前置/後置步驟,和判定層一樣歸位到它所屬的測資底下。 -->
									{#if testcase.lifecycle.beforeEach.length + testcase.lifecycle.afterEach.length > 0}
										<div class="mt-3 flex flex-col gap-3">
											{#each testcase.lifecycle.beforeEach as run, index}
												<LifecycleResult {run} title={`Before Each #${index + 1}`} />
											{/each}
											{#each testcase.lifecycle.afterEach as run, index}
												<LifecycleResult {run} title={`After Each #${index + 1}`} />
											{/each}
										</div>
									{/if}
								</AccordionItem>
							{/each}
						</Accordion>
					</AccordionItem>
				{/each}
			</Accordion>
		{/if}

		<Modal bind:open={openNotify} outsideclose title="通知選手">
			<form id="notify-form" class="space-y-4" onsubmit={submitNotify}>
				<div class="text-sm text-gray-600 dark:text-gray-400">
					收件選手:<span class="font-bold">{submission.user.name} ({submission.user.username})</span>
					—— 這是<strong>定向</strong>訊息,只有這位選手看得到,不是公告。
				</div>
				{#if staffNoteDraft !== (judgment.staffNote ?? '')}
					<div
						class="rounded-md border border-amber-300 bg-amber-50 p-2 text-sm text-amber-800 dark:border-amber-700 dark:bg-amber-950 dark:text-amber-200"
					>
						<strong>附註尚未儲存</strong>:通知仍會送出,但選手打開提交詳情看到的是先前的附註。若要選手看到你剛打的附註,請先關掉這裡去按「儲存附註」。
					</div>
				{/if}
				<div class="flex flex-wrap gap-2">
					{#each templates as t (t.key)}
						<Button size="xs" color="light" type="button" onclick={() => applyTemplate(t)}>
							{t.label}
						</Button>
					{/each}
				</div>
				<div>
					<Label for="notify-subject"><span class="align-middle text-red-500">*</span> 標題</Label>
					<Input id="notify-subject" required bind:value={notifySubject} />
				</div>
				<div>
					<Label for="notify-content"><span class="align-middle text-red-500">*</span> 內容</Label>
					<Textarea id="notify-content" class="h-32" required bind:value={notifyContent} />
				</div>
			</form>
			<svelte:fragment slot="footer">
				<Button color="green" form="notify-form" type="submit" class="w-full" disabled={notifyBusy}>
					送出通知
				</Button>
			</svelte:fragment>
		</Modal>
	</div>
{/if}
