<script lang="ts">
	import { goto } from '$app/navigation';
	import { useGroupApp } from '@/apps/admin/application/group';
	import { Button, Input, Label } from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const groupApp = useGroupApp();
	const payload = $state({ name: '', description: '' });
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>新增群組</span>
	</h1>

	<form
		id="submit-form"
		class="space-y-6"
		onsubmit={async (e) => {
			e.preventDefault();
			groupApp
				.createGroup(payload)
				.then((id) => {
					Swal.fire({
						title: '新增成功',
						icon: 'success',
						showConfirmButton: false,
						timer: 1500
					});
					goto(`/dashboard/groups/${id}`);
				})
				.catch((error) => {
					console.error(error);
					Swal.fire({
						title: '新增失敗',
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

		<Button color="green" class="cursor-pointer" type="submit">新增</Button>
	</form>
</div>
