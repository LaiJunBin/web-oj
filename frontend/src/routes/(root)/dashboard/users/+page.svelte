<script lang="ts">
	import { goto } from '$app/navigation';
	import { useUserApp } from '@/apps/admin/application/user';
	import { isSelf, leaveAfterSelfDelete } from '$lib/profileRefresh';
	import { useUsersStorage } from '@/apps/admin/domain/user';
	import {
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
	import { escapeHtml, sanitizeHtml } from '$lib/sanitize';

	const userApp = useUserApp();
	const users = $derived(useUsersStorage().get()!);

</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>使用者管理</span>
		<Button href="/dashboard/users/create" color="green">新增</Button>
		<Button
			class="cursor-pointer"
			color="blue"
			onclick={() => {
				const fileInput = document.createElement('input');
				fileInput.type = 'file';
				fileInput.accept = '.json';
				fileInput.onchange = async (event) => {
					const input = event.target as HTMLInputElement;
					const file = input.files?.[0];
					if (!file) {
						return;
					}

					const text = await file.text();
					Swal.fire({
						title: '匯入的資料',
						html: `<pre style="text-align: left; max-height: 50vh;">${escapeHtml(text)}</pre>`,
						confirmButtonText: '確定',
						showCancelButton: true,
						cancelButtonText: '取消'
					}).then((result) => {
						if (!result.isConfirmed) {
							return;
						}

						userApp
							.importUsers(file)
							.then((total) => {
								Swal.fire({
									title: '匯入成功',
									text: `成功匯入 ${total} 位使用者`,
									icon: 'success',
									confirmButtonText: '確定'
								});
								userApp.fetchUsers();
							})
							.catch((error) => {
								console.error(error);
								Swal.fire({
									title: '匯入失敗',
									html: `錯誤訊息: <pre>${sanitizeHtml(String(error))}</pre>`,
									icon: 'error',
									confirmButtonText: '確定'
								});
							});
					});
				};
				fileInput.click();
			}}>匯入</Button
		>
	</h1>

	<Table striped hoverable shadow>
		<TableHead>
			<TableHeadCell>#</TableHeadCell>
			<TableHeadCell>使用者名稱</TableHeadCell>
			<TableHeadCell>名字</TableHeadCell>
			<TableHeadCell>角色</TableHeadCell>
			<TableHeadCell>
				<span class="sr-only">Actions</span>
			</TableHeadCell>
		</TableHead>
		<TableBody tableBodyClass="divide-y">
			{#each users as user, index}
				<TableBodyRow
					class="cursor-pointer"
					onclick={(e) => {
						const target = e.target as HTMLElement;
						if (target.tagName === 'TD') {
							goto(`/dashboard/users/${user.id}`);
						}
					}}
				>
					<TableBodyCell>{index + 1}</TableBodyCell>
					<TableBodyCell>{user.username}</TableBodyCell>
					<TableBodyCell>{user.name}</TableBodyCell>
					<TableBodyCell>{user.roles.join(', ')}</TableBodyCell>
					<TableBodyCell>
						<ButtonGroup>
							<Button pill color="primary" href="/dashboard/users/{user.id}/edit">編輯</Button>
							<Button
								pill
								color="red"
								class="cursor-pointer"
								onclick={() => {
									Swal.fire({
										title: '確定要刪除嗎？',
										text: `確定要刪除使用者 ${user.username} 嗎？`,
										icon: 'warning',
										showCancelButton: true,
										confirmButtonText: '確定',
										cancelButtonText: '取消'
									}).then((result) => {
										if (!result.isConfirmed) {
											return;
										}
										userApp
											.deleteUser(user.id)
											.then(() => {
												// 刪的是自己:這個 session 已經死了,重抓清單只會 401。
												if (isSelf(user.username)) return leaveAfterSelfDelete();
												Swal.fire({
													title: '刪除成功',
													icon: 'success',
													confirmButtonText: '確定'
												});
												userApp.fetchUsers();
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
						</ButtonGroup>
					</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>
</div>
