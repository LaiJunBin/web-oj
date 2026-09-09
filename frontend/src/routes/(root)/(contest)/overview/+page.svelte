<script lang="ts">
	import { Accordion, AccordionItem, Button, Tooltip } from 'flowbite-svelte';
	import Swal from 'sweetalert2';
	import { useProjectStorage } from '@/apps/contest/domain/project';
	import { useProjectApp } from '@/apps/contest/application/project';
	import { DownloadSolid } from 'flowbite-svelte-icons';
	import { sanitizeHtml } from '$lib/sanitize';

	const projectStorage = useProjectStorage();
	const projectApp = useProjectApp();

	const project = $derived(projectStorage.get());
</script>

{#if !project}
	<p>Loading...</p>
{:else}
	<div class="p-4">
		<h1
			class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
		>
			{project.name}
		</h1>
		<Accordion multiple>
			<AccordionItem open class="cursor-pointer">
				<span slot="header">專案描述</span>
				<section>
					{#if project.timeLeft < 0}
						<p>待比賽開始後公開</p>
					{:else}
						<!-- Descriptions are admin- or bundle-authored and rendered as HTML;
						     sanitize so an imported third-party bundle cannot inject script. -->
						{@html sanitizeHtml(project.description || '無')}
					{/if}
				</section>
			</AccordionItem>
			<AccordionItem open class="cursor-pointer">
				<span slot="header">提供素材 ({project.timeLeft < 0 ? '?' : project.assets.length})</span>
				<section>
					<section
						class="grid gap-2"
						style="grid-template-columns: repeat(auto-fill, minmax(10em, 1fr))"
					>
						{#each project.assets as asset}
							<Button
								outline
								class="h-24 w-full cursor-pointer"
								onclick={() => {
									projectApp
										.withLoading()
										.downloadAsset(project.id, asset.id, asset.name)
										.catch((err) => {
											Swal.fire({
												title: '下載失敗',
												text: err.message,
												icon: 'error',
												confirmButtonText: '確定'
											});
										});
								}}
							>
								<DownloadSolid />
								<span class="max-w-full overflow-hidden overflow-ellipsis whitespace-nowrap">
									{asset.name}
								</span>
							</Button>
							<Tooltip placement="bottom">
								{asset.name}
							</Tooltip>
						{:else}
							{#if project.timeLeft < 0}
								<p>待比賽開始後公開</p>
							{:else}
								<p>無提供素材</p>
							{/if}
						{/each}
					</section>
				</section>
			</AccordionItem>
		</Accordion>
	</div>
{/if}
