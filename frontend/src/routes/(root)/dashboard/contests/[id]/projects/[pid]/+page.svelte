<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { useContestProjectApp } from '@/apps/admin/application/contest-project';
	import { useContestProjectStorage } from '@/apps/admin/domain/contest-project';
	import ClarificationsPanel from '$lib/components/admin/ClarificationsPanel.svelte';
	import UserScoresPanel from '$lib/components/admin/UserScoresPanel.svelte';
	import { datetimeString } from '@/lib';
	import {
		Badge,
		Button,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const contestProjectApp = useContestProjectApp();
	const contestProjectStorage = useContestProjectStorage();
	const contestProject = $derived(contestProjectStorage.get()!);

	const getDuration = (startTime: string, endTime: string) => {
		const total =
			new Date(new Date(endTime).getTime() - new Date(startTime).getTime()).getTime() / 1000;
		const day = Math.floor(total / 86400).toString();
		const hour = Math.floor((total % 86400) / 3600)
			.toString()
			.padStart(2, '0');
		const minute = Math.floor((total % 3600) / 60)
			.toString()
			.padStart(2, '0');
		const second = Math.floor(total % 60).toString().padStart(2, '0');

		if (day !== '0') {
			return `${day} 天 ${hour} 小時 ${minute} 分鐘 ${second} 秒`;
		}

		if (hour === '00') {
			if (minute === '00') {
				return `${second} 秒`;
			}
			return `${minute} 分鐘 ${second} 秒`;
		}

		return `${hour} 小時 ${minute} 分鐘 ${second} 秒`;
	};

	const statusColorMap: Record<string, 'dark' | 'red' | 'yellow' | 'green' | 'blue'> = {
		Created: 'dark',
		Queued: 'dark',
		Judging: 'blue',
		Judged: 'green',
		Cancelled: 'yellow',
		Error: 'red'
	};
</script>

{#if !contestProject}
	<div>Loading...</div>
{:else}
	<div class="flex flex-col p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/dashboard/contests/{page.params.id}"
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
				<span>{contestProject.name}</span>
				<Badge color={contestProject.isPublic ? 'green' : 'red'} class="ml-2">
					{contestProject.isPublic ? '公開' : '非公開'}
				</Badge>
				<Button
					href="/dashboard/contests/{page.params.id}/projects/{page.params.pid}/analytics"
					color="light"
					class="ml-auto cursor-pointer"
				>
					查看評分分析
				</Button>
				<Button
					color="primary"
					class="cursor-pointer"
					onclick={() => {
						Swal.fire({
							title: '確定要請求重新評分嗎？',
							showCancelButton: true,
							input: 'text',
							inputLabel: '請輸入理由',
							inputValidator: (value) => {
								if (!value) {
									return '請輸入理由';
								}
							},
							confirmButtonText: '確定',
							cancelButtonText: '取消'
						}).then((result) => {
							if (result.isConfirmed) {
								contestProjectApp
									.withLoading()
									.rejudgeContestProject(contestProject.id, {
										reason: result.value
									})
									.then(() => {
										contestProjectApp
											.fetchContestProject(contestProject.id)
											.then(() => {
												Swal.fire({
													icon: 'success',
													title: '操作成功'
												});
											})
											.catch((err) => {
												console.error(err);
												Swal.fire({
													icon: 'error',
													title: '操作失敗'
												});
											});
									})
									.catch((err) => {
										console.error(err);
										Swal.fire({
											icon: 'error',
											title: '操作失敗',
											text: err.message
										});
									});
							}
						});
					}}
				>
					重新評分
				</Button>
			</div>
		</h1>

		<div class="mb-4">
			<h2 class="text-lg font-bold">專案描述</h2>
			<p>{contestProject.description}</p>
		</div>

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<div>
				<h2 class="text-lg font-bold">專案開啟時間</h2>
				<p>{datetimeString(contestProject.activateTime)}</p>
			</div>

			<div>
				<h2 class="text-lg font-bold">開始競賽時間</h2>
				<p>{datetimeString(contestProject.startTime)}</p>
			</div>

			<div>
				<h2 class="text-lg font-bold">專案結束時間</h2>
				<p>{datetimeString(contestProject.deactivateTime)}</p>
			</div>

			<div>
				<h2 class="text-lg font-bold">結束競賽時間</h2>
				<p>{datetimeString(contestProject.endTime)}</p>
			</div>

			<div>
				<h2 class="text-lg font-bold">評分優先權策略</h2>
				<p>{contestProject.judgePriorityStrategy}</p>
			</div>

			<div>
				<h2 class="text-lg font-bold">競賽時長</h2>
				<p>{getDuration(contestProject.startTime, contestProject.endTime)}</p>
			</div>

			<div>
				<h2 class="text-lg font-bold">專案執行環境</h2>
				<a href="/dashboard/images/{contestProject.image}" class="border-b">
					{contestProject.image}
				</a>
			</div>

			<div>
				<h2 class="text-lg font-bold">專案測試環境</h2>
				<a href="/dashboard/images/{contestProject.testImage}" class="border-b">
					{contestProject.testImage}
				</a>
			</div>
		</div>

		<hr class="my-4" />
		<div class="flex flex-col items-center justify-center">
			<h2 class="mb-4 flex items-center gap-x-2 text-2xl font-bold">
				<p>提交紀錄</p>
			</h2>
			<Table>
				<TableHead>
					<TableHeadCell>編號</TableHeadCell>
					<TableHeadCell>提交類型</TableHeadCell>
					<TableHeadCell>狀態</TableHeadCell>
					<TableHeadCell>使用者</TableHeadCell>
					<TableHeadCell>得分/總分</TableHeadCell>
					<TableHeadCell>提交時間</TableHeadCell>
				</TableHead>
				<TableBody>
					{#each contestProject.submissions as submission, index}
						<TableBodyRow
							class="cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
							onclick={(e) => {
								const target = e.target as HTMLElement;
								if (target.tagName === 'TD') {
									goto(`/dashboard/submissions/${submission.id}`);
								}
							}}
						>
							<TableBodyCell>{contestProject.submissions.length - index}</TableBodyCell>
							<TableBodyCell>
								<Badge color={submission.isPartial ? 'pink' : 'indigo'}>
									{submission.isPartial ? '部分提交' : '完整提交'}</Badge
								>
							</TableBodyCell>
							<TableBodyCell>
								<Badge color={statusColorMap[submission.status]}>{submission.status}</Badge>
							</TableBodyCell>
							<TableBodyCell>
								<a href={`/dashboard/users/${submission.user.id}`} class="border-b">
									{submission.user.name}({submission.user.username})
								</a>
							</TableBodyCell>
							<TableBodyCell>{submission.score} / {submission.totalScore}</TableBodyCell>
							<TableBodyCell>{datetimeString(submission.createdAt)}</TableBodyCell>
						</TableBodyRow>
					{:else}
						<TableBodyRow>
							<TableBodyCell colspan={8} class="text-center">沒有提交紀錄</TableBodyCell>
						</TableBodyRow>
					{/each}
				</TableBody>
			</Table>
		</div>

		<!-- 提問與分數也整合在此,讓已結束/未開始的專案(不在競賽過程選單內)
		     一樣有完整入口。 -->
		<hr class="my-4" />
		<ClarificationsPanel projectId={contestProject.id} projectName={contestProject.name} />

		<hr class="my-4" />
		<h2 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">分數統計</h2>
		<UserScoresPanel projectId={contestProject.id} projectName={contestProject.name} />
	</div>
{/if}
