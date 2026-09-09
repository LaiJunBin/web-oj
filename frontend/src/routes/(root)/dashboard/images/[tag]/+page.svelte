<script lang="ts">
	import { goto } from '$app/navigation';
	import { useImageApp } from '@/apps/admin/application/image';
	import { askExportMode, showExporting, showExportError } from '@/lib/exportMode';
	import { BuildLogEventDataSchema, type BuildLog } from '@/apps/admin/domain/build';
	import { useImageStorage } from '@/apps/admin/domain/image';
	import { datetimeString } from '@/lib';
	import { streamBuildImageLog } from '@/lib/api/stream';
	import { Badge, Button, ButtonGroup, Select } from 'flowbite-svelte';
	import { tick } from 'svelte';
	import Swal from 'sweetalert2';

	const imageApp = useImageApp();
	const image = $derived(useImageStorage().get()!);
	let pre: HTMLElement | null = $state(null);
	let logs: BuildLog[] = $state([]);
	let selectedBuildIndex: number | '' = $state(
		image === null || image.builds.length === 0 ? '' : image.builds.length - 1
	);
	let streaming = $state(true);
	let follow = $state(true);

	$effect(() => {
		if (!image || selectedBuildIndex === '' || !streaming) return;
		const build = image.builds[selectedBuildIndex];
		const eventSource = streamBuildImageLog(image.tag, build.batch);

		eventSource.onopen = (event) => {
		};

		eventSource.addEventListener('log', (event) => {
			try {
				const { success, data, error } = BuildLogEventDataSchema.safeParse(JSON.parse(event.data));
				if (!success) {
					console.error(error);
					return;
				}

				if (build) {
					if (build.status === 'PENDING') {
						build.status = 'BUILDING';
					}
				}

				logs.push(data);

				if (follow) {
					tick().then(() => {
						pre?.scrollTo(0, pre.scrollHeight);
					});
				}
			} catch (e) {
				console.error('Failed to parse build log event', e);
			}
		});

		eventSource.onerror = (event) => {
			streaming = false;
			console.error(event);
			imageApp.fetchImage(image.tag);
		};

		return () => {
			eventSource.close();
		};
	});
</script>

{#if !image}
	<div>Loading...</div>
{:else}
	<div class="flex h-full flex-col p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/dashboard/images"
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
				<p>{image.tag}</p>
				<p class="my-4 flex items-center gap-2 text-gray-500 dark:text-gray-400">
					<Badge color={image.isReady ? 'green' : 'red'}>
						{image.isReady ? '可用' : '不可用'}
					</Badge>
				</p>
				<ButtonGroup class="ml-auto">
					<Button color="primary" href={`/dashboard/images/${image.tag}/edit`}>編輯</Button>
					<Button
						color="alternative"
						class="cursor-pointer"
						onclick={async () => {
							const mode = await askExportMode(
								'匯出映像',
								`要把「${image.tag}」帶到另一台機器。選映像以哪種形態隨行:`
							);
							if (!mode) return;
							const done = showExporting();
							imageApp
								.exportImage(image.tag, mode)
								.then(done)
								.catch((error) => showExportError(error));
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
										await imageApp.deleteImage(image.tag).then(() => {
											Swal.fire({
												title: '刪除成功',
												icon: 'success',
												showConfirmButton: false,
												timer: 1500
											});
										});
										goto('/dashboard/images');
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

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<div>
				<h2 class="text-lg font-bold">描述訊息</h2>
				<p>{image.description || '無'}</p>
			</div>

			<div>
				<h2 class="text-lg font-bold">掛載路徑</h2>
				<p>{image.mountPath}</p>
			</div>
		</div>

		<div class="mt-8 flex h-0 flex-grow flex-col">
			<h2 class="mb-2 flex items-center space-x-2 text-lg font-bold">
				<p>建置紀錄</p>
				{#if image.builds.length === 0 || ['SUCCESS', 'ERROR', 'CANCELLED'].includes(image.builds[image.builds.length - 1].status)}
					<Button
						color="primary"
						class="cursor-pointer"
						onclick={() => {
							imageApp
								.rebuildImage(image.tag)
								.then(() => {
									imageApp.fetchImage(image.tag).then(() => {
										logs = [];
										streaming = true;
										selectedBuildIndex = image.builds.length - 1;
									});
								})
								.catch((error) => {
									Swal.fire({
										icon: 'error',
										title: '重置失敗',
										text: error.message
									});
								});
						}}
					>
						重建
					</Button>
				{:else}
					<Button
						color="red"
						class="cursor-pointer"
						onclick={async () => {
							await imageApp.cancelBuildImage(image.tag);
							await imageApp.fetchImage(image.tag);
						}}
					>
						取消
					</Button>
				{/if}
			</h2>
			<Select
				items={image.builds.map((build, i) => ({
					name: `批次: ${build.batch} - 狀態: ${build.status}`,
					value: i
				}))}
				placeholder="選擇建置紀錄"
				class="w-full"
				bind:value={selectedBuildIndex}
				onchange={() => {
					logs = [];
					streaming = true;
				}}
			/>

			<div class="mt-4 flex h-0 flex-grow flex-col">
				<p class="text-lg font-bold">建置日誌</p>
				{#if selectedBuildIndex === ''}
					<p>無日誌</p>
				{:else}
					<pre
						class="mt-4 flex-1 overflow-auto rounded-lg bg-gray-100 p-4 whitespace-nowrap dark:bg-gray-700"
						bind:this={pre}
						onscroll={(e) => {
							if (!pre) return;
							follow = pre.scrollHeight - pre.scrollTop === pre.clientHeight;
						}}>
					{#each logs as log}
							{#if log?.message.trim()}
								<div class="[content-visibility:auto] [contain-intrinsic-size:auto_3rem]">
								<div class="text-sm text-gray-600 dark:text-gray-400">
									{datetimeString(log?.time)}
								</div>
								<div class="text-lg font-bold">{log?.message}</div>
							</div>
							{/if}
						{/each}

					{#if streaming}
							<div
								class="mt-2 h-4 w-4 animate-spin rounded-full border-t-2 border-b-2 border-gray-900 dark:border-gray-100"></div>
						{/if}
				</pre>
				{/if}
			</div>
		</div>
	</div>
{/if}
