<script lang="ts">
	import { useImageApp } from '@/apps/admin/application/image';
	import { goto } from '$app/navigation';
	import Swal from 'sweetalert2';
	import { Button, Checkbox, Fileupload, Input, Label } from 'flowbite-svelte';
	import { useImageStorage, type UpdateImagePayload } from '@/apps/admin/domain/image';

	const imageApp = useImageApp();
	const image = useImageStorage().get()!;
	const payload: UpdateImagePayload = $state({
		mountPath: image?.mountPath,
		description: image?.description,
		noCache: false
	});
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>編輯執行環境</span>
	</h1>

	<form
		id="submit-form"
		class="space-y-6"
		onsubmit={async (e) => {
			e.preventDefault();
			interface SubmitProjectFormElement extends HTMLFormControlsCollection {
				project: HTMLInputElement;
			}
			const { project } = e.currentTarget.elements as unknown as SubmitProjectFormElement;
			const file = project.files?.[0];

			if (file) {
				payload.project = file;
			}
			imageApp
				.updateImage(image.tag, payload)
				.then(() => {
					Swal.fire({
						title: '編輯成功',
						icon: 'success',
						showConfirmButton: false,
						timer: 1500
					});
					goto(`/dashboard/images/${image.tag}`);
				})
				.catch((error) => {
					console.error(error);
					Swal.fire({
						title: '編輯失敗',
						text: error.message,
						icon: 'error'
					});
				});
		}}
	>
		<div>
			<Label for="description-input">描述訊息</Label>
			<Input id="description-input" name="description" bind:value={payload.description} />
		</div>

		<div>
			<Label for="mount-path-input">
				<span class="align-middle text-red-500">*</span>
				掛載路徑
			</Label>
			<Input required id="mount-path-input" name="mount-path" bind:value={payload.mountPath} />
		</div>

		<div>
			<Label for="project-file-input">專案壓縮檔(.zip)</Label>
			<Fileupload id="project-file-input" name="project" accept=".zip" />
		</div>

		<div class="flex items-center">
			<Checkbox
				id="no-cache-checkbox"
				name="no-cache-checkbox"
				class="mr-2 ml-1"
				bind:checked={payload.noCache}
			/>
			<Label for="no-cache-checkbox">建置時停用快取</Label>
		</div>

		<Button color="green" class="cursor-pointer" type="submit">編輯</Button>
	</form>
</div>
