<script lang="ts">
	import { goto } from '$app/navigation';
	import { refreshContestState } from '@/lib/refreshContestState';
	import { useProjectApp } from '@/apps/admin/application/project';
	import { useProjectsStorage } from '@/apps/admin/domain/project';
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

	const projectApp = useProjectApp();
	const projectsStorage = useProjectsStorage();
	const projects = $derived(projectsStorage.get()!);
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>專案管理</span>
		<Button href="/dashboard/projects/create" color="green">新增</Button>
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
					projectApp
						.importProject(file)
						.then((result) => {
							Swal.fire({ icon: 'success', title: '匯入成功', text: ((r) => `已建立題目「${r.name}」` + (r.missingImageTags.length ? `,但缺少映像:${r.missingImageTags.join("、")}` : ""))(result) });
							projectApp.fetchProjects();
							// 其他題目異動都會刷新 contest 端;匯入是同一類事。
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
			<TableHeadCell>專案執行環境</TableHeadCell>
			<TableHeadCell>測試執行環境</TableHeadCell>
			<TableHeadCell>
				<span class="sr-only">Actions</span>
			</TableHeadCell>
		</TableHead>
		<TableBody tableBodyClass="divide-y">
			{#each projects as project, index}
				<TableBodyRow
					class="cursor-pointer"
					onclick={(e) => {
						const target = e.target as HTMLElement;
						if (target.tagName === 'TD') {
							goto(`/dashboard/projects/${project.id}`);
						}
					}}
				>
					<TableBodyCell>{index + 1}</TableBodyCell>
					<TableBodyCell>{project.name}</TableBodyCell>
					<TableBodyCell>
						<button
							class="cursor-pointer"
							onclick={(e) => {
								e.stopPropagation();
								Swal.fire({
									title: '描述訊息',
									input: 'text',
									inputValue: project.description,
									showCancelButton: true,
									confirmButtonText: '確定',
									cancelButtonText: '取消'
								}).then((result) => {
									if (!result.isConfirmed) {
										return;
									}
									const description = result.value;
									projectApp
										.updateProject(project.id, { description })
										.then(() => {
											refreshContestState();
											Swal.fire({
												title: '更新成功',
												icon: 'success',
												confirmButtonText: '確定'
											});
											projectApp.fetchProjects();
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
							{project.description}
						</button>
					</TableBodyCell>
					<TableBodyCell>
						<a href={`/dashboard/images/${project.image}`} class="border-b">{project.image}</a>
					</TableBodyCell>
					<TableBodyCell>
						<a href={`/dashboard/images/${project.testImage}`} class="border-b"
							>{project.testImage}</a
						>
					</TableBodyCell>
					<TableBodyCell>
						<Button
							pill
							color="red"
							class="cursor-pointer"
							onclick={() => {
								Swal.fire({
									title: '確定要刪除嗎？',
									text: `確定要刪除 ${project.name} 嗎？`,
									icon: 'warning',
									showCancelButton: true,
									confirmButtonText: '確定',
									cancelButtonText: '取消'
								}).then((result) => {
									if (!result.isConfirmed) {
										return;
									}
									projectApp
										.deleteProject(project.id)
										.then(() => {
											refreshContestState();
											Swal.fire({
												title: '刪除成功',
												icon: 'success',
												confirmButtonText: '確定'
											});
											projectApp.fetchProjects();
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
