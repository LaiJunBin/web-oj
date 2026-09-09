<script lang="ts">
	import { goto } from '$app/navigation';
	import { useProjectStorage } from '@/apps/admin/domain/project';
	import Swal from 'sweetalert2';
	import { Button, Input, Label, Select, Tooltip } from 'flowbite-svelte';
	import { useGradingItemApp } from '@/apps/admin/application/grading-item';
	import {
		OutputFormatEnum,
		ScoringStrategyEnum,
		useGradingItemStorage,
		type OutputFormat,
		type ScoringStrategy,
		type UpdateGradingItemPayload
	} from '@/apps/admin/domain/grading-item';
	import { useProjectTestcaseGradingItemsStorage } from '@/apps/admin/domain/testcase';

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

	const gradingItemApp = useGradingItemApp();
	const project = projectStorage.get()!;
	const gradingItem = useGradingItemStorage().get()!;
	let updateGradingItemPayload: UpdateGradingItemPayload = $state({
		name: gradingItem.name,
		label: gradingItem.label,
		description: gradingItem.description,
		score: gradingItem.score,
		scoringStrategy: gradingItem.scoringStrategy as ScoringStrategy,
		outputFormat: gradingItem.outputFormat as OutputFormat,
		resultParser: gradingItem.resultParser,
		depends: gradingItem.depends.map((testcase) => testcase.id)
	});
	const gradingItems = useProjectTestcaseGradingItemsStorage().get()!;

	const formatGradingItemName = (item: { label?: string; name: string }) =>
		item.label ? item.label + ": " + item.name : item.name;
</script>

{#if !project || !gradingItems}
	<div>Loading...</div>
{:else}
	<div class="p-4">
		<h1
			class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
		>
			<span>編輯 {project.name} 的評分項目</span>
		</h1>

		<form
			class="space-y-6"
			onsubmit={(e) => {
				e.preventDefault();
				gradingItemApp
					.updateGradingItem(gradingItem.id, updateGradingItemPayload)
					.then(() => {
						Swal.fire({
							title: '編輯成功',
							icon: 'success',
							confirmButtonText: '確定'
						});
						goto(`/dashboard/projects/${project.id}/grading-items/${gradingItem.id}`);
					})
					.catch((error) => {
						console.error(error);
						Swal.fire({
							title: '編輯失敗',
							icon: 'error',
							text: error.message,
							confirmButtonText: '確定'
						});
					});
			}}
		>
			<div>
				<Label for="label-input">
					<span class="align-middle text-red-500">*</span>
					項次
				</Label>
				<Input required id="label-input" name="label" bind:value={updateGradingItemPayload.label} />
			</div>

			<div>
				<Label for="name-input">
					<span class="align-middle text-red-500">*</span>
					名稱
				</Label>
				<Input required id="name-input" name="name" bind:value={updateGradingItemPayload.name} />
			</div>

			<div>
				<Label for="description-input">描述訊息</Label>
				<Input
					id="description-input"
					name="description"
					bind:value={updateGradingItemPayload.description}
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
					bind:value={updateGradingItemPayload.score}
				/>
			</div>

			<div>
				<Label for="scoring-strategy-select">
					<span class="align-middle text-red-500">*</span>
					評分策略
				</Label>
				<Select
					required
					id="scoring-strategy-select"
					name="scoringStrategy"
					placeholder=""
					bind:value={updateGradingItemPayload.scoringStrategy}
				>
					<option value={ScoringStrategyEnum.decrease}
						>扣分制: 當測試用例失敗時，從滿分扣除分數</option
					>
					<option value={ScoringStrategyEnum.increase}
						>加分制: 當測試用例通過時，從 0 分開始加分直到滿分。</option
					>
				</Select>
			</div>

			<div>
				<Label for="output-format-select">
					<span class="align-middle text-red-500">*</span>
					輸出顯示格式
				</Label>
				<Select
					required
					id="output-format-select"
					name="outputFormat"
					placeholder=""
					items={Object.values(OutputFormatEnum).map((value) => ({
						name: value,
						value
					}))}
					bind:value={updateGradingItemPayload.outputFormat}
				></Select>
			</div>

			<div>
				<Label for="result-parser-select" class="inline-flex items-center gap-2">
					測試結果解析方式
					<span class="cursor-help text-gray-500" aria-label="測試結果解析方式說明">ⓘ</span>
					<Tooltip class="pointer-events-none max-w-md">
						<div class="space-y-1 text-sm leading-relaxed">
							<p>可覆蓋專案設定，決定此評分項目的測試結果如何判分。</p>
							<p>建議使用繼承自專案。</p>
							<p>繼承自專案：沿用專案設定。</p>
							<p>auto：先嘗試用 stdout 規則解析，失敗時改用 exitcode。</p>
							<p>exitcode：exit code 為 0 視為通過，非 0 視為失敗。</p>
							<p>stdout：讀取輸出格式 pass [score] 或 fail [score]。</p>
						</div>
					</Tooltip>
				</Label>
				<Select id="result-parser-select" name="resultParser" bind:value={updateGradingItemPayload.resultParser}>
					<option value="">繼承自專案</option>
					<option value="auto">auto（自動）</option>
					<option value="exitcode">exitcode（退出碼）</option>
					<option value="stdout">stdout（輸出解析）</option>
				</Select>
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
								disabled={item.testcases.length === 0 || item.id === gradingItem.id}
								name="depends"
								value={item.id}
								checked={item.testcases.some((testcase) =>
									updateGradingItemPayload.depends!.includes(testcase.id)
								) && item.testcases.length !== 0}
								onchange={(e) => {
									if (item.testcases.length === 0) return;
									if (e.currentTarget.checked) {
										updateGradingItemPayload.depends = [
											...new Set([
												...updateGradingItemPayload.depends!,
												...item.testcases.map((testcase) => testcase.id)
											])
										];
									} else {
										updateGradingItemPayload.depends = updateGradingItemPayload.depends!.filter(
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
													class="rounded-lg border-2 border-gray-300 p-2 disabled:cursor-not-allowed disabled:opacity-50"
													type="checkbox"
													name="depends"
													value={testcase.id}
													checked={updateGradingItemPayload.depends!.includes(testcase.id)}
													disabled={item.id === gradingItem.id}
													onchange={(e) => {
														if (e.currentTarget.checked) {
															updateGradingItemPayload.depends = [
																...updateGradingItemPayload.depends!,
																testcase.id
															];
														} else {
															updateGradingItemPayload.depends =
																updateGradingItemPayload.depends!.filter(
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

				<Button color="green" class="cursor-pointer" type="submit">編輯評分項目</Button>
			</div>
		</form>
	</div>
{/if}
