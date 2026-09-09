<script lang="ts">
	import { goto } from '$app/navigation';
	import { useProjectStorage } from '@/apps/admin/domain/project';
	import Swal from 'sweetalert2';
	import { Button, Checkbox, Input, Label, MultiSelect, Select } from 'flowbite-svelte';
	import { useGradingItemStorage } from '@/apps/admin/domain/grading-item';
	import {
		useProjectTestcaseGradingItemsStorage,
		type CreateTestcasePayload
	} from '@/apps/admin/domain/testcase';
	import { useTestcaseApp } from '@/apps/admin/application/testcase';

	const projectStorage = useProjectStorage();
	if (projectStorage.error !== null) {
		Swal.fire({
			title: 'Error',
			text: projectStorage.error.message,
			icon: 'error',
			showConfirmButton: false,
			timer: 1500
		});
		goto('/dashboard/projects');
	}

	const project = projectStorage.get()!;
	const gradingItemStorage = useGradingItemStorage();
	if (gradingItemStorage.error !== null) {
		Swal.fire({
			title: 'Error',
			text: gradingItemStorage.error.message,
			icon: 'error',
			showConfirmButton: false,
			timer: 1500
		});
		goto(`/dashboard/projects/${project.id}/grading-items`);
	}

	const gradingItems = useProjectTestcaseGradingItemsStorage().get()!;
	const gradingItem = gradingItemStorage.get()!;

	const formatGradingItemName = (item: { label?: string; name: string }) =>
		item.label ? item.label + ": " + item.name : item.name;

	const testcaseApp = useTestcaseApp();
	let createTestcasePayload: CreateTestcasePayload = $state({
		name: '',
		command: '',
		score: 0,
		isSample: false,
		depends: []
	});
</script>

{#if !project || !gradingItem || !gradingItems}
	<div>Loading...</div>
{:else}
	<div class="p-4">
		<h1
			class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
		>
			<span>新增 {project.name} 的評分項目 {formatGradingItemName(gradingItem)} 的測試案例</span>
		</h1>

		<form
			class="space-y-6"
			onsubmit={(e) => {
				e.preventDefault();
				testcaseApp
					.createTestcase(gradingItem.id, createTestcasePayload)
					.then(() => {
						Swal.fire({
							title: '新增成功',
							icon: 'success',
							confirmButtonText: '確定'
						});
						goto(`/dashboard/projects/${project.id}/grading-items/${gradingItem.id}`);
					})
					.catch((error) => {
						console.error(error);
						Swal.fire({
							title: '新增失敗',
							icon: 'error',
							text: error.message,
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
				<Input required id="name-input" name="name" bind:value={createTestcasePayload.name} />
			</div>

			<div>
				<Label for="command-input">
					<span class="align-middle text-red-500">*</span>
					指令
				</Label>
				<Input
					id="command-input"
					name="command"
					required
					bind:value={createTestcasePayload.command}
				/>
			</div>

			<div>
				<Label for="score-input">
					<span class="align-middle text-red-500">*</span>
					分數
				</Label>
				<Input
					required
					id="score-input"
					name="score"
					type="number"
					bind:value={createTestcasePayload.score}
				/>
			</div>

			<div>
				<Label class="flex items-center gap-x-2" for="is-sample-checkbox"
					>範例測試案例?

					<Checkbox
						id="is-sample-checkbox"
						name="isSample"
						bind:checked={createTestcasePayload.isSample}
					/>
					是
				</Label>
			</div>

			<div>
				<Label for="depends-multi-select">依賴的測試案例</Label>
				<div class="m-4 flex flex-col gap-y-2 font-medium">
					{#each gradingItems as item}
						<div class="flex items-center gap-x-2">
							<input
								id={`grading-item-${item.id}`}
								class="mt-1 self-start rounded-lg border-2 border-gray-300 p-2 disabled:cursor-not-allowed disabled:opacity-50"
								type="checkbox"
								disabled={item.testcases.length === 0}
								name="depends"
								value={item.id}
								checked={item.testcases.some((testcase) =>
									createTestcasePayload.depends.includes(testcase.id)
								) && item.testcases.length !== 0}
								onchange={(e) => {
									if (item.testcases.length === 0) return;
									if (e.currentTarget.checked) {
										createTestcasePayload.depends = [
											...new Set([
												...createTestcasePayload.depends,
												...item.testcases.map((testcase) => testcase.id)
											])
										];
									} else {
										createTestcasePayload.depends = createTestcasePayload.depends.filter(
											(id) => !item.testcases.map((testcase) => testcase.id).includes(id)
										);
									}
								}}
							/>
							<details>
								<summary>{formatGradingItemName(item)} ({item.testcases.length} testcases)</summary>
								<div class="flex flex-col gap-y-2">
									{#if item.testcases.length === 0}
										<div>No testcases</div>
									{:else}
										{#each item.testcases as testcase}
											<label class="flex items-center gap-x-2">
												<input
													class="rounded-lg border-2 border-gray-300 p-2"
													type="checkbox"
													name="depends"
													value={testcase.id}
													checked={createTestcasePayload.depends.includes(testcase.id)}
													onchange={(e) => {
														if (e.currentTarget.checked) {
															createTestcasePayload.depends = [
																...createTestcasePayload.depends,
																testcase.id
															];
														} else {
															createTestcasePayload.depends = createTestcasePayload.depends.filter(
																(id) => id !== testcase.id
															);
														}
													}}
												/>
												{testcase.name}
											</label>
										{/each}
									{/if}
								</div>
							</details>
						</div>
					{/each}
				</div>

				<Button color="green" class="cursor-pointer" type="submit">建立測試案例</Button>
			</div>
		</form>
	</div>
{/if}
