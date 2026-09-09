<script lang="ts">
	import { goto } from '$app/navigation';
	import { useUserApp } from '@/apps/admin/application/user';
	import { useRolesStorage } from '@/apps/admin/domain/role';
	import { Button, Input, Label, MultiSelect, Select } from 'flowbite-svelte';
	import { onMount } from 'svelte';
	import Swal from 'sweetalert2';

	const userApp = useUserApp();
	const roles = useRolesStorage().get();
	const payload = $state({ username: '', name: '', password: '', roles: [] });

	onMount(() => {
		const roleSelect = document.getElementById('role-select');
		if (roleSelect) {
			roleSelect.removeAttribute('hidden');
		}
	});
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>新增使用者</span>
	</h1>

	<form
		id="submit-form"
		class="space-y-6"
		onsubmit={async (e) => {
			e.preventDefault();
			userApp.createUser(payload).then((id) => {
				Swal.fire({
					title: '新增成功',
					icon: 'success',
					showConfirmButton: false,
					timer: 1500
				});
				goto(`/dashboard/users/${id}`);
			});
		}}
	>
		<div>
			<Label for="username-input">
				<span class="align-middle text-red-500">*</span>
				使用者名稱
			</Label>
			<Input required id="username-input" name="username" bind:value={payload.username} />
		</div>

		<div>
			<Label for="name-input">
				<span class="align-middle text-red-500">*</span>
				名字
			</Label>
			<Input required id="name-input" name="name" type="text" bind:value={payload.name} />
		</div>

		<div>
			<Label for="password-input">
				<span class="align-middle text-red-500">*</span>
				密碼
			</Label>
			<Input
				required
				id="password-input"
				name="password"
				type="password"
				bind:value={payload.password}
			/>
		</div>

		<div id="role-group">
			<Label for="role-select">
				<span class="align-middle text-red-500">*</span>
				角色
			</Label>

			<MultiSelect
				id="role-select"
				name="role"
				required
				items={roles?.map((role) => ({ name: role.name, value: role.id }))}
				bind:value={payload.roles}
			></MultiSelect>
		</div>

		<Button color="green" class="cursor-pointer" type="submit">新增</Button>
	</form>
</div>

<style>
	:global(#role-select) {
		position: absolute;
		opacity: 0;
		height: 0;
		z-index: -1;
	}
</style>
