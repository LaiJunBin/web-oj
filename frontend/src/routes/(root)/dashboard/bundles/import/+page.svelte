<script lang="ts">
	import { goto } from '$app/navigation';
	import { useBundleApp } from '@/apps/admin/application/bundle';
	import {
		BundleConflictError,
		type BundleImportReport,
		type BundlePreview
	} from '@/apps/admin/domain/bundle';
	import { fileSize } from '@/lib';
	import {
		Badge,
		Button,
		Fileupload,
		Label,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';
	import { sanitizeHtml } from '$lib/sanitize';
	import { refreshContestState } from '@/lib/refreshContestState';

	const bundleApp = useBundleApp();

	let file = $state<File | null>(null);
	let preview = $state<BundlePreview | null>(null);
	let report = $state<BundleImportReport | null>(null);
	let busy = $state(false);

	// blockers 裡混了兩種性質不同的東西,後端匯入時也是這樣分的:
	//
	// - 缺少映像:force 也救不回來(替換不出一個誰都沒有的映像),是真正的死路。
	// - 映像內容衝突:可以覆蓋,只是要先讓操作者知道會影響誰——這正是 force 的用途。
	//
	// 所以不能一律用 blockers 非空來停用按鈕,那會讓衝突永遠無法匯入。判斷依據
	// 取結構化的 missingImageTags,而不是去解析 blockers 的字串。
	const missingImages = $derived(
		preview ? [...new Set(preview.projects.flatMap((p) => p.missingImageTags))] : []
	);
	const hardBlocked = $derived(missingImages.length > 0);
	const willOverwrite = $derived(
		!!preview && !hardBlocked && preview.images.some((image) => image.status === 'conflict')
	);
	const canImport = $derived(!!preview && !hardBlocked && !busy);

	const statusBadge = (status: string) =>
		status === 'new'
			? { color: 'green' as const, text: '新建' }
			: status === 'identical'
				? { color: 'dark' as const, text: '已存在(相同)' }
				: { color: 'yellow' as const, text: '會覆蓋' };

	const onFileChange = (event: Event) => {
		const input = event.target as HTMLInputElement;
		file = input.files?.[0] ?? null;
		preview = null;
		report = null;
	};

	const runPreview = () => {
		if (!file) return;
		busy = true;
		bundleApp
			.previewBundle(file)
			.then((result) => {
				preview = result;
			})
			.catch((error) => {
				preview = null;
				Swal.fire({
					icon: 'error',
					title: '無法讀取這個檔案',
					html: `<pre style="text-align:left;white-space:pre-wrap">${sanitizeHtml(error.message)}</pre>`
				});
			})
			.finally(() => {
				busy = false;
			});
	};

	const runImport = (force = false) => {
		if (!file) return;
		busy = true;
		bundleApp
			.importBundle(file, force)
			.then((result) => {
				report = result;
				// 包裹會建立一場比賽;導覽列的比賽下拉要跟上,不能等使用者重新整理。
				refreshContestState();
			})
			.catch((error) => {
				// 409 不是失敗,是「會覆蓋既有東西,請你確認」。後端已經把影響範圍寫成
				// 人話,逐條原文顯示比我們自己改寫更準確。
				if (error instanceof BundleConflictError) {
					Swal.fire({
						icon: 'warning',
						title: '這次匯入會覆蓋既有資料',
						html: `<div style="text-align:left">
							<p>${sanitizeHtml(error.message)}</p>
							<ul style="margin-top:.5rem;padding-left:1.2rem;list-style:disc">
								${error.details.map((d) => `<li>${sanitizeHtml(d)}</li>`).join('')}
							</ul>
							<p style="margin-top:.75rem">確定要覆蓋嗎?此動作無法復原。</p>
						</div>`,
						showCancelButton: true,
						confirmButtonText: '覆蓋並匯入',
						cancelButtonText: '取消',
						confirmButtonColor: '#d33'
					}).then((result) => {
						if (result.isConfirmed) runImport(true);
					});
					return;
				}
				Swal.fire({
					icon: 'error',
					title: '匯入失敗',
					html: `<pre style="text-align:left;white-space:pre-wrap">${sanitizeHtml(error.message)}</pre>`
				});
			})
			.finally(() => {
				busy = false;
			});
	};
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center gap-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>匯入完整競賽</span>
	</h1>

	{#if report}
		<!-- 匯入報告 -->
		<div
			class="rounded-lg border border-green-300 bg-green-50 p-4 dark:border-green-700 dark:bg-green-900/30"
		>
			<h2 class="text-lg font-bold text-green-800 dark:text-green-200">匯入完成</h2>
			<ul class="mt-3 list-disc space-y-1 ps-5 text-sm">
				{#if report.createdImages.length}
					<li>新建映像:{report.createdImages.join('、')}</li>
				{/if}
				{#if report.reusedImages.length}
					<li>沿用既有映像:{report.reusedImages.join('、')}</li>
				{/if}
				{#if report.createdProjectIds.length}
					<li>新建題目:{report.createdProjectIds.length} 個</li>
				{/if}
				{#if report.contestId}
					<li>
						新建競賽:<a
							class="underline"
							href={`/dashboard/contests/${report.contestId}`}
							data-sveltekit-reload>前往查看</a
						>
					</li>
				{/if}
			</ul>
		</div>

		{#if report.queuedBuilds.length}
			<!-- source 模式匯入的映像還要建置完才能評測,不導流過去的話,使用者會在
				 開賽前才發現題目不能跑 -->
			<div
				class="mt-3 rounded-lg border border-blue-300 bg-blue-50 p-4 text-sm dark:border-blue-700 dark:bg-blue-900/30"
			>
				<p class="font-bold">有 {report.queuedBuilds.length} 個映像正在建置</p>
				<p class="mt-1">
					{report.queuedBuilds.join('、')}
				</p>
				<p class="mt-2">建置完成前這些題目無法評測。請到評分環境管理確認建置結果。</p>
				<Button class="mt-3" color="blue" href="/dashboard/images">前往評分環境管理</Button>
			</div>
		{/if}

		{#if report.warnings.length}
			<div
				class="mt-3 rounded-lg border border-yellow-300 bg-yellow-50 p-4 text-sm dark:border-yellow-700 dark:bg-yellow-900/30"
			>
				<!-- 匯入中途失敗現在是非 2xx,會走上面的錯誤分支;留在這裡的是「匯入成功、
					 但有事情想讓你知道」的備註,所以標題不能再寫成失敗 -->
				<p class="font-bold">匯入時的提醒</p>
				<ul class="mt-2 list-disc space-y-1 ps-5">
					{#each report.warnings as warning}
						<li>{warning}</li>
					{/each}
				</ul>
			</div>
		{/if}

		{#if report.unresolvedProjects.length}
			<div
				class="mt-3 rounded-lg border border-yellow-300 bg-yellow-50 p-4 text-sm dark:border-yellow-700 dark:bg-yellow-900/30"
			>
				<p class="font-bold">競賽中有題目未能對應</p>
				<p class="mt-1">{report.unresolvedProjects.join('、')}</p>
				<p class="mt-2">這些欄位在競賽中留空,請自行掛上對應的題目。</p>
			</div>
		{/if}

		<div
			class="mt-3 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
		>
			<p><strong>參賽者不會一起匯入</strong>,匯入後請自行把選手或群組加入這場競賽。</p>
		</div>

		<div class="mt-4 flex gap-2">
			{#if report.contestId}
				<Button color="green" onclick={() => goto(`/dashboard/contests/${report?.contestId}`)}>
					前往競賽
				</Button>
			{/if}
			<Button
				color="alternative"
				class="cursor-pointer"
				onclick={() => {
					file = null;
					preview = null;
					report = null;
				}}>再匯入一個</Button
			>
		</div>
	{:else}
		<div class="max-w-xl">
			<Label for="bundle-file" class="mb-2">完整匯出檔(.zip)</Label>
			<Fileupload id="bundle-file" accept=".zip" onchange={onFileChange} />
			{#if file}
				<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{file.name}({fileSize(file.size)})
				</p>
			{/if}
			<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
				這一頁只接受競賽詳情頁「<strong>完整匯出</strong>」產生的檔案。單一題目、映像或競賽設定,請到各自的列表頁用「匯入」。
			</p>
			<Button class="mt-3" color="blue" disabled={!file || busy} onclick={runPreview}>
				{busy && !preview ? '讀取中…' : '檢視內容'}
			</Button>
		</div>

		{#if preview}
			<hr class="my-6" />
			<h2 class="mb-3 text-xl font-bold">匯入後會發生什麼</h2>

			{#if hardBlocked}
				<div
					class="mb-4 rounded-lg border border-red-300 bg-red-50 p-4 text-sm dark:border-red-700 dark:bg-red-900/30"
				>
					<p class="font-bold text-red-800 dark:text-red-200">無法匯入</p>
					<ul class="mt-2 list-disc space-y-1 ps-5">
						{#each preview.blockers as blocker}
							<li>{blocker}</li>
						{/each}
					</ul>
					<p class="mt-2">
						缺少的映像({missingImages.join('、')})必須先存在於這台機器,或在來源機器改用含映像的形態重新「完整匯出」。
					</p>
				</div>
			{:else if willOverwrite}
				<div
					class="mb-4 rounded-lg border border-yellow-300 bg-yellow-50 p-4 text-sm dark:border-yellow-700 dark:bg-yellow-900/30"
				>
					<p class="font-bold">會覆蓋既有資料</p>
					<ul class="mt-2 list-disc space-y-1 ps-5">
						{#each preview.blockers as blocker}
							<li>{blocker}</li>
						{/each}
					</ul>
					<p class="mt-2">可以繼續匯入,執行時會再請你確認一次。</p>
				</div>
			{/if}

			{#if preview.images.length}
				<h3 class="mt-4 mb-2 font-bold">映像</h3>
				<Table divClass="relative overflow-auto">
					<TableHead>
						<TableHeadCell>Tag</TableHeadCell>
						<TableHeadCell>形態</TableHeadCell>
						<TableHeadCell>結果</TableHeadCell>
						<TableHeadCell>影響的題目</TableHeadCell>
					</TableHead>
					<TableBody tableBodyClass="divide-y">
						{#each preview.images as image}
							{@const badge = statusBadge(image.status)}
							<TableBodyRow>
								<TableBodyCell>{image.tag}</TableBodyCell>
								<TableBodyCell>{image.mode === 'archive' ? '映像檔' : '建置原始檔'}</TableBodyCell>
								<TableBodyCell><Badge color={badge.color}>{badge.text}</Badge></TableBodyCell>
								<TableBodyCell>
									{#if image.usedByProjects.length}
										<span class="text-sm">{image.usedByProjects.join('、')}</span>
									{:else}
										<span class="text-sm text-gray-400">—</span>
									{/if}
								</TableBodyCell>
							</TableBodyRow>
						{/each}
					</TableBody>
				</Table>
			{/if}

			{#if preview.projects.length}
				<h3 class="mt-6 mb-2 font-bold">題目</h3>
				<Table divClass="relative overflow-auto">
					<TableHead>
						<TableHeadCell>名稱</TableHeadCell>
						<TableHeadCell>結果</TableHeadCell>
						<TableHeadCell>缺少的映像</TableHeadCell>
					</TableHead>
					<TableBody tableBodyClass="divide-y">
						{#each preview.projects as project}
							{@const badge = statusBadge(project.status)}
							<TableBodyRow>
								<TableBodyCell>{project.name}</TableBodyCell>
								<TableBodyCell><Badge color={badge.color}>{badge.text}</Badge></TableBodyCell>
								<TableBodyCell>
									{#if project.missingImageTags.length}
										<span class="text-sm text-red-600 dark:text-red-400"
											>{project.missingImageTags.join('、')}</span
										>
									{:else}
										<span class="text-sm text-gray-400">—</span>
									{/if}
								</TableBodyCell>
							</TableBodyRow>
						{/each}
					</TableBody>
				</Table>
			{/if}

			{#if preview.contest}
				<h3 class="mt-6 mb-2 font-bold">競賽</h3>
				<p class="text-sm">
					{preview.contest.name}(共 {preview.contest.projectSlots} 個題目欄位)
					{#if preview.contest.unresolvedProjects.length}
						<span class="text-yellow-700 dark:text-yellow-400">
							—— 其中 {preview.contest.unresolvedProjects.join('、')} 無法對應,會留空
						</span>
					{/if}
				</p>
			{/if}

			<div
				class="mt-6 rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm dark:border-gray-700 dark:bg-gray-800"
			>
				<p><strong>參賽者不會一起匯入</strong>,匯入後請自行把選手或群組加入這場競賽。</p>
			</div>

			<Button class="mt-4" color="green" disabled={!canImport} onclick={() => runImport(false)}>
				{busy ? '匯入中…' : '執行匯入'}
			</Button>
			{#if hardBlocked}
				<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">請先解決上方的問題再匯入。</p>
			{/if}
		{/if}
	{/if}
</div>
