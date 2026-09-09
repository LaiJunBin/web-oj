<script lang="ts">
	import { goto } from '$app/navigation';
	import { useClarificationApp } from '@/apps/admin/application/clarification';
	import {
		useClarificationsStorage,
		useClarificationStorage
	} from '@/apps/admin/domain/clarification';
	import { datetimeString } from '@/lib';
	import {
		Badge,
		Button,
		Input,
		Label,
		Modal,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell,
		Textarea
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	// Admin view of a contest project's announcements / questions / replies.
	// Shared between the 競賽過程 page (driven by the sidebar picker) and the
	// 競賽管理 project page (any project, including closed ones), so both stay in
	// sync feature-wise.
	const { projectId, projectName }: { projectId: number; projectName: string } = $props();

	const clarificationApp = useClarificationApp();
	const clarifications = $derived(useClarificationsStorage().get());
	const clarification = $derived(useClarificationStorage().get());
	const announcements = $derived(
		clarifications?.filter((clarification) => clarification.sendToAll) || []
	);
	const questions = $derived(
		clarifications?.filter(
			(clarification) => !clarification.sendToAll && !clarification.receiverUser
		) || []
	);
	const responses = $derived(
		clarifications?.filter(
			(clarification) => !clarification.sendToAll && clarification.receiverUser
		) || []
	);

	$effect(() => {
		clarificationApp.withLoading().fetchClarifications(projectId);
	});

	let announcementPayload = $state({ subject: '', content: '' });
	let openCreateAnnouncementModal = $state(false);
	let openShowAnnouncementModal = $state(false);

	const selectedAnnouncement = $derived(clarification);
</script>

{#if !clarifications}
	<p>Loading...</p>
{:else}
	<h2
		class="mb-4 flex items-center gap-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<div>{projectName} 的公告</div>
		<Button
			color="green"
			onclick={() => {
				announcementPayload = { subject: '', content: '' };
				openCreateAnnouncementModal = true;
			}}
		>
			新增公告
		</Button>
	</h2>

	<Table striped hoverable shadow>
		<TableHead>
			<TableHeadCell>#</TableHeadCell>
			<TableHeadCell>標題</TableHeadCell>
			<TableHeadCell>時間</TableHeadCell>
		</TableHead>
		<TableBody tableBodyClass="divide-y">
			{#each announcements as announcement, index}
				<TableBodyRow
					class="cursor-pointer"
					onclick={(e) => {
						const target = e.target as HTMLElement;
						if (target.tagName === 'TD') {
							clarificationApp
								.withLoading()
								.fetchClarification(announcement.id)
								.then(() => {
									openShowAnnouncementModal = true;
								})
								.catch((error) => {
									console.error(error);
									Swal.fire({
										icon: 'error',
										title: '讀取失敗',
										text: error.message
									});
								});
						}
					}}
				>
					<TableBodyCell>{index + 1}</TableBodyCell>
					<TableBodyCell>{announcement.subject}</TableBodyCell>
					<TableBodyCell>
						{datetimeString(announcement.createdAt)}
					</TableBodyCell>
				</TableBodyRow>
			{:else}
				<TableBodyRow>
					<TableBodyCell class="text-center" colspan={3}>無公告</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>

	<hr class="my-4" />

	<h2
		class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		{projectName} 的選手提問
	</h2>

	<Table striped hoverable shadow>
		<TableHead>
			<TableHeadCell>#</TableHeadCell>
			<TableHeadCell>狀態(點擊切換)</TableHeadCell>
			<TableHeadCell>發送者</TableHeadCell>
			<TableHeadCell>標題</TableHeadCell>
			<TableHeadCell>時間</TableHeadCell>
		</TableHead>
		<TableBody tableBodyClass="divide-y">
			{#each questions as question, index}
				<TableBodyRow
					class="cursor-pointer"
					onclick={(e) => {
						const target = e.target as HTMLElement;
						if (target.tagName === 'TD') {
							goto(`/dashboard/clarifications/${question.id}`);
						}
					}}
				>
					<TableBodyCell>{index + 1}</TableBodyCell>
					<TableBodyCell>
						<Badge
							color={question.isAnswered ? 'green' : 'red'}
							onclick={async () => {
								try {
									if (question.isAnswered) {
										await clarificationApp.withLoading().markClarificationAsUnanswered(question.id);
									} else {
										await clarificationApp.withLoading().markClarificationAsAnswered(question.id);
									}
									clarificationApp.withLoading().fetchClarifications(projectId);
								} catch (error) {
									console.error(error);
									Swal.fire({
										icon: 'error',
										title: '操作失敗'
									});
								}
							}}
						>
							{question.isAnswered ? '已解決' : '未解決'}
						</Badge>
					</TableBodyCell>
					<TableBodyCell>
						<a href="/dashboard/users/{question.senderUser.id}" class="border-b">
							{question.senderUser.name}({question.senderUser.username})
						</a>
					</TableBodyCell>
					<TableBodyCell>{question.subject}</TableBodyCell>
					<TableBodyCell>
						{datetimeString(question.createdAt)}
					</TableBodyCell>
				</TableBodyRow>
			{:else}
				<TableBodyRow>
					<TableBodyCell class="text-center" colspan={5}>無提問</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>

	<hr class="my-4" />

	<h2
		class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		管理員回覆
	</h2>

	<Table striped hoverable shadow>
		<TableHead>
			<TableHeadCell>#</TableHeadCell>
			<TableHeadCell>發送者</TableHeadCell>
			<TableHeadCell>接收者</TableHeadCell>
			<TableHeadCell>標題</TableHeadCell>
			<TableHeadCell>時間</TableHeadCell>
		</TableHead>
		<TableBody tableBodyClass="divide-y">
			{#each responses as response, index}
				<TableBodyRow
					class="cursor-pointer"
					onclick={(e) => {
						const target = e.target as HTMLElement;
						if (target.tagName === 'TD') {
							goto(`/dashboard/clarifications/${response.id}`);
						}
					}}
				>
					<TableBodyCell>{index + 1}</TableBodyCell>
					<TableBodyCell>
						<a href="/dashboard/users/{response.senderUser.id}" class="border-b">
							{response.senderUser.name}({response.senderUser.username})
						</a>
					</TableBodyCell>
					<TableBodyCell>
						<a href="/dashboard/users/{response.receiverUser!.id}" class="border-b">
							{response.receiverUser!.name}({response.receiverUser!.username})
						</a>
					</TableBodyCell>
					<TableBodyCell>{response.subject}</TableBodyCell>
					<TableBodyCell>
						{datetimeString(response.createdAt)}
					</TableBodyCell>
				</TableBodyRow>
			{:else}
				<TableBodyRow>
					<TableBodyCell class="text-center" colspan={5}>無管理員回覆</TableBodyCell>
				</TableBodyRow>
			{/each}
		</TableBody>
	</Table>

	<Modal bind:open={openCreateAnnouncementModal} outsideclose title="新增公告">
		<form
			id="create-announcement-form"
			class="space-y-6"
			onsubmit={(e) => {
				e.preventDefault();
				clarificationApp
					.sendClarification(projectId, {
						subject: announcementPayload.subject,
						content: announcementPayload.content,
						sendToAll: true
					})
					.then(() => {
						Swal.fire({
							title: '發布成功',
							icon: 'success',
							showConfirmButton: false,
							timer: 1500
						});
						clarificationApp.fetchClarifications(projectId);
						openCreateAnnouncementModal = false;
					})
					.catch((error) => {
						console.error(error);
						Swal.fire({
							icon: 'error',
							title: '發布失敗',
							text: error.message
						});
					});
			}}
		>
			<div>
				<Label for="subject-input">
					<span class="align-middle text-red-500">*</span>
					標題
				</Label>
				<Input required id="subject-input" name="subject" bind:value={announcementPayload.subject} />
			</div>

			<div>
				<Label for="content-input">內容</Label>
				<Textarea id="content-input" name="content" bind:value={announcementPayload.content} />
			</div>
		</form>
		<svelte:fragment slot="footer">
			<Button color="green" form="create-announcement-form" class="w-full" type="submit">新增</Button>
		</svelte:fragment>
	</Modal>

	<Modal
		bind:open={openShowAnnouncementModal}
		autoclose
		outsideclose
		title="公告: {selectedAnnouncement?.subject}"
	>
		<pre>{selectedAnnouncement?.content || '無內容'}</pre>
		<svelte:fragment slot="footer">
			<Button color="green">確定</Button>
		</svelte:fragment>
	</Modal>
{/if}
