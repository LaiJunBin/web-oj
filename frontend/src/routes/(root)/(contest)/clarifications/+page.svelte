<script lang="ts">
	import Swal from 'sweetalert2';
	import {
		Accordion,
		AccordionItem,
		Badge,
		Button,
		Card,
		Input,
		Label,
		Modal,
		Textarea,
		type TransitionTypes
	} from 'flowbite-svelte';
	import { useClarificationsStorage } from '@/apps/contest/domain/clarification';
	import { datetimeString } from '$lib';
	import { onMount, tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { useUserProfile } from '@/apps/contest/application/user';
	import { useProjectStorage } from '@/apps/contest/domain/project';
	import { useClarificationApp } from '@/apps/contest/application/clarification';

	const clarificationApp = useClarificationApp();
	const userProfile = useUserProfile();

	const clarificationsStorage = useClarificationsStorage();
	const projectStorage = useProjectStorage();

	const project = $derived(projectStorage.get());

	const clarifications = $derived(
		userProfile.isLoggedIn ? (clarificationsStorage.get() ?? []) : []
	);

	const announcements = $derived(
		clarifications.filter((clarification) => clarification.isAdminBroadcast)
	);
	const questions = $derived(
		clarifications.filter((clarification) => !clarification.isAdminBroadcast)
	);

	let lastFetchedProjectId: number | null = null;
	$effect(() => {
		const project = projectStorage.get();
		if (!project || !userProfile.isLoggedIn) return;
		if (lastFetchedProjectId === project.id) return;
		lastFetchedProjectId = project.id;
		clarificationApp.withLoading().fetchClarifications(project.id);
	});

	let openCreateModal = $state(false);
	let subject = $state('');
	let content = $state('');
	let transitionType: TransitionTypes = $state('fly');
	onMount(() => {
		tick().then(() => {
			setTimeout(() => {
				transitionType = 'slide';
			}, 10);
		});
	});
</script>

{#if !project}
	<p>Loading...</p>
{:else}
	<div class="p-4">
		<h1
			class="mb-4 flex items-center space-x-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
		>
			<span>{project.name} 提問與回應</span>
			{#if userProfile.isLoggedIn}
				<Button
					class="cursor-pointer px-4 py-2"
					color="light"
					onclick={() => (openCreateModal = true)}>+</Button
				>
			{/if}
		</h1>
		<Accordion multiple>
			<AccordionItem open class="cursor-pointer" bind:transitionType>
				<span slot="header"> 公告列表({announcements.length}) </span>
				<section>
					{#if announcements.length > 0}
						<div class="flex flex-col gap-y-4">
							{#each announcements as clarification}
								<Card class="block max-w-none" href="/clarifications/{clarification.id}">
									<div class="flex items-center justify-between gap-x-4">
										<p class="flex-1 overflow-hidden text-ellipsis">{clarification.subject}</p>
										<p class="w-max">{datetimeString(clarification.createdAt)}</p>
									</div>
								</Card>
							{/each}
						</div>
					{:else}
						<div class="text-gray-500">暫無公告</div>
					{/if}
				</section>
			</AccordionItem>

			<AccordionItem open class="cursor-pointer" bind:transitionType>
				<span slot="header">
					提問列表({questions.length})
				</span>
				<section>
					{#if questions.length > 0}
						<div class="flex flex-col gap-y-4">
							{#each questions as clarification}
								<Card class="block max-w-none" href="/clarifications/{clarification.id}">
									<div class="flex items-center justify-between gap-x-4">
										<p class="flex flex-1 items-center gap-x-2 overflow-hidden">
											<!-- 定向的工作人員通知(isAdminResponse、非公告、無父問題)會落在提問桶,
											     和自己的提問長得一樣;補一個識別標讓選手一眼分得出。 -->
											{#if clarification.isAdminResponse}
												<Badge color="purple" class="shrink-0">工作人員通知</Badge>
											{/if}
											<span class="overflow-hidden text-ellipsis">{clarification.subject}</span>
										</p>
										<p class="w-max">{datetimeString(clarification.createdAt)}</p>
									</div>
								</Card>
							{/each}
						</div>
					{:else}
						<div class="text-gray-500">暫無提問，點擊 + 新增提問</div>
					{/if}
				</section>
			</AccordionItem>
		</Accordion>
	</div>

	{#if userProfile.isLoggedIn}
		<Modal
			bind:open={openCreateModal}
			autoclose
			outsideclose
			title="新增提問"
			footerClass="flex justify-end items-center p-4 md:p-5 space-x-3 rtl:space-x-reverse rounded-b-lg"
		>
			<form
				id="submit-form"
				class="space-y-6"
				onsubmit={(e) => {
					e.preventDefault();
					clarificationApp
						.withLoading()
						.sendClarification(project.id, subject, content)
						.then((id) => {
							openCreateModal = false;
							Swal.fire({
								title: '提問成功'
							});
							goto(`/clarifications/${id}`);
						})
						.catch((error) => {
							console.error(error);
							Swal.fire({
								title: '送出失敗'
							});
						});
				}}
			>
				<div>
					<Label for="subject-input">
						<span class="align-middle text-red-500">*</span>
						主題
					</Label>

					<Input id="subject-input" name="subject" required bind:value={subject} />
				</div>

				<div>
					<Label for="content-input">
						<span class="align-middle text-red-500">*</span>
						內容
					</Label>

					<Textarea id="content-input" name="content" class="h-32" required bind:value={content} />
				</div>
			</form>
			<svelte:fragment slot="footer">
				<Button
					color="green"
					class="cursor-pointer"
					type="submit"
					form="submit-form"
					onclick={(e: any) => {
						e.stopPropagation();
					}}>送出</Button
				>
			</svelte:fragment>
		</Modal>
	{/if}
{/if}
