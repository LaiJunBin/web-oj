<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { useClarificationApp } from '@/apps/admin/application/clarification';
	import {
		useClarificationStorage,
		type ClarificationDetail,
		type ReplyClarificationPayload
	} from '@/apps/admin/domain/clarification';
	import { datetimeString } from '@/lib';
	import {
		Accordion,
		AccordionItem,
		Button,
		Card,
		Checkbox,
		Input,
		Label,
		Textarea
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const clarificationApp = useClarificationApp();
	const id = $derived(Number(page.params.id));
	const clarification = $derived(useClarificationStorage().get());

	function flattenReplies(reply: ClarificationDetail): ClarificationDetail[] {
		let replies: ClarificationDetail[] = [];

		if (reply.replyClarification) {
			replies.push(reply.replyClarification);
			replies = replies.concat(flattenReplies(reply.replyClarification));
		}

		return replies;
	}

	let replies = $derived(clarification ? flattenReplies(clarification) : []);
	let payload: ReplyClarificationPayload = $state({
		subject: '',
		content: '',
		sendToAll: false
	});

	$effect(() => {
		clarificationApp.fetchClarification(id).then(() => {
			payload = {
				subject: `解答: ${clarification?.subject}`,
				content: '',
				sendToAll: false
			};
		});
	});
</script>

{#if !clarification}
	<p>Loading...</p>
{:else}
	<div class="p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/dashboard/clarifications"
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
				<p>{clarification.subject}</p>
			</div>
		</h1>

		<p class="my-4 flex items-center gap-2 text-gray-500 dark:text-gray-400">
			<span>建立時間</span>
			<span>{datetimeString(clarification.createdAt)}</span>
		</p>

		<Accordion multiple>
			<AccordionItem
				class={replies.length
					? 'cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700'
					: 'pointer-events-none'}
			>
				<span slot="header">先前的內容 ({replies.length})</span>
				<section class="space-y-6">
					<div class="space-y-4">
						{#each replies as reply}
							<Card class="block max-w-none">
								<h3 class="text-xl font-semibold text-gray-800 dark:text-gray-200">
									{reply.subject}
								</h3>
								<p class="mt-2 text-gray-800 dark:text-gray-100">{reply.content}</p>
							</Card>

							<hr class="my-6 border-t-2 border-gray-300 dark:border-gray-600" />
						{/each}
					</div>
				</section>
			</AccordionItem>

			<AccordionItem open class="pointer-events-none">
				<span slot="header">內容</span>
				<pre>{clarification.content}</pre>
			</AccordionItem>

			<AccordionItem open class="pointer-events-none">
				<span slot="header">回覆</span>
				<section>
					<form
						class="space-y-6"
						onsubmit={(e) => {
							e.preventDefault();
							clarificationApp
								.replyClarification(clarification.id, payload)
								.then((id) => {
									Swal.fire({
										icon: 'success',
										title: '回覆成功'
									});
									goto(`/dashboard/clarifications/${id}`);
								})
								.catch((err) => {
									console.error(err);
									Swal.fire({
										icon: 'error',
										title: '回覆失敗'
									});
								});
						}}
					>
						<div>
							<Label for="subject-input">
								<span class="align-middle text-red-500">*</span>
								標題
							</Label>
							<Input required id="subject-input" name="subject" bind:value={payload.subject} />
						</div>

						<div>
							<Label for="content-input">
								<span class="align-middle text-red-500">*</span>
								內容
							</Label>

							<Textarea
								id="content-input"
								name="content"
								class="h-32"
								required
								bind:value={payload.content}
							/>
						</div>

						<div>
							<Label for="send-to-all-checkbox" class="flex items-center">
								<Checkbox
									id="send-to-all-checkbox"
									name="sendToAll"
									bind:checked={payload.sendToAll}
								/>
								<span class="ml-2">發送給所有人</span>
							</Label>
						</div>

						<Button color="green" type="submit" class="w-full">回覆</Button>
					</form>
				</section>
			</AccordionItem>
		</Accordion>
	</div>
{/if}
