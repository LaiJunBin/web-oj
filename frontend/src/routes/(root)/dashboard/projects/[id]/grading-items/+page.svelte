<script lang="ts">
	import { goto } from '$app/navigation';
	import { useGradingItemApp } from '@/apps/admin/application/grading-item';
	import { useGradingItemsStorage } from '@/apps/admin/domain/grading-item';
	import { useProjectStorage } from '@/apps/admin/domain/project';
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

	const gradingItemApp = useGradingItemApp();
	const gradingItemsStorage = useGradingItemsStorage();
	const gradingItems = $derived(gradingItemsStorage.get());
	const project = $derived(useProjectStorage().get());

	const formatGradingItemName = (item: { label?: string; name: string }) =>
		item.label ? item.label + ": " + item.name : item.name;
</script>

{#if !project || !gradingItems}
	<div>Loading...</div>
{:else}
	<div class="p-4">
		<h1
			class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
		>
			<Button
				href="/dashboard/projects/{project.id}"
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
			<span>專案 {project.name} 的評分項目 </span>
			<Button
				color="green"
				class="cursor-pointer"
				href="/dashboard/projects/{project.id}/grading-items/create">新增</Button
			>
		</h1>

		<Table striped hoverable shadow>
			<TableHead>
				<TableHeadCell>項次</TableHeadCell>
				<TableHeadCell>名稱</TableHeadCell>
				<TableHeadCell>描述訊息</TableHeadCell>
				<TableHeadCell>分數</TableHeadCell>
				<TableHeadCell>測試案例數量</TableHeadCell>
				<TableHeadCell>
					<span class="sr-only">Actions</span>
				</TableHeadCell>
			</TableHead>
			<TableBody tableBodyClass="divide-y">
				{#each gradingItems as gradingItem}
					<TableBodyRow
						class="cursor-pointer"
						onclick={(e) => {
							const target = e.target as HTMLElement;
							if (target.tagName === 'TD') {
								goto(`/dashboard/projects/${project.id}/grading-items/${gradingItem.id}`);
							}
						}}
					>
						<TableBodyCell>{gradingItem.label}</TableBodyCell>
						<TableBodyCell tdClass="px-6 py-4 font-medium text-gray-900 dark:text-white">{formatGradingItemName(gradingItem)}</TableBodyCell>
						<TableBodyCell tdClass="px-6 py-4 font-medium text-gray-900 dark:text-white">
							<button
								class="cursor-pointer text-left"
								onclick={(e) => {
									e.stopPropagation();
									Swal.fire({
										title: '描述訊息',
										input: 'text',
										inputValue: gradingItem.description,
										showCancelButton: true,
										confirmButtonText: '確定',
										cancelButtonText: '取消'
									}).then((result) => {
										if (!result.isConfirmed) {
											return;
										}
										const description = result.value;
										gradingItemApp
											.updateGradingItem(gradingItem.id, { description })
											.then(() => {
												Swal.fire({
													title: '更新成功',
													icon: 'success',
													confirmButtonText: '確定'
												});
												gradingItemApp.fetchGradingItems(project.id);
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
								{gradingItem.description}
							</button>
						</TableBodyCell>
						<TableBodyCell>
							{gradingItem.score}
						</TableBodyCell>
						<TableBodyCell>
							{gradingItem.testcaseLength}
						</TableBodyCell>
						<TableBodyCell>
							<Button
								pill
								color="red"
								class="cursor-pointer"
								onclick={() => {
									Swal.fire({
										title: '確定要刪除嗎？',
										text: `確定要刪除 ${formatGradingItemName(gradingItem)} 嗎？`,
										icon: 'warning',
										showCancelButton: true,
										confirmButtonText: '確定',
										cancelButtonText: '取消'
									}).then((result) => {
										if (!result.isConfirmed) {
											return;
										}
										gradingItemApp
											.deleteGradingItem(gradingItem.id)
											.then(() => {
												Swal.fire({
													title: '刪除成功',
													icon: 'success',
													confirmButtonText: '確定'
												});
												gradingItemApp.fetchGradingItems(project.id);
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
{/if}
