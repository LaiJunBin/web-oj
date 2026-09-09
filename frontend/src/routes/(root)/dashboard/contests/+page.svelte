<script lang="ts">
	import { goto } from '$app/navigation';
	import { refreshContestState, resetContestState } from '@/lib/refreshContestState';
	import { useContestApp } from '@/apps/admin/application/contest';
	import { useContestsStorage } from '@/apps/admin/domain/contest';
	import { datetimeString } from '@/lib';
	import {
		Button,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';
	import { sanitizeHtml } from '$lib/sanitize';

	const contestApp = useContestApp();
	const contestsStorage = useContestsStorage();
	const contests = $derived(contestsStorage.get()!);
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>競賽管理</span>
		<Button href="/dashboard/contests/create" color="green">新增</Button>
		<Button
			class="cursor-pointer"
			color="blue"
			onclick={() => {
				const input = document.createElement('input');
				input.type = 'file';
				input.accept = '.zip';
				input.onchange = () => {
					const file = input.files?.[0];
					if (!file) return;
					contestApp
						.importContest(file)
						.then((result) => {
							Swal.fire({ icon: 'success', title: '匯入成功', text: ((r) => `已建立競賽「${r.name}」,含 ${r.createdProjects} 個題目` + (r.unresolvedProjects.length ? `;未對應:${r.unresolvedProjects.join("、")}` : ""))(result) });
							contestApp.fetchContests();
							// 導覽列的比賽下拉讀的是 contest 端的 store,不是上面那個。
							refreshContestState();
						})
						.catch((error) =>
							Swal.fire({
								icon: 'error',
								title: '匯入失敗',
								html: `<pre style="text-align:left;white-space:pre-wrap">${sanitizeHtml(error.message)}</pre>`
							})
						);
				};
				input.click();
			}}>匯入</Button
		>
	</h1>

	<Table striped hoverable shadow>
		<TableHead>
			<TableHeadCell>#</TableHeadCell>
			<TableHeadCell>名稱</TableHeadCell>
			<TableHeadCell>描述訊息</TableHeadCell>
			<TableHeadCell>起訖時間</TableHeadCell>
			<TableHeadCell>是否公開</TableHeadCell>
			<TableHeadCell>專案數量</TableHeadCell>
			<TableHeadCell>
				<span class="sr-only">Actions</span>
			</TableHeadCell>
		</TableHead>
		<TableBody tableBodyClass="divide-y">
			{#each contests as contest, index}
				<TableBodyRow
					class="cursor-pointer"
					onclick={(e) => {
						const target = e.target as HTMLElement;
						if (target.tagName === 'TD') {
							goto(`/dashboard/contests/${contest.id}`);
						}
					}}
				>
					<TableBodyCell>{index + 1}</TableBodyCell>
					<TableBodyCell>{contest.name}</TableBodyCell>
					<TableBodyCell>
						<button
							class="cursor-pointer"
							onclick={(e) => {
								e.stopPropagation();
								Swal.fire({
									title: '描述訊息',
									input: 'text',
									inputValue: contest.description,
									showCancelButton: true,
									confirmButtonText: '確定',
									cancelButtonText: '取消'
								}).then((result) => {
									if (!result.isConfirmed) {
										return;
									}
									const description = result.value;
									contestApp
										.updateContest(contest.id, { description })
										.then(() => {
											refreshContestState();
											Swal.fire({
												title: '更新成功',
												icon: 'success',
												confirmButtonText: '確定'
											});
											contestApp.fetchContests();
										})
										.catch((error) => {
											console.error(error);
											Swal.fire({
												title: '更新失敗',
												icon: 'error',
												confirmButtonText: '確定'
											});
										});
								});
							}}
						>
							{contest.description}
						</button>
					</TableBodyCell>
					<TableBodyCell>
						{datetimeString(contest.activateTime)} ~ {datetimeString(contest.deactivateTime)}
					</TableBodyCell>
					<TableBodyCell>
						{contest.isPublic ? '是' : '否'}
					</TableBodyCell>
					<TableBodyCell>
						{contest.totalProjects}
					</TableBodyCell>
					<TableBodyCell>
						<Button
							pill
							color="red"
							class="cursor-pointer"
							onclick={() => {
								Swal.fire({
									title: '確定要刪除嗎？',
									text: `確定要刪除 ${contest.name} 嗎？`,
									icon: 'warning',
									showCancelButton: true,
									confirmButtonText: '確定',
									cancelButtonText: '取消'
								}).then((result) => {
									if (!result.isConfirmed) {
										return;
									}
									contestApp
										.deleteContest(contest.id)
										.then(() => {
											resetContestState();
											Swal.fire({
												title: '刪除成功',
												icon: 'success',
												confirmButtonText: '確定'
											});
											contestApp.fetchContests();
										})
										.catch((error) => {
											console.error(error);
											Swal.fire({
												title: '刪除失敗',
												icon: 'error',
												confirmButtonText: '確定'
											});
										});
								});
							}}>刪除</Button
						>
					</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>
</div>
