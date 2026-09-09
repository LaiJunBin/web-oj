<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSubmissionApp } from '@/apps/admin/application/submission';
	import { useSubmissionStorage } from '@/apps/admin/domain/submission';
	import { datetimeString } from '@/lib';
	import {
		Badge,
		Button,
		ButtonGroup,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const statusColorMap: Record<string, 'dark' | 'red' | 'yellow' | 'green' | 'blue'> = {
		Created: 'dark',
		Queued: 'dark',
		Judging: 'blue',
		Judged: 'green',
		Cancelled: 'yellow',
		Error: 'red'
	};

	const submissionApp = useSubmissionApp();
	const submission = $derived(useSubmissionStorage().get()!);

	const formatGradingItemName = (item: { label?: string; name: string }) =>
		item.label ? item.label + ': ' + item.name : item.name;
</script>

{#if !submission}
	<div>Loading...</div>
{:else}
	<div class="p-4">
		<h1 class="mb-4 flex items-center border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<span>
				{datetimeString(submission.createdAt)}的提交
			</span>
			<Badge color={submission.isPartial ? 'yellow' : 'green'} class="ml-2">
				{submission.isPartial ? '部分提交' : '完整提交'}
			</Badge>
			{#if submission.ignored}
				<Badge color="red" class="ml-2">已忽略</Badge>
			{/if}
			<ButtonGroup class="ml-auto">
				<Button
					class="cursor-pointer"
					color="primary"
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
								submissionApp
									.withLoading()
									.rejudgeSubmission(submission.id, {
										reason: result.value
									})
									.then(() => {
										submissionApp
											.fetchSubmission(submission.id)
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
					請求重新評分
				</Button>
				<Button
					class="cursor-pointer"
					color="alternative"
					onclick={async () => {
						try {
							if (submission.ignored) {
								await submissionApp.withLoading().unignoreSubmission(submission.id);
							} else {
								await Swal.fire({
									title: '確定要忽略這個提交嗎？',
									showCancelButton: true,
									confirmButtonText: '確定',
									cancelButtonText: '取消'
								}).then(async (result) => {
									if (result.isConfirmed) {
										await submissionApp.withLoading().ignoreSubmission(submission.id);
									}
								});
							}
							await submissionApp.fetchSubmission(submission.id);
						} catch (error) {
							console.error(error);
							Swal.fire({
								icon: 'error',
								title: '操作失敗'
							});
						}
					}}
				>
					{submission.ignored ? '取消忽略' : '忽略'}
				</Button>
			</ButtonGroup>
		</h1>

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<div>
				<h3 class="text-lg font-bold">提交者</h3>
				<a href={`/dashboard/users/${submission.user.id}`} class="border-b">
					{submission.user.name}
				</a>
			</div>

			<div>
				<h3 class="text-lg font-bold">競賽</h3>
				<a href={`/dashboard/contests/${submission.contest.id}`} class="border-b">
					{submission.contest.name}
				</a>
			</div>

			<div>
				<h3 class="text-lg font-bold">最終採計分數</h3>
				<p>
					{submission.judgments[submission.judgments.length - 1].score} /
					{submission.judgments[submission.judgments.length - 1].totalScore}
				</p>
			</div>

			<div>
				<h3 class="text-lg font-bold">專案</h3>
				<a href={`/dashboard/projects/${submission.project.id}`} class="border-b">
					{submission.project.name}
				</a>
			</div>
		</div>

		{#if submission.isPartial}
			<hr class="my-4" />
			<h2 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">該次提交的評分項</h2>
			<Table>
				<TableHead>
					<TableHeadCell>項次</TableHeadCell>
					<TableHeadCell>評分項目</TableHeadCell>
					<TableHeadCell>分數</TableHeadCell>
				</TableHead>
				<TableBody>
					{#each submission.gradingItems as gradingItem}
						<TableBodyRow
							class="cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
							onclick={(e) => {
								const target = e.target as HTMLElement;
								if (target.tagName === 'TD') {
									goto(
										`/dashboard/projects/${submission.project.id}/grading-items/${gradingItem.id}`
									);
								}
							}}
						>
							<TableBodyCell>{gradingItem.label}</TableBodyCell>
							<TableBodyCell>{formatGradingItemName(gradingItem)}</TableBodyCell>
							<TableBodyCell>{gradingItem.score}</TableBodyCell>
						</TableBodyRow>
					{:else}
						<TableBodyRow>
							<TableBodyCell colspan={3} class="text-center">所有</TableBodyCell>
						</TableBodyRow>
					{/each}
				</TableBody>
			</Table>
		{/if}

		<hr class="my-4" />
		<h2 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">裁判紀錄</h2>
		<Table>
			<TableHead>
				<TableHeadCell>#</TableHeadCell>
				<TableHeadCell>評分種類</TableHeadCell>
				<TableHeadCell>分數</TableHeadCell>
				<TableHeadCell>狀態</TableHeadCell>
				<TableHeadCell>提交時間</TableHeadCell>
			</TableHead>
			<TableBody>
				{#each submission.judgments as judgment}
					<TableBodyRow
						class="cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
						onclick={(e) => {
							const target = e.target as HTMLElement;
							if (target.tagName === 'TD') {
								goto(`/dashboard/submissions/${submission.id}/judgments/${judgment.id}`);
							}
						}}
					>
						<TableBodyCell>
							{submission.judgments.length - submission.judgments.indexOf(judgment)}
						</TableBodyCell>
						<TableBodyCell>
							{judgment.judgeType}
						</TableBodyCell>
						<TableBodyCell>{judgment.score} / {judgment.totalScore}</TableBodyCell>
						<TableBodyCell>
							<Badge color={statusColorMap[judgment.status]}>{judgment.status}</Badge>
						</TableBodyCell>
						<TableBodyCell>{datetimeString(judgment.createdAt)}</TableBodyCell>
					</TableBodyRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}

