<script lang="ts">
	import Swal from 'sweetalert2';
	import { Accordion, AccordionItem, Alert, Badge, Button } from 'flowbite-svelte';
	import { datetimeString } from '$lib';
	import { nextSummaryScore } from '$lib/summaryScore';
	import { goto } from '$app/navigation';
	import { InfoCircleSolid } from 'flowbite-svelte-icons';
	import { streamUserSubmission } from '$lib/api/stream';
	import {
		SubmissionUpdateEventDataSchema,
		useSubmissionStorage,
		type SubmissionGradingItem,
		type SubmissionTestcase
	} from '@/apps/contest/domain/submission';
	import { useProjectStorage, type UserContestProjectDetail } from '@/apps/contest/domain/project';
	import { useSubmissionApp } from '@/apps/contest/application/submission';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import ShadowHtml from '@/lib/components/ShadowHtml/ShadowHtml.svelte';
	import { AnsiUp } from 'ansi_up';

	const ansiUp = new AnsiUp();
	let showCancelButton = $state(true);
	const { id } = page.params;
	const submissionApp = useSubmissionApp();
	const submissionStorage = useSubmissionStorage();
	const submission = $derived(submissionStorage.get());
	const projectStorage = useProjectStorage();
	const project = $derived(projectStorage.get() as UserContestProjectDetail | null);

	const formatGradingItemName = (item: { label?: string; name: string }) =>
		item.label ? item.label + ": " + item.name : item.name;

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
	// Colour the "得分" badge by lifecycle until the result is final, then by score.
	// Before a conclusion an item's score is not meaningful — a not-yet-scored 0
	// would read as red (looks failed) and a partial mid-run score as green/yellow
	// (looks decided), both misleading. So Pending/NotSubmitted stay neutral (grey),
	// Running is blue (its number still updates live via SSE), and Cancelled — a
	// terminal state that is not a "scored 0" — stays neutral too. Only Passed /
	// Failed / Error / Timeout show the score meaning: green / red / yellow.
	const getScoreColor = (
		status: string,
		score: number,
		totalScore: number
	): 'dark' | 'blue' | 'green' | 'red' | 'yellow' => {
		if (status === 'Pending' || status === 'NotSubmitted' || status === 'Cancelled') return 'dark';
		if (status === 'Running') return 'blue';
		if (score === totalScore) return 'green';
		if (score === 0) return 'red';
		return 'yellow';
	};

	const submissionGradingItemMap = $derived(
		submission?.gradingItems.reduce(
			(acc, gradingItem) => {
				acc[gradingItem.id] = {
					...gradingItem,
					testcases: gradingItem.testcases.reduce(
						(acc, testcase) => {
							acc[testcase.id] = testcase;
							return acc;
						},
						{} as Record<number, SubmissionTestcase>
					)
				};
				return acc;
			},
			{} as Record<
				number,
				Omit<SubmissionGradingItem, 'testcases'> & { testcases: Record<number, SubmissionTestcase> }
			>
		)
	);

	$effect(() => {
		if (submissionStorage.error) {
			Swal.fire({
				icon: 'error',
				title: '錯誤',
				text: '找不到該次提交紀錄'
			});
			submissionStorage.clear();
			goto('/submissions');
		}
	});

	$effect(() => {
		if (!project || !submission) return;

		const eventSource = streamUserSubmission(submission.id);
		eventSource.addEventListener('message', (event) => {
			let parsed;
			try {
				parsed = JSON.parse(event.data);
			} catch (e) {
				console.error('Failed to parse submission event', e);
				return;
			}
			const { success, data, error } = SubmissionUpdateEventDataSchema.safeParse(parsed);
			if (!success) {
				console.error(error);
				return;
			}

			submission.score = data.score;
			submission.status = data.status;
			// reason 目前保證由 REST 初次載入帶入;串流 frame 是否附帶仍在後端定稿,
			// 所以從原始物件防禦性地讀,有才覆寫,不依賴 schema 宣告。
			if (typeof parsed.reason === 'string') submission.reason = parsed.reason;
			// 契約:frame 未帶 staffNote 欄位 = 附註已被清除,所以沒帶就清空(reason 無清除語義,不動)。
			submission.staffNote = typeof parsed.staffNote === 'string' ? parsed.staffNote : undefined;
			const newGradingItemMap = data.gradingItems.reduce(
				(acc, gradingItem) => {
					acc[gradingItem.id] = gradingItem;
					acc[gradingItem.id].testcases = gradingItem.testcases.reduce(
						(acc, testcase) => {
							acc[testcase.id] = testcase;
							return acc;
						},
						{} as Record<number, SubmissionTestcase>
					);
					return acc;
				},
				{} as Record<
					number,
					Omit<SubmissionGradingItem, 'testcases'> & {
						testcases: Record<number, SubmissionTestcase>;
					}
				>
			);
			for (const gradingItem of submission.gradingItems) {
				const newGradingItem = newGradingItemMap[gradingItem.id];
				if (!newGradingItem) continue;
				gradingItem.score = newGradingItem.score;
				gradingItem.status = newGradingItem.status;
				for (let i = 0; i < gradingItem.testcases.length; i++) {
					const newTestcase = newGradingItem.testcases[gradingItem.testcases[i].id];
					if (!newTestcase) continue;
					gradingItem.testcases[i] = newTestcase;
				}
			}
			const submissions = project.submissions;
			const targetSubmission = submissions.find((s) => s.id === submission.id);
			if (!targetSubmission) return;
			targetSubmission.score = data.score;
			targetSubmission.status = data.status;
			// 部分提交不採計彙總分數,只有完整提交能推高側欄分數;規則與 fetchProject 的
			// 彙總計算一致(isPartial ? 0 : score),抽成 nextSummaryScore 讓它可單元測試。
			// 上面兩行是這筆提交自己的得分/狀態,照常更新,只是不進側欄彙總。
			project.score.current = nextSummaryScore(
				project.score.current,
				data.score,
				targetSubmission.isPartial
			);
		});
		eventSource.onerror = (event) => {
			console.error(event);
			const status = (event as MessageEvent & { status?: number }).status;
			if (status === 401 || status === 403) {
				eventSource.close();
				goto('/login');
			}
		};

		return () => {
			eventSource.close();
		};
	});

	onMount(() => {
		if (id) {
			submissionApp.withLoading().fetchSubmission(id);
		}

		return () => {
			submissionStorage.clear();
		};
	});
</script>

{#if !project || !submission}
	<p>Loading...</p>
{:else}
	<div class="p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/submissions"
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
				<p>提交時間: {datetimeString(submission.createdAt)}</p>
				<p class="my-4 flex items-center gap-2 text-gray-500 dark:text-gray-400">
					<Badge color={submission.isPartial ? 'pink' : 'indigo'}>
						{submission.isPartial ? '部分提交' : '完整提交'}</Badge
					>
					<Badge color={statusColorMap[submission.status]}>{submission.status}</Badge>
					<Badge color="blue">
						得分 ({submission.score} / {submission.totalScore})
					</Badge>
				</p>

				{#if showCancelButton && (submission.status === 'Queued' || submission.status === 'Judging')}
					<Button
						color="primary"
						class="ml-auto cursor-pointer"
						on:click={() => {
							Swal.fire({
								title: '確定要取消提交嗎?',
								showCancelButton: true,
								confirmButtonText: '確定',
								cancelButtonText: '取消'
							}).then((result) => {
								if (result.isConfirmed) {
									submissionApp
										.cancelSubmission(submission.id)
										.then(() => {
											Swal.fire({
												icon: 'success',
												title: '提交請求已送出，會在稍後取消評分'
											});
											showCancelButton = false;
										})
										.catch((error) => {
											console.error(error);
											Swal.fire({
												icon: 'error',
												title: '錯誤',
												text: '無法取消提交'
											});
										});
								}
							});
						}}
					>
						取消提交
					</Button>
				{/if}
			</div>
		</h1>

		{#if submission.status === 'Error'}
			<!-- 判定 Error 是評測環境或前置步驟出問題,不是選手的程式。後端會在 reason 給一句
			     中性說明(如「評測前置步驟失敗,請聯絡現場工作人員」);底下這行固定指引不論
			     reason 有沒有都顯示,因為別的 Error(如「無法載入評分項設定」)不含行動指示,
			     指引行要自己把「找人」講清楚。選手看不到任何 lifecycle 指令或輸出——user 端
			     模型本來就沒有。 -->
			<Alert color="red" class="mb-4 dark:bg-red-900 dark:text-red-200">
				<InfoCircleSolid slot="icon" class="h-5 w-5" />
				{#if submission.reason}
					<span class="font-bold">{submission.reason}</span><br />
				{/if}
				此結果並非你的提交造成,請向現場工作人員或裁判確認原因。
			</Alert>
		{:else if submission.status === 'Cancelled' && submission.reason}
			<!-- 取消不是故障,reason 本身就是說明(預設「評測已由工作人員取消」,或管理員
			     自訂、選手可見的文字)。用中性灰、不加 Error 的「並非你的提交造成」指引行。
			     無 reason 的舊資料就不掛告示,狀態徽章已足夠。 -->
			<div
				class="mb-4 flex items-center gap-3 rounded-lg border border-gray-300 bg-gray-50 p-4 text-sm text-gray-700 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-300"
				role="status"
			>
				<InfoCircleSolid class="h-5 w-5 shrink-0" />
				<span>{submission.reason}</span>
			</div>
		{/if}
		{#if submission.staffNote}
			<!-- 工作人員附註:裁判寫給你的話,與系統原因並存、不取代它;任何狀態都可能有。
			     樣式與上面的 reason 區隔(藍),一眼看得出這是人寫的。 -->
			<div
				class="mb-4 rounded-lg border border-blue-300 bg-blue-50 p-4 text-sm text-blue-900 dark:border-blue-700 dark:bg-blue-950 dark:text-blue-200"
				role="note"
			>
				<span class="font-bold">工作人員附註:</span>{submission.staffNote}
			</div>
		{/if}
		<Alert color="yellow" class="mb-4 dark:bg-yellow-800 dark:text-yellow-200">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			僅提供公開的測試案例的輸出結果
		</Alert>

		<Accordion multiple>
			{#each project.gradingItems as gradingItem, i}
				{#if submissionGradingItemMap && submissionGradingItemMap[gradingItem.id]}
					<AccordionItem class="cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700">
						<span slot="header" class="flex items-center space-x-2">
							<span>{formatGradingItemName(gradingItem)}</span>
							<Badge color={testcaseColorMap[submissionGradingItemMap[gradingItem.id].status]}>
								{submissionGradingItemMap[gradingItem.id].status}
							</Badge>
							<Badge
								color={getScoreColor(
									submissionGradingItemMap[gradingItem.id].status,
									submissionGradingItemMap[gradingItem.id].score,
									gradingItem.score
								)}
							>
								得分 {submissionGradingItemMap[gradingItem.id].score} / {gradingItem.score}分
							</Badge>
						</span>
						<section>
							{#if gradingItem.description}
								<p
									class="mb-4 border-b border-gray-200 pb-4 whitespace-pre-wrap text-gray-600 dark:border-gray-700 dark:text-gray-400"
								>
									{gradingItem.description}
								</p>
							{/if}
							<div class="flex flex-col gap-y-4">
								{#each gradingItem.publicTestcases as testcase}
									<Accordion multiple>
										<AccordionItem class="cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700">
											<span slot="header" class="flex items-center space-x-2">
												<span>{testcase.name}</span>
												<Badge
													color={testcaseColorMap[
														submissionGradingItemMap[gradingItem.id].testcases[testcase.id].status
													]}
												>
													{submissionGradingItemMap[gradingItem.id].testcases[testcase.id].status}
												</Badge>
											</span>
											<section>
												<div class="flex flex-col gap-y-4">
													{#if gradingItem.outputFormat === 'text'}
														<pre
															class="flex-1 overflow-auto rounded-lg bg-gray-100 p-4 dark:bg-gray-800"><ShadowHtml
																html={ansiUp.ansi_to_html(
																	submissionGradingItemMap[gradingItem.id].testcases[testcase.id]
																		.stdout || '無輸出'
																)}></ShadowHtml></pre>
													{:else if gradingItem.outputFormat === 'html'}
														<ShadowHtml
															html={submissionGradingItemMap[gradingItem.id].testcases[testcase.id]
																.stdout || '無輸出'}
														></ShadowHtml>
													{/if}
												</div>
											</section>
										</AccordionItem>
									</Accordion>
								{:else}
									<p>無公開的測試案例</p>
								{/each}
							</div>
						</section>
					</AccordionItem>
				{/if}
			{/each}
		</Accordion>
	</div>
{/if}
