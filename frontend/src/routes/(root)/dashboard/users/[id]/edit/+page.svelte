<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { useRolesStorage } from '@/apps/admin/domain/role';
	import { useUserStorage, type UpdateUserPayload } from '@/apps/admin/domain/user';
	import { useUserApp } from '@/apps/admin/application/user';
	import { useUserProfile } from '@/apps/contest/application/user';
	import { refreshProfileIfSelf } from '$lib/profileRefresh';
	import { onMount } from 'svelte';
	import Swal from 'sweetalert2';
	import { Button, Input, Label, MultiSelect } from 'flowbite-svelte';

	const user = $state(useUserStorage().get());
	if (!user && browser) {
		Swal.fire({
			title: '找不到使用者',
			icon: 'error',
			confirmButtonText: '確定'
		}).then(() => {
			goto('/dashboard/users');
		});
	}

	const userApp = useUserApp();
	const userProfile = useUserProfile();
	const roles = useRolesStorage().get();
	const payload: UpdateUserPayload = $state({
		username: user?.username,
		name: user?.name,
		password: '',
		roles: roles?.filter((role) => user?.roles.includes(role.name)).map((role) => role.id)
	});

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
		<span>編輯使用者: {user?.name}</span>
	</h1>

	<form
		id="submit-form"
		class="space-y-6"
		onsubmit={async (e) => {
			e.preventDefault();

			if (payload.password?.trim() === '') {
				delete payload.password;
			}

			userApp.updateUser(user!.id, payload).then(async () => {
				Swal.fire({
					title: '編輯成功',
					icon: 'success',
					showConfirmButton: false,
					timer: 1500
				});
				// 改到的若是自己,角色要立刻反映在導覽列(後台入口、提交鈕)與串流身分上。
				// 自己拿掉了 admin 的話,後台守門已在 invalidate 期間把人送回首頁,
				// 這裡就不要再往後台導。
				const self = await refreshProfileIfSelf(user!.username);
				if (self && !userProfile.isAdmin) return;
				goto(`/dashboard/users/${user!.id}`);
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
			<Label for="password-input">密碼</Label>
			<Input id="password-input" name="password" type="password" bind:value={payload.password} />
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

		<Button color="green" class="cursor-pointer" type="submit">編輯</Button>
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
