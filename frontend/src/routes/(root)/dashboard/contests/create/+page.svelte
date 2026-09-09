<script lang="ts">
	import { goto } from '$app/navigation';
	import { optionalNumber, toDatetimeLocal } from '@/lib';
	import { refreshContestState } from '@/lib/refreshContestState';
	import { useContestApp } from '@/apps/admin/application/contest';
	import type { CreateContestPayload } from '@/apps/admin/domain/contest';
	import {
		useJudgePriorityStrategiesStorage,
		useSystemDefaultsStorage
	} from '@/apps/admin/domain/judge-worker';
	import { useProjectsStorage } from '@/apps/admin/domain/project';
	import {
		Button,
		Checkbox,
		Input,
		Label,
		Select,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const contestApp = useContestApp();
	const projects = useProjectsStorage().get()!;
	const judgePriorityStrategies = useJudgePriorityStrategiesStorage().get()!;

	// 欄位留空時套用的伺服器設定值。取不到就退回空泛的「預設」——顯示不了實際
	// 數字只是少了提示,不該讓整個表單開不起來。
	const systemDefaults = useSystemDefaultsStorage().get();

	// 這些欄位的系統預設值裡,任何非正數都代表「不限制」——0 是設定檔寫明的關閉
	// 方式,負數則是它記載的另一種明確關閉寫法(見 backend config 的 judge 區段)。
	// 把它們照數字印出來對管理員沒有意義,所以一律翻成語意。
	const defaultText = (value: number | undefined, unit = '') =>
		value === undefined ? null : value <= 0 ? '不限制' : `${value}${unit}`;
	const withDefault = (text: string | null) => (text ? `預設(${text})` : '預設');

	const cooldownDefaultText = defaultText(systemDefaults?.submissionCooldownSeconds, ' 秒');
	const judgeDurationPlaceholder = withDefault(
		defaultText(systemDefaults?.maxJudgeDurationMinutes, ' 分鐘')
	);
	const cooldownPlaceholder = withDefault(cooldownDefaultText);
	const pidsPlaceholder = withDefault(defaultText(systemDefaults?.pidsLimit));
	const memoryPlaceholder = withDefault(defaultText(systemDefaults?.memoryLimitMB, ' MB'));
	const payload: CreateContestPayload = $state({
		name: '',
		description: '',
		activateTime: toDatetimeLocal(),
		deactivateTime: toDatetimeLocal(),
		isPublic: false,
		projects: []
	});
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>新增競賽</span>
	</h1>

	<form
		id="submit-form"
		class="space-y-6"
		onsubmit={(e) => {
			e.preventDefault();
			// 送出前正規化成 number | undefined（undefined = 使用系統全域預設）
			payload.projects = payload.projects.map((p) => ({
				...p,
				maxJudgeDurationMinutes: optionalNumber(p.maxJudgeDurationMinutes),
				submissionCooldownSeconds: optionalNumber(p.submissionCooldownSeconds),
				pidsLimit: optionalNumber(p.pidsLimit),
				memoryLimitMB: optionalNumber(p.memoryLimitMB)
			}));
			contestApp
				.createContest(payload)
				.then((id) => {
					refreshContestState();
					Swal.fire({
						title: '新增成功',
						icon: 'success',
						showConfirmButton: false,
						timer: 1500
					});
					goto(`/dashboard/contests/${id}`);
				})
				.catch((error) => {
					Swal.fire({
						title: '新增失敗',
						icon: 'error',
						text: error.message
					});
				});
		}}
	>
		<div>
			<Label for="name-input">
				<span class="align-middle text-red-500">*</span>
				競賽名稱
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
			<Label for="activate-time-input">
				<span class="align-middle text-red-500">*</span>
				啟用時間
			</Label>
			<Input
				required
				id="activate-time-input"
				name="activate-time"
				type="datetime-local"
				bind:value={payload.activateTime}
			/>
		</div>

		<div>
			<Label for="deactivate-time-input">
				<span class="align-middle text-red-500">*</span>
				結束時間
			</Label>
			<Input
				required
				id="deactivate-time-input"
				name="deactivate-time"
				type="datetime-local"
				bind:value={payload.deactivateTime}
			/>
		</div>

		<div class="flex items-center gap-x-2">
			<Checkbox id="is-public-checkbox" bind:checked={payload.isPublic} />
			<Label for="is-public-checkbox">公開?</Label>
		</div>

		<dir>
			<h2 class="text-lg font-bold text-gray-900 dark:text-white">專案</h2>
			<Table divClass="relative h-full overflow-auto">
				<TableHead>
					<TableHeadCell>#</TableHeadCell>
					<TableHeadCell class="min-w-60">選擇專案/競賽專案名稱</TableHeadCell>
					<TableHeadCell>專案開啟/關閉時間</TableHeadCell>
					<TableHeadCell>開始/結束時間</TableHeadCell>
					<TableHeadCell>公開?</TableHeadCell>
					<TableHeadCell class="min-w-60">評分優先權策略</TableHeadCell>
					<TableHeadCell class="min-w-36">評測資源上限</TableHeadCell>
					<TableHeadCell class="min-w-28">提交冷卻(秒)</TableHeadCell>
					<TableHeadCell>計分板</TableHeadCell>
					<TableHeadCell>
						<span class="sr-only">Actions</span>
					</TableHeadCell>
				</TableHead>
				<TableBody tableBodyClass="divide-y">
					{#each payload.projects as project, index}
						<TableBodyRow>
							<TableBodyCell>
								{index + 1}
							</TableBodyCell>
							<TableBodyCell class="space-y-2">
								<Select
									bind:value={project.projectId}
									required
									placeholder=""
									items={projects.map((project) => ({ name: project.name, value: project.id }))}
									onchange={() => {
										const newProject = projects.find((p) => p.id === project.projectId)!;
										project.name = newProject.name;
									}}
								></Select>
								<Input bind:value={project.name} required />
							</TableBodyCell>
							<TableBodyCell class="space-y-2">
								<div>
									<Input
										required
										name="activate-time-{index}"
										type="datetime-local"
										bind:value={project.activateTime}
									/>
								</div>
								<div>
									<Input
										required
										name="deactivate-time-{index}"
										type="datetime-local"
										bind:value={project.deactivateTime}
									/>
								</div>
							</TableBodyCell>
							<TableBodyCell class="space-y-2">
								<div>
									<Input
										required
										name="start-time-{index}"
										type="datetime-local"
										bind:value={project.startTime}
									/>
								</div>
								<div>
									<Input
										required
										name="end-time-{index}"
										type="datetime-local"
										bind:value={project.endTime}
									/>
								</div>
							</TableBodyCell>
							<TableBodyCell>
								<label class="flex items-center gap-x-2">
									<Checkbox bind:checked={project.isPublic} />
									是
								</label>
							</TableBodyCell>
							<TableBodyCell>
								<Select
									bind:value={project.judgePriorityStrategy}
									required
									placeholder=""
									items={judgePriorityStrategies.map((strategy) => ({
										name: strategy,
										value: strategy
									}))}
								></Select>
							</TableBodyCell>
							<!-- 評測容器的三個資源上限;留空 = 沿用系統預設,0 = 不限制 -->
							<TableBodyCell class="space-y-2">
								<div>
									<span class="text-xs text-gray-500 dark:text-gray-400">時間(分)</span>
									<Input
										type="number"
										min="1"
										placeholder={judgeDurationPlaceholder}
										name="max-judge-duration-{index}"
										bind:value={project.maxJudgeDurationMinutes}
									/>
								</div>
								<div>
									<span class="text-xs text-gray-500 dark:text-gray-400">行程數</span>
									<Input
										type="number"
										min="0"
										placeholder={pidsPlaceholder}
										name="pids-limit-{index}"
										bind:value={project.pidsLimit}
									/>
								</div>
								<div>
									<span class="text-xs text-gray-500 dark:text-gray-400">記憶體(MB)</span>
									<Input
										type="number"
										min="0"
										placeholder={memoryPlaceholder}
										name="memory-limit-{index}"
										bind:value={project.memoryLimitMB}
									/>
								</div>
							</TableBodyCell>
							<TableBodyCell>
								<!-- 留空 = 沿用系統預設;0 = 明確不限制 -->
								<Input
									type="number"
									min="0"
									placeholder={cooldownPlaceholder}
									name="submission-cooldown-{index}"
									bind:value={project.submissionCooldownSeconds}
								/>
							</TableBodyCell>
							<TableBodyCell class="space-y-2">
								<label class="flex items-center gap-x-2 whitespace-nowrap">
									<Checkbox bind:checked={project.scoreboardCompact} />
									精簡樣式
								</label>
								<label class="flex items-center gap-x-2 whitespace-nowrap">
									<Checkbox bind:checked={project.scoreboardMaskNames} />
									姓名隱碼
								</label>
							</TableBodyCell>
							<TableBodyCell>
								<Button
									color="red"
									class="cursor-pointer"
									on:click={() => {
										payload.projects.splice(index, 1);
									}}
								>
									刪除
								</Button>
							</TableBodyCell>
						</TableBodyRow>
					{/each}
					<TableBodyRow>
						<TableBodyCell colspan={10}>
							<div class="flex justify-center">
								<Button
									color="green"
									class="cursor-pointer"
									on:click={() => {
										payload.projects.push({
											projectId: projects[0].id,
											name: projects[0].name,
											activateTime: toDatetimeLocal(),
											deactivateTime: toDatetimeLocal(),
											startTime: toDatetimeLocal(),
											endTime: toDatetimeLocal(),
											isPublic: false,
											judgePriorityStrategy: 'default',
											scoreboardCompact: false,
											scoreboardMaskNames: false
										});
									}}
								>
									新增專案
								</Button>
							</div>
						</TableBodyCell>
					</TableBodyRow>
				</TableBody>
			</Table>
			<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
				<strong>提交冷卻(秒)</strong>:同一位選手對同一題兩次提交之間的最短間隔。留空沿用系統預設{cooldownDefaultText
					? `(目前為${cooldownDefaultText})`
					: ''},填 <strong>0</strong> 代表明確不限制。
			</p>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
				<strong>評測資源上限</strong>:評測容器的時間、行程數與記憶體上限,同樣留空沿用系統預設、填
				<strong>0</strong> 代表不限制。<strong>行程數算的是 task(含執行緒)</strong>,不只是行程——chromium
				是少數幾個行程但數百條執行緒,估算時別照行程數抓。
			</p>
		</dir>

		<Button color="green" class="cursor-pointer" type="submit">建立競賽</Button>
	</form>
</div>
