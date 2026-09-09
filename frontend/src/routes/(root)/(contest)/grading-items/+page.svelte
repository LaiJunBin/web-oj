<script lang="ts">
	import { Accordion, AccordionItem } from 'flowbite-svelte';
	import { page } from '$app/state';
	import { tick } from 'svelte';
	import { useProjectStorage } from '@/apps/contest/domain/project';

	const projectStorage = useProjectStorage();

	let gradingItemElements: HTMLElement[] = $state([]);
	const project = $derived(projectStorage.get());

	const formatGradingItemName = (item: { label?: string; name: string }) =>
		item.label ? item.label + ": " + item.name : item.name;
	$effect(() => {
		if (!page.state?.id || !project) return;
		tick().then(() => {
			const index = project.gradingItems.findIndex(
				(gradingItem) => gradingItem.id === page.state.id
			);
			if (index === -1) return;
			gradingItemElements[index].scrollIntoView({ behavior: 'smooth' });
		});
	});
</script>

{#if !project}
	<p>Loading...</p>
{:else}
	<div class="p-4">
		<h1
			class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
		>
			{project.name} 評分項目
		</h1>
		<Accordion multiple>
			<AccordionItem open class="pointer-events-none">
				<span slot="header">
					評分項目 ({project.timeLeft < 0 ? '?' : project.gradingItems.length})
				</span>
				<section>
					{#if project.timeLeft < 0}
						<p>評分項目將在比賽開始後公開</p>
					{:else}
						<Accordion multiple>
							{#if project.gradingItems.length}
								{#each project.gradingItems as gradingItem, i}
									<AccordionItem class="cursor-pointer" open={page.state?.id === gradingItem.id}>
										<span slot="header">
											<span bind:this={gradingItemElements[i]}
												>{formatGradingItemName(gradingItem)} ({gradingItem.score}分)</span
											>
										</span>
										<section>
											{gradingItem.description}
										</section>
										<hr class="my-4" />
										<section>
											公開的測試案例:
											<section class="pl-4">
												{#each gradingItem.publicTestcases as testcase, j}
													<p>{j + 1}. {testcase.name}</p>
												{:else}
													<p>無公開的測試案例</p>
												{/each}
											</section>
										</section>
									</AccordionItem>
								{/each}
							{:else}
								<AccordionItem open class="cursor-pointer">
									<span slot="header"> 無評分項目 </span>
								</AccordionItem>
							{/if}
						</Accordion>
					{/if}
				</section>
			</AccordionItem>
		</Accordion>
	</div>
{/if}
