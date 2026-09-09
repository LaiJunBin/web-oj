<script lang="ts">
	import { goto } from '$app/navigation';
	import { useImageApp } from '@/apps/admin/application/image';
	import { useImagesStorage } from '@/apps/admin/domain/image';
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

	const imageApp = useImageApp();
	const imagesStorage = useImagesStorage();
	const images = $derived(imagesStorage.get()!);
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>評分環境管理</span>
		<Button href="/dashboard/images/create" color="green">新增</Button>
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
					imageApp
						.importImage(file)
						.then((result) => {
							Swal.fire({ icon: 'success', title: '匯入成功', text: ((r) => `已建立映像「${r.tag}」` + (r.buildQueued ? ",已排入建置" : ""))(result) });
							imageApp.fetchImages();
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
			<TableHeadCell>標籤</TableHeadCell>
			<TableHeadCell>描述訊息</TableHeadCell>
			<TableHeadCell>是否可用?</TableHeadCell>
			<TableHeadCell>
				<span class="sr-only">Actions</span>
			</TableHeadCell>
		</TableHead>
		<TableBody tableBodyClass="divide-y">
			{#each images as image, index}
				<TableBodyRow
					class="cursor-pointer"
					onclick={(e) => {
						const target = e.target as HTMLElement;
						if (target.tagName === 'TD') {
							goto(`/dashboard/images/${image.tag}`);
						}
					}}
				>
					<TableBodyCell>{index + 1}</TableBodyCell>
					<TableBodyCell>{image.tag}</TableBodyCell>
					<TableBodyCell>
						<button
							class="cursor-pointer"
							onclick={(e) => {
								e.stopPropagation();
								Swal.fire({
									title: '描述訊息',
									input: 'text',
									inputValue: image.description,
									showCancelButton: true,
									confirmButtonText: '確定',
									cancelButtonText: '取消'
								}).then((result) => {
									if (!result.isConfirmed) {
										return;
									}
									const description = result.value;
									imageApp
										.updateImage(image.tag, { description })
										.then(() => {
											Swal.fire({
												title: '更新成功',
												icon: 'success',
												confirmButtonText: '確定'
											});
											imageApp.fetchImages();
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
							{image.description}
						</button>
					</TableBodyCell>
					<TableBodyCell>{image.isReady ? '是' : '否'}</TableBodyCell>
					<TableBodyCell>
						<Button
							pill
							color="red"
							class="cursor-pointer"
							onclick={() => {
								Swal.fire({
									title: '確定要刪除嗎？',
									text: `確定要刪除 ${image.tag} 嗎？`,
									icon: 'warning',
									showCancelButton: true,
									confirmButtonText: '確定',
									cancelButtonText: '取消'
								}).then((result) => {
									if (!result.isConfirmed) {
										return;
									}
									imageApp
										.deleteImage(image.tag)
										.then(() => {
											Swal.fire({
												title: '刪除成功',
												icon: 'success',
												confirmButtonText: '確定'
											});
											imageApp.fetchImages();
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
