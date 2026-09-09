<script lang="ts">
	import { goto } from '$app/navigation';
	import { useGroupApp } from '@/apps/admin/application/group';
	import { useGroupStorage } from '@/apps/admin/domain/group';
	import { useUsersStorage } from '@/apps/admin/domain/user';
	import { refreshContestProject } from '@/lib/refreshContestState';
	import {
		Button,
		ButtonGroup,
		Input,
		Label,
		MultiSelect,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const groupApp = useGroupApp();
	const groupStorage = useGroupStorage();
	const group = $derived(groupStorage.get()!);
	const users = $derived(useUsersStorage().get()!);

	const members = $derived(users.filter((user) => group.userIds.includes(user.id)));
	const candidates = $derived(users.filter((user) => !group.userIds.includes(user.id)));

	let selectedUserIds = $state<number[]>([]);
	const payload = $state({ name: '', description: '' });

	$effect(() => {
		payload.name = group.name;
		payload.description = group.description;
	});

	const refresh = () => groupApp.fetchGroup(group.id);
	// 群組可能掛在某場比賽底下,成員異動就是參賽資格異動 —— 也包括改到自己。
	// 和競賽頁的 refreshParticipants 一樣無條件重抓目前題目,讓導覽列的提交鈕
	// 不用離開後台就跟上;一次 GET 而已,不值得為了省它去追「自己在不在裡面」。
	const refreshMembers = () => {
		refreshContestProject();
		return refresh();
	};
</script>

{#if !group}
	<div>Loading...</div>
{:else}
	<div class="flex flex-col p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/dashboard/groups"
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
				<p>{group.name}</p>
				<ButtonGroup class="ml-auto">
					<Button
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
							}).then(async (result) => {
								if (!result.isConfirmed) {
									return;
								}
								await groupApp.deleteGroup(group.id).then(() => {
									// 刪掉的群組若掛在目前的比賽下,成員(可能含自己)就不再是參賽者。
									refreshContestProject();
									Swal.fire({
										title: '刪除成功',
										icon: 'success',
										showConfirmButton: false,
										timer: 1500
									});
								});
								goto('/dashboard/groups');
							});
						}}>刪除</Button
					>
				</ButtonGroup>
			</div>
		</h1>

		<form
			class="space-y-6"
			onsubmit={(e) => {
				e.preventDefault();
				groupApp
					.updateGroup(group.id, payload)
					.then(() => {
						Swal.fire({
							title: '更新成功',
							icon: 'success',
							showConfirmButton: false,
							timer: 1500
						});
						refresh();
					})
					.catch((error) => {
						console.error(error);
						Swal.fire({
							title: '更新失敗',
							icon: 'error',
							confirmButtonText: '確定'
						});
					});
			}}
		>
			<div>
				<Label for="name-input">
					<span class="align-middle text-red-500">*</span>
					名稱
				</Label>
				<Input required id="name-input" name="name" bind:value={payload.name} />
			</div>

			<div>
				<Label for="description-input">描述</Label>
				<Input
					id="description-input"
					name="description"
					type="text"
					bind:value={payload.description}
				/>
			</div>

			<Button color="primary" class="cursor-pointer" type="submit">更新</Button>
		</form>

		<hr class="my-4" />

		<h2 class="mb-4 flex items-center gap-x-2 text-2xl font-bold">
			<span>群組成員</span>
		</h2>

		<div class="mb-4 flex items-end gap-2">
			<div class="grow">
				<Label for="member-select">加入使用者</Label>
				<MultiSelect
					id="member-select"
					items={candidates.map((user) => ({
						name: `${user.name} (${user.username})`,
						value: user.id
					}))}
					bind:value={selectedUserIds}
				></MultiSelect>
			</div>
			<Button
				color="green"
				class="cursor-pointer"
				disabled={selectedUserIds.length === 0}
				onclick={() => {
					groupApp
						.addUsersToGroup(group.id, selectedUserIds)
						.then(() => {
							selectedUserIds = [];
							refreshMembers();
						})
						.catch((error) => {
							console.error(error);
							Swal.fire({
								title: '加入失敗',
								icon: 'error',
								confirmButtonText: '確定'
							});
						});
				}}>加入</Button
			>
		</div>

		<Table striped hoverable shadow>
			<TableHead>
				<TableHeadCell>#</TableHeadCell>
				<TableHeadCell>使用者名稱</TableHeadCell>
				<TableHeadCell>名字</TableHeadCell>
				<TableHeadCell>
					<span class="sr-only">Actions</span>
				</TableHeadCell>
			</TableHead>
			<TableBody tableBodyClass="divide-y">
				{#each members as member, index}
					<TableBodyRow>
						<TableBodyCell>{index + 1}</TableBodyCell>
						<TableBodyCell>{member.username}</TableBodyCell>
						<TableBodyCell>{member.name}</TableBodyCell>
						<TableBodyCell>
							<Button
								pill
								color="red"
								class="cursor-pointer"
								onclick={() => {
									groupApp
										.removeUsersFromGroup(group.id, [member.id])
										.then(() => {
											refreshMembers();
										})
										.catch((error) => {
											console.error(error);
											Swal.fire({
												title: '移除失敗',
												icon: 'error',
												confirmButtonText: '確定'
											});
										});
								}}>移除</Button
							>
						</TableBodyCell>
					</TableBodyRow>
				{:else}
					<TableBodyRow>
						<TableBodyCell colspan={4} class="text-center">尚無成員</TableBodyCell>
					</TableBodyRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
