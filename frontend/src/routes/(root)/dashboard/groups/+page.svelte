<script lang="ts">
	import { goto } from '$app/navigation';
	import { useGroupApp } from '@/apps/admin/application/group';
	import { useGroupsStorage } from '@/apps/admin/domain/group';
	import { refreshContestProject } from '@/lib/refreshContestState';
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

	const groupApp = useGroupApp();
	const groups = $derived(useGroupsStorage().get()!);
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>群組管理</span>
		<Button href="/dashboard/groups/create" color="green">新增</Button>
	</h1>

	<Table striped hoverable shadow>
		<TableHead>
			<TableHeadCell>#</TableHeadCell>
			<TableHeadCell>名稱</TableHeadCell>
			<TableHeadCell>描述</TableHeadCell>
			<TableHeadCell>
				<span class="sr-only">Actions</span>
			</TableHeadCell>
		</TableHead>
		<TableBody tableBodyClass="divide-y">
			{#each groups as group, index}
				<TableBodyRow
					class="cursor-pointer"
					onclick={(e) => {
						const target = e.target as HTMLElement;
						if (target.tagName === 'TD') {
							goto(`/dashboard/groups/${group.id}`);
						}
					}}
				>
					<TableBodyCell>{index + 1}</TableBodyCell>
					<TableBodyCell>{group.name}</TableBodyCell>
					<TableBodyCell>{group.description || '無'}</TableBodyCell>
					<TableBodyCell>
						<ButtonGroup>
							<Button pill color="primary" href="/dashboard/groups/{group.id}">管理</Button>
							<Button
								pill
								color="red"
								class="cursor-pointer"
								onclick={() => {
									Swal.fire({
										title: '確定要刪除嗎？',
										text: `確定要刪除群組 ${group.name} 嗎？刪除後群組成員將失去以此群組取得的參賽資格。`,
										icon: 'warning',
										showCancelButton: true,
										confirmButtonText: '確定',
										cancelButtonText: '取消'
									}).then((result) => {
										if (!result.isConfirmed) {
											return;
										}
										groupApp
											.deleteGroup(group.id)
											.then(() => {
												Swal.fire({
													title: '刪除成功',
													icon: 'success',
													confirmButtonText: '確定'
												});
												groupApp.fetchGroups();
												// 刪掉的群組若掛在目前的比賽下,成員(可能含自己)就不再是參賽者。
												refreshContestProject();
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
			{:else}
				<TableBodyRow>
					<TableBodyCell colspan={4} class="text-center">尚未建立任何群組</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>
</div>
