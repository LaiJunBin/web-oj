<script lang="ts">
	import { goto } from '$app/navigation';
	import { useRejudgeApp } from '@/apps/admin/application/rejudge';
	import { useRejudgeStorage } from '@/apps/admin/domain/rejudge';
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
		TableHeadCell,
		type ColorVariant
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const rejudgeApp = useRejudgeApp();
	const rejudge = $derived(useRejudgeStorage().get()!);

	const rejudgeColorMap: Record<string, ColorVariant> = {
		Pending: 'dark',
		Running: 'blue',
		Finished: 'primary',
		Approved: 'green',
		Rejected: 'pink',
		Canceled: 'yellow',
		Error: 'red'
	};
</script>

{#if !rejudge}
	<div>Loading...</div>
{:else}
	<div class="p-4">
		<h1 class="mb-4 flex items-center border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<span class="flex items-center gap-2">
				<span>{datetimeString(rejudge.createdAt)}的重新評分</span>
				<Badge color={rejudgeColorMap[rejudge.status]}>{rejudge.status}</Badge>
			</span>
			<ButtonGroup class="ml-auto">
				{#if rejudge.status === 'Pending' || rejudge.status === 'Running'}
					<Button
						class="cursor-pointer"
						color="primary"
						onclick={() => {
							Swal.fire({
								title: '確定要取消重新評分嗎？',
								showCancelButton: true,
								confirmButtonText: '確定',
								cancelButtonText: '取消'
							}).then((result) => {
								if (result.isConfirmed) {
									rejudgeApp
										.withLoading()
										.cancelRejudge(rejudge.id)
										.then(() => {
											rejudgeApp
												.fetchRejudge(rejudge.id)
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
												title: '操作失敗'
											});
										});
								}
							});
						}}
					>
						取消重新評分
					</Button>
				{/if}

				{#if rejudge.status === 'Finished'}
					<Button
						class="cursor-pointer"
						color="green"
						onclick={() => {
							Swal.fire({
								title: '做出決策',
								text: '通過重新評分？',
								showDenyButton: true,
								showCancelButton: true,
								confirmButtonText: '通過',
								denyButtonText: '拒絕',
								cancelButtonText: '取消'
							}).then(async (result) => {
								if (result.isConfirmed) {
									await rejudgeApp
										.withLoading()
										.approveRejudge(rejudge.id)
										.then(() => {
											rejudgeApp
												.fetchRejudge(rejudge.id)
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
												title: '操作失敗'
											});
										});
								} else if (result.isDenied) {
									await rejudgeApp
										.withLoading()
										.rejectRejudge(rejudge.id)
										.then(() => {
											rejudgeApp
												.fetchRejudge(rejudge.id)
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
												title: '操作失敗'
											});
										});
								}
							});
						}}
					>
						做出決策
					</Button>
				{/if}
			</ButtonGroup>
		</h1>

		<h2 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">重新評分原因</h2>
		<p>{rejudge.reason}</p>

		<hr class="my-4" />
		<h2 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			此次重新評分受影響的提交
		</h2>
		<Table>
			<TableHead>
				<TableHeadCell>#</TableHeadCell>
				<TableHeadCell>狀態</TableHeadCell>
				<TableHeadCell>分數(前/後)</TableHeadCell>
				<TableHeadCell>總分(前/後)</TableHeadCell>
			</TableHead>
			<TableBody>
				{#each rejudge.submissions as submission, index}
					<TableBodyRow
						class="cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
						onclick={(e) => {
							const target = e.target as HTMLElement;
							if (target.tagName === 'TD') {
								goto(`/dashboard/rejudgings/${rejudge.id}/submissions/${submission.id}`);
							}
						}}
					>
						<TableBodyCell>
							{rejudge.submissions.length - index}
						</TableBodyCell>
						<TableBodyCell>
							{submission.after.status}
						</TableBodyCell>
						<TableBodyCell>
							{submission.before ? submission.before.score : '無前次評分'} / {submission.after.score}
						</TableBodyCell>
						<TableBodyCell>
							{submission.before ? submission.before.totalScore : '無前次評分'} / {submission
								.after.totalScore}
						</TableBodyCell>
					</TableBodyRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
