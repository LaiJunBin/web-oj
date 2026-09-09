<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { refreshContestState } from '@/lib/refreshContestState';
	import { useProjectApp } from '@/apps/admin/application/project';
	import { useImagesStorage } from '@/apps/admin/domain/image';
	import { useProjectStorage, type UpdateProjectPayload } from '@/apps/admin/domain/project';
	import {
		Button,
		Tooltip,
		Input,
		Label,
		Select,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell,
		Textarea
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const projectApp = useProjectApp();
	const project = useProjectStorage().get()!;
	const images = useImagesStorage().get()!;
	const payload: UpdateProjectPayload = $state({
		name: project.name,
		description: project.description,
		projectImageTag: project.image?.tag,
		projectEnv: project.env,
		testImageTag: project.testImage?.tag,
		testEnv: project.testEnv,
		resultParser: project.resultParser,
		dynamicFields: project.dynamicFields.slice()
	});
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>編輯專案 {project.name}</span>
	</h1>

	<form
		id="submit-form"
		class="space-y-6"
		onsubmit={async (e) => {
			e.preventDefault();
			projectApp.updateProject(project.id, payload).then(() => {
				refreshContestState();
				Swal.fire({
					title: '編輯成功',
					icon: 'success',
					showConfirmButton: false,
					timer: 1500
				});
				invalidateAll().then(() => {
					goto(`/dashboard/projects/${project.id}`);
				});
			});
		}}
	>
		<div>
			<Label for="name-input">
				<span class="align-middle text-red-500">*</span>
				專案名稱
			</Label>
			<Input required id="name-input" name="name" bind:value={payload.name} />
		</div>

		<div>
			<Label for="description-input">描述訊息</Label>
			<Input
				id="description-input"
				name="description"
				type="text"
				bind:value={payload.description}
			/>
		</div>

		<div>
			<Label for="project-image-tag-select">
				<span class="align-middle text-red-500">*</span>
				專案執行環境
			</Label>
			<Select
				required
				id="project-image-tag-select"
				name="projectImageTag"
				bind:value={payload.projectImageTag}
			>
				{#each images as image}
					<option value={image.tag} disabled={!image.isReady}>{image.tag}</option>
				{/each}
			</Select>
		</div>

		<div>
			<Label for="project-env-input">專案執行環境變數</Label>
			<Textarea id="project-env-input" name="projectEnv" bind:value={payload.projectEnv} />
		</div>

		<div>
			<Label for="test-image-tag-select">
				<span class="align-middle text-red-500">*</span>
				測試執行環境
			</Label>
			<Select
				required
				id="test-image-tag-select"
				name="testImageTag"
				bind:value={payload.testImageTag}
			>
				{#each images as image}
					<option value={image.tag} disabled={!image.isReady}>{image.tag}</option>
				{/each}
			</Select>
		</div>

		<div>
			<Label for="test-env-input">測試執行環境變數</Label>
			<Textarea id="test-env-input" name="testEnv" bind:value={payload.testEnv} />
		</div>

		<div>
			<Label for="result-parser-select" class="inline-flex items-center gap-2">
				測試結果解析方式
				<span class="cursor-help text-gray-500" aria-label="測試結果解析方式說明">ⓘ</span>
				<Tooltip class="pointer-events-none max-w-md">
					<div class="space-y-1 text-sm leading-relaxed">
						<p>決定系統如何把測試結果轉成通過/失敗與分數。</p>
						<p>建議使用 auto。</p>
						<p>auto：先嘗試用 stdout 規則解析，失敗時改用 exitcode。</p>
						<p>exitcode：exit code 為 0 視為通過，非 0 視為失敗。</p>
						<p>stdout：讀取輸出格式 pass [score] 或 fail [score]。</p>
					</div>
				</Tooltip>
			</Label>
			<Select id="result-parser-select" name="resultParser" bind:value={payload.resultParser}>
				<option value="auto">auto（自動）</option>
				<option value="exitcode">exitcode（退出碼）</option>
				<option value="stdout">stdout（輸出解析）</option>
			</Select>
		</div>

		<dir>
			<h2 class="text-lg font-bold text-gray-900 dark:text-white">動態欄位</h2>
			<Table>
				<TableHead>
					<TableHeadCell>Label</TableHeadCell>
					<TableHeadCell>Name</TableHeadCell>
					<TableHeadCell>Type</TableHeadCell>
					<TableHeadCell>Default Value</TableHeadCell>
					<TableHeadCell>Is Required</TableHeadCell>
					<TableHeadCell>
						<span class="sr-only">Actions</span>
					</TableHeadCell>
				</TableHead>
				<TableBody tableBodyClass="divide-y">
					{#each payload.dynamicFields! as field, index}
						<TableBodyRow>
							<TableBodyCell>
								<Input
									id={`dynamic-field-label-${index}`}
									name={`dynamicFields[${index}].label`}
									bind:value={field.label}
									required
								/>
							</TableBodyCell>
							<TableBodyCell>
								<Input
									id={`dynamic-field-name-${index}`}
									name={`dynamicFields[${index}].name`}
									bind:value={field.name}
									required
								/>
							</TableBodyCell>
							<TableBodyCell>
								<Select
									id={`dynamic-field-type-${index}`}
									name={`dynamicFields[${index}].type`}
									bind:value={field.type}
									onchange={() => {
										field.defaultValue = undefined;
										if (field.type === 'boolean') {
											field.defaultValue = 'true';
										}
									}}
								>
									<option value="text">Text</option>
									<option value="boolean">Boolean</option>
									<option value="file">File</option>
								</Select>
							</TableBodyCell>
							<TableBodyCell>
								{#if field.type === 'text'}
									<Input
										id={`dynamic-field-default-value-${index}`}
										name={`dynamicFields[${index}].defaultValue`}
										bind:value={field.defaultValue}
									/>
								{:else if field.type === 'boolean'}
									<Select
										id={`dynamic-field-default-value-${index}`}
										name={`dynamicFields[${index}].defaultValue`}
										placeholder=""
										required
										bind:value={field.defaultValue}
									>
										<option value="true">True</option>
										<option value="false">False</option>
									</Select>
								{:else if field.type === 'file'}
									x
								{/if}
							</TableBodyCell>
							<TableBodyCell>
								<input
									id={`dynamic-field-is-required-${index}`}
									name={`dynamicFields[${index}].isRequired`}
									type="checkbox"
									bind:checked={field.isRequired}
								/>
							</TableBodyCell>
							<TableBodyCell>
								<Button
									color="red"
									class="cursor-pointer"
									on:click={() => {
										payload.dynamicFields!.splice(index, 1);
									}}
								>
									刪除
								</Button>
							</TableBodyCell>
						</TableBodyRow>
					{/each}
					<TableBodyRow>
						<TableBodyCell colspan={6}>
							<div class="flex justify-center">
								<Button
									color="green"
									class="cursor-pointer"
									on:click={() => {
										payload.dynamicFields!.push({
											label: '',
											name: '',
											type: 'text',
											defaultValue: undefined,
											isRequired: false
										});
									}}
								>
									新增動態欄位
								</Button>
							</div>
						</TableBodyCell>
					</TableBodyRow>
				</TableBody>
			</Table>
		</dir>

		<Button color="green" class="cursor-pointer" type="submit">編輯專案</Button>
	</form>
</div>
