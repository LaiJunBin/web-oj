<script lang="ts">
	import { goto } from '$app/navigation';
	import { refreshContestState } from '@/lib/refreshContestState';
	import { useAssetApp } from '@/apps/admin/application/asset';
	import { useLifecycleApp } from '@/apps/admin/application/lifecycle';
	import { useProjectApp } from '@/apps/admin/application/project';
	import {
		LifecycleRunsOnEnum,
		LifecycleTypeEnum,
		type CreateLifecyclePayload,
		type LifecycleRunsOn,
		type LifecycleType
	} from '@/apps/admin/domain/lifecycle';
	import { useProjectStorage } from '@/apps/admin/domain/project';
	import { datetimeString, fileSize } from '@/lib';
	import {
		Accordion,
		AccordionItem,
		Badge,
		Button,
		ButtonGroup,
		Checkbox,
		Input,
		Label,
		Modal,
		Select,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';
	import TreeView from './TreeView.svelte';

	const projectApp = useProjectApp();
	const projectStorage = useProjectStorage();
	const project = $derived(projectStorage.get()!);

	const assetApp = useAssetApp();

	const lifecycleApp = useLifecycleApp();
	let openCreateLifecycleModal = $state(false);
	let lifecyclePayload: CreateLifecyclePayload = $state({
		runsOn: 'project',
		type: 'before_all',
		name: '',
		description: '',
		command: ''
	});

	let selectedFileContent = $state('');
	let selectedFilePath = $state('');

	async function fetchFileContent(filePath: string) {
		return projectApp.withLoading().readGradingProjectFile(project.id, btoa(filePath));
	}

	function base64DecodeUnicode(base64String: string): string {
		return decodeURIComponent(
			atob(base64String)
				.split('')
				.map(function (c) {
					return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
				})
				.join('')
		);
	}
	async function handleFileClick(filePath: string) {
		selectedFilePath = filePath;
		selectedFileContent = base64DecodeUnicode(await fetchFileContent(filePath));
	}

	$effect(() => {
		if (openCreateLifecycleModal) {
			lifecyclePayload = {
				runsOn: 'project',
				type: 'before_all',
				name: '',
				description: '',
				command: ''
			};
		}
	});

</script>

{#if !project}
	<div>Loading...</div>
{:else}
	<div class="flex flex-col p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/dashboard/projects"
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
				<p>{project.name}</p>
				<Button color="blue" href={`/dashboard/projects/${project.id}/grading-items`}>
					管理評分項目
				</Button>
				<ButtonGroup class="ml-auto">
					<Button color="primary" href={`/dashboard/projects/${project.id}/edit`}>編輯</Button>
					<Button
						color="alternative"
						class="cursor-pointer"
						onclick={() => {
							projectApp
								.exportProject(project.id)
								.catch((error) =>
									Swal.fire({ icon: 'error', title: '匯出失敗', text: error.message })
								);
						}}>匯出</Button
					>
					<Button
						color="red"
						class="cursor-pointer"
						onclick={async () => {
							Swal.fire({
								title: '確認刪除',
								text: '確定要刪除嗎？',
								icon: 'warning',
								showCancelButton: true,
								confirmButtonText: '確定',
								cancelButtonText: '取消'
							})
								.then(async (result) => {
									if (result.isConfirmed) {
										await projectApp.deleteProject(project.id).then(() => {
											refreshContestState();
											Swal.fire({
												title: '刪除成功',
												icon: 'success',
												showConfirmButton: false,
												timer: 1500
											});
										});
										goto('/dashboard/projects');
									}
								})
								.catch((error) => {
									Swal.fire({
										icon: 'error',
										title: '刪除失敗',
										text: error.message
									});
								});
						}}
					>
						刪除
					</Button>
				</ButtonGroup>
			</div>
		</h1>

		<div class="my-2">
			<h2 class="text-lg font-bold">描述訊息</h2>
			<p>{project.description || '無'}</p>
		</div>

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<div class="md:col-span-2">
				<h2 class="text-lg font-bold">測試結果解析方式</h2>
				<p>{project.resultParser || 'auto'}</p>
			</div>
			<div class="flex flex-col gap-y-2">
				<h2 class="text-lg font-bold">專案執行環境</h2>
				{#if project.image}
					<dir>
						<a href={`/dashboard/images/${project.image.tag}`} class="w-fit border-b"
							>{project.image.tag}
						</a>
						<Badge color={project.image.isReady ? 'green' : 'red'}>
							{project.image.isReady ? '可用' : '不可用'}
						</Badge>
					</dir>
				{:else}
					<p>無</p>
				{/if}

				<h3 class="flex items-center gap-x-2 text-lg font-bold">
					環境變數
					<Button
						color="primary"
						class="h-8 cursor-pointer"
						onclick={() => {
							Swal.fire({
								title: '環境變數',
								input: 'textarea',
								inputValue: project.env,
								showCancelButton: true,
								confirmButtonText: '確定',
								cancelButtonText: '取消'
							}).then((result) => {
								if (!result.isConfirmed) {
									return;
								}
								const projectEnv = result.value;
								projectApp
									.updateProject(project.id, { projectEnv })
									.then(() => {
										Swal.fire({
											title: '更新成功',
											icon: 'success',
											confirmButtonText: '確定'
										});
										projectApp.fetchProject(project.id);
									})
									.catch((error) => {
										console.error(error);
										Swal.fire({
											title: '更新失敗',
											text: error.message,
											icon: 'error',
											confirmButtonText: '確定'
										});
									});
							});
						}}>編輯</Button
					>
				</h3>
				<code
					class="block max-h-60 min-h-20 overflow-auto rounded-md bg-gray-100 p-2 whitespace-pre text-gray-600"
				>
					{project.env}
				</code>
			</div>

			<div class="flex flex-col gap-y-2">
				<h2 class="text-lg font-bold">測試執行環境</h2>
				{#if project.testImage}
					<div>
						<a href={`/dashboard/images/${project.testImage.tag}`} class="w-fit border-b"
							>{project.testImage.tag}</a
						>
						<Badge color={project.testImage.isReady ? 'green' : 'red'}>
							{project.testImage.isReady ? '可用' : '不可用'}
						</Badge>
					</div>
				{:else}
					<p>無</p>
				{/if}

				<h3 class="flex items-center gap-x-2 text-lg font-bold">
					環境變數
					<Button
						color="primary"
						class="h-8 cursor-pointer"
						onclick={() => {
							Swal.fire({
								title: '環境變數',
								input: 'textarea',
								inputValue: project.testEnv,
								showCancelButton: true,
								confirmButtonText: '確定',
								cancelButtonText: '取消'
							}).then((result) => {
								if (!result.isConfirmed) {
									return;
								}
								const testEnv = result.value;
								projectApp
									.updateProject(project.id, { testEnv })
									.then(() => {
										Swal.fire({
											title: '更新成功',
											icon: 'success',
											confirmButtonText: '確定'
										});
										projectApp.fetchProject(project.id);
									})
									.catch((error) => {
										console.error(error);
										Swal.fire({
											title: '更新失敗',
											text: error.message,
											icon: 'error',
											confirmButtonText: '確定'
										});
									});
							});
						}}>編輯</Button
					>
				</h3>
				<code
					class="block max-h-60 min-h-20 overflow-auto rounded-md bg-gray-100 p-2 whitespace-pre text-gray-600"
				>
					{project.testEnv}
				</code>
			</div>
		</div>

		<hr class="my-4" />
		<h2 class="mb-2 flex items-center gap-x-2 text-lg font-bold">
			提供素材
			<Button
				color="green"
				class="h-8 w-8 cursor-pointer"
				onclick={() => {
					const fileInput = document.createElement('input');
					fileInput.type = 'file';
					fileInput.onchange = async (event) => {
						const input = event.target as HTMLInputElement;
						const file = input.files?.[0];
						if (!file) {
							return;
						}

						assetApp
							.withLoading()
							.uploadAsset(project.id, {
								name: file.name,
								file
							})
							.then(() => {
								Swal.fire({
									title: '上傳素材成功',
									icon: 'success',
									showConfirmButton: false,
									timer: 1500
								});
								projectApp.fetchProject(project.id);
							})
							.catch((error) => {
								console.error(error);
								Swal.fire({
									title: '上傳素材失敗',
									icon: 'error',
									text: error.message,
									confirmButtonText: '確定'
								});
							});
					};
					fileInput.click();
				}}>+</Button
			>
		</h2>

		<div>
			<Table>
				<TableHead>
					<TableHeadCell>啟用?</TableHeadCell>
					<TableHeadCell>素材名稱(點擊可修改)</TableHeadCell>
					<TableHeadCell>大小(Bytes)</TableHeadCell>
					<TableHeadCell>最後修改時間</TableHeadCell>
					<TableHeadCell>
						<span class="sr-only">Actions</span>
					</TableHeadCell>
				</TableHead>
				<TableBody tableBodyClass="divide-y">
					{#each project.assets as asset}
						<TableBodyRow>
							<TableBodyCell>
								<Checkbox
									checked={asset.isEnabled}
									onchange={async (event) => {
										const isEnabled = event.currentTarget.checked;
										await assetApp.updateAsset(asset.id, { isEnabled }).then(() => {
											projectApp.fetchProject(project.id);
										});
									}}
								/>
							</TableBodyCell>
							<TableBodyCell>
								<button
									class="cursor-pointer"
									onclick={(e) => {
										e.stopPropagation();
										Swal.fire({
											title: '素材名稱',
											input: 'text',
											inputValue: asset.name,
											showCancelButton: true,
											confirmButtonText: '確定',
											cancelButtonText: '取消'
										}).then((result) => {
											if (!result.isConfirmed) {
												return;
											}
											const name = result.value;
											assetApp
												.updateAsset(asset.id, { name })
												.then(() => {
													Swal.fire({
														title: '更新成功',
														icon: 'success',
														confirmButtonText: '確定'
													});
													projectApp.fetchProject(project.id);
												})
												.catch((error) => {
													console.error(error);
													Swal.fire({
														title: '更新失敗',
														icon: 'error',
														confirmButtonText: '確定'
													});
												});
										});
									}}
								>
									{asset.name}
								</button>
							</TableBodyCell>
							<TableBodyCell>
								{fileSize(asset.size)}
							</TableBodyCell>
							<TableBodyCell>
								{datetimeString(asset.lastModified)}
							</TableBodyCell>
							<TableBodyCell>
								<ButtonGroup>
									<Button
										color="blue"
										class="cursor-pointer"
										onclick={() => {
											assetApp.withLoading().downloadAsset(asset.id, asset.name);
										}}>下載</Button
									>
									<Button
										color="red"
										class="cursor-pointer"
										onclick={() => {
											Swal.fire({
												title: '確認刪除',
												text: '確定要刪除嗎？',
												icon: 'warning',
												showCancelButton: true,
												confirmButtonText: '確定',
												cancelButtonText: '取消'
											})
												.then(async (result) => {
													if (result.isConfirmed) {
														await assetApp.deleteAsset(asset.id).then(() => {
															Swal.fire({
																title: '刪除成功',
																icon: 'success',
																showConfirmButton: false,
																timer: 1500
															});
														});
														projectApp.fetchProject(project.id);
													}
												})
												.catch((error) => {
													Swal.fire({
														icon: 'error',
														title: '刪除失敗',
														text: error.message
													});
												});
										}}>刪除</Button
									>
								</ButtonGroup>
							</TableBodyCell>
						</TableBodyRow>
					{:else}
						<TableBodyRow>
							<TableBodyCell colspan={4} class="text-center">無</TableBodyCell>
						</TableBodyRow>
					{/each}
				</TableBody>
			</Table>
		</div>

		<Modal bind:open={openCreateLifecycleModal} outsideclose title="新增生命週期">
			<form
				id="create-lifecycle-form"
				class="space-y-6"
				onsubmit={(e) => {
					e.preventDefault();
					lifecycleApp
						.createLifecycle(project.id, lifecyclePayload)
						.then(() => {
							openCreateLifecycleModal = false;
							projectApp.fetchProject(project.id);
							Swal.fire({
								icon: 'success',
								title: '新增成功',
								showConfirmButton: false,
								timer: 1500
							});
						})
						.catch((error) => {
							Swal.fire({
								icon: 'error',
								title: '新增失敗',
								text: error.message,
								confirmButtonText: '確定'
							});
						});
				}}
			>
				<div>
					<Label for="runsOn">
						<span class="align-middle text-red-500">*</span>
						目標執行環境
					</Label>
					<Select
						required
						id="runsOn"
						bind:value={lifecyclePayload.runsOn}
						items={Object.values(LifecycleRunsOnEnum).map((value) => ({
							value,
							name: value
						}))}
					/>
				</div>

				<div>
					<Label for="type">
						<span class="align-middle text-red-500">*</span>
						生命週期類型
					</Label>
					<Select
						id="type"
						required
						bind:value={lifecyclePayload.type}
						items={Object.values(LifecycleTypeEnum).map((value) => ({
							value,
							name: value
						}))}
					/>
				</div>

				<div>
					<Label for="name">
						<span class="align-middle text-red-500">*</span>
						名稱
					</Label>
					<Input id="name" bind:value={lifecyclePayload.name} required />
				</div>

				<div>
					<Label for="command">
						<span class="align-middle text-red-500">*</span>
						指令
					</Label>
					<Input id="command" bind:value={lifecyclePayload.command} required />
				</div>

				<div>
					<Label for="description">描述</Label>
					<Input id="description" bind:value={lifecyclePayload.description} />
				</div>
			</form>

			<svelte:fragment slot="footer">
				<Button
					color="green"
					class="w-full cursor-pointer"
					type="submit"
					form="create-lifecycle-form"
					onclick={(e: any) => e.stopPropagation()}>建立生命週期</Button
				>
			</svelte:fragment>
		</Modal>

		<hr class="my-4" />
		<h2 class="mb-2 flex items-center gap-x-2 text-lg font-bold">
			生命週期
			<Button
				color="green"
				class="h-8 w-8 cursor-pointer"
				onclick={() => (openCreateLifecycleModal = true)}>+</Button
			>
		</h2>
		<Accordion multiple>
			<AccordionItem class="cursor-pointer">
				<span slot="header" class="text-lg font-bold"
					>專案執行環境 ({Object.values(project.lifecycle).flat().length})</span
				>
				<section>
					{#each Object.keys(project.lifecycle) as key}
						<h3 class="text-lg font-bold">
							{key.charAt(0).toUpperCase() + key.slice(1)} ({project.lifecycle[
								key as keyof typeof project.lifecycle
							].length})
						</h3>
						<Table>
							<TableHead>
								<TableHeadCell>Name</TableHeadCell>
								<TableHeadCell>Description</TableHeadCell>
								<TableHeadCell>Command</TableHeadCell>
								<TableHeadCell>
									<span class="sr-only">Actions</span>
								</TableHeadCell>
							</TableHead>
							<TableBody tableBodyClass="divide-y">
								{#each project.lifecycle[key as keyof typeof project.lifecycle] as field}
									<TableBodyRow>
										<TableBodyCell>
											{field.name}
										</TableBodyCell>
										<TableBodyCell>
											{field.description}
										</TableBodyCell>
										<TableBodyCell>
											{field.command}
										</TableBodyCell>
										<TableBodyCell>
											<Button
												color="red"
												class="cursor-pointer"
												onclick={() => {
													Swal.fire({
														title: '確認刪除',
														text: '確定要刪除嗎？',
														icon: 'warning',
														showCancelButton: true,
														confirmButtonText: '確定',
														cancelButtonText: '取消'
													})
														.then(async (result) => {
															if (result.isConfirmed) {
																await lifecycleApp.deleteLifecycle(field.id).then(() => {
																	Swal.fire({
																		title: '刪除成功',
																		icon: 'success',
																		showConfirmButton: false,
																		timer: 1500
																	});
																});
																projectApp.fetchProject(project.id);
															}
														})
														.catch((error) => {
															Swal.fire({
																icon: 'error',
																title: '刪除失敗',
																text: error.message
															});
														});
												}}>刪除</Button
											>
										</TableBodyCell>
									</TableBodyRow>
								{:else}
									<TableBodyRow>
										<TableBodyCell colspan={4} class="text-center">無</TableBodyCell>
									</TableBodyRow>
								{/each}
							</TableBody>
						</Table>
					{/each}
				</section>
			</AccordionItem>

			<AccordionItem class="cursor-pointer">
				<span slot="header" class="text-lg font-bold"
					>測試執行環境 ({Object.values(project.testLifecycle).flat().length})</span
				>
				<section>
					{#each Object.keys(project.testLifecycle) as key}
						<h3 class="text-lg font-bold">
							{key.charAt(0).toUpperCase() + key.slice(1)} ({project.testLifecycle[
								key as keyof typeof project.testLifecycle
							].length})
						</h3>
						<Table>
							<TableHead>
								<TableHeadCell>Name</TableHeadCell>
								<TableHeadCell>Description</TableHeadCell>
								<TableHeadCell>Command</TableHeadCell>
								<TableHeadCell>
									<span class="sr-only">Actions</span>
								</TableHeadCell>
							</TableHead>
							<TableBody tableBodyClass="divide-y">
								{#each project.testLifecycle[key as keyof typeof project.testLifecycle] as field}
									<TableBodyRow>
										<TableBodyCell>
											{field.name}
										</TableBodyCell>
										<TableBodyCell>
											{field.description}
										</TableBodyCell>
										<TableBodyCell>
											{field.command}
										</TableBodyCell>
										<TableBodyCell>
											<Button
												color="red"
												class="cursor-pointer"
												onclick={() => {
													Swal.fire({
														title: '確認刪除',
														text: '確定要刪除嗎？',
														icon: 'warning',
														showCancelButton: true,
														confirmButtonText: '確定',
														cancelButtonText: '取消'
													})
														.then(async (result) => {
															if (result.isConfirmed) {
																await lifecycleApp.deleteLifecycle(field.id).then(() => {
																	Swal.fire({
																		title: '刪除成功',
																		icon: 'success',
																		showConfirmButton: false,
																		timer: 1500
																	});
																});
																projectApp.fetchProject(project.id);
															}
														})
														.catch((error) => {
															Swal.fire({
																icon: 'error',
																title: '刪除失敗',
																text: error.message
															});
														});
												}}>刪除</Button
											>
										</TableBodyCell>
									</TableBodyRow>
								{:else}
									<TableBodyRow>
										<TableBodyCell colspan={4} class="text-center">無</TableBodyCell>
									</TableBodyRow>
								{/each}
							</TableBody>
						</Table>
					{/each}
				</section>
			</AccordionItem>
		</Accordion>

		<hr class="my-4" />

		<h2 class="text-lg font-bold">動態欄位</h2>
		<div>
			<Table>
				<TableHead>
					<TableHeadCell>Label</TableHeadCell>
					<TableHeadCell>Name</TableHeadCell>
					<TableHeadCell>Type</TableHeadCell>
					<TableHeadCell>Default Value</TableHeadCell>
					<TableHeadCell>Is Required</TableHeadCell>
				</TableHead>
				<TableBody tableBodyClass="divide-y">
					{#each project.dynamicFields as field, index}
						<TableBodyRow>
							<TableBodyCell>
								{field.label}
							</TableBodyCell>
							<TableBodyCell>
								{field.name}
							</TableBodyCell>
							<TableBodyCell>
								{field.type}
							</TableBodyCell>
							<TableBodyCell>
								{field.defaultValue}
							</TableBodyCell>
							<TableBodyCell>
								{field.isRequired ? '是' : '否'}
							</TableBodyCell>
						</TableBodyRow>
					{:else}
						<TableBodyRow>
							<TableBodyCell colspan={5} class="text-center">無</TableBodyCell>
						</TableBodyRow>
					{/each}
				</TableBody>
			</Table>
		</div>

		<hr class="my-4" />

		<h2 class="flex items-center gap-x-2 text-lg font-bold">
			<span>評分專案</span>
			<Button
				color="green"
				class="cursor-pointer"
				onclick={() => {
					const fileInput = document.createElement('input');
					fileInput.type = 'file';
					fileInput.onchange = async (event) => {
						const input = event.target as HTMLInputElement;
						const file = input.files?.[0];
						if (!file) {
							return;
						}
						projectApp
							.withLoading()
							.putGradingProject(project.id, {
								project: file
							})
							.then(() => {
								Swal.fire({
									title: '上傳評分專案成功',
									icon: 'success',
									showConfirmButton: false,
									timer: 1500
								});
								projectApp.fetchProject(project.id);
							})
							.catch((error) => {
								console.error(error);
								Swal.fire({
									title: '上傳評分專案失敗',
									icon: 'error',
									text: error.message,
									confirmButtonText: '確定'
								});
							});
					};
					fileInput.click();
				}}>PUT 評分專案</Button
			>

			<Button
				color="blue"
				class="cursor-pointer"
				onclick={() => {
					projectApp.withLoading().downloadGradingProject(project.id);
				}}>下載評分專案</Button
			>
		</h2>

		<hr class="my-4" />
		<h2 class="text-lg font-bold">檔案結構</h2>
		<div class="flex h-[70vh] gap-4 text-black">
			<!-- Tree View -->
			<div class="w-72 overflow-auto border border-gray-300 bg-gray-100 px-4 py-2">
				{#if project.files.length === 0}
					<p class="text-center text-gray-500">無檔案</p>
				{:else}
					<TreeView files={project.files} {selectedFilePath} onFileClick={handleFileClick} />
				{/if}
			</div>

			<div class="flex-1 overflow-y-auto border border-gray-300 bg-white p-4">
				{#if project.files.length === 0}
					<p class="text-center text-gray-500">無檔案</p>
				{:else}
					<!-- File Content -->
					<pre class="break-words whitespace-pre-wrap">{selectedFileContent || '選擇一個檔案'}</pre>
				{/if}
			</div>
		</div>
	</div>
{/if}
