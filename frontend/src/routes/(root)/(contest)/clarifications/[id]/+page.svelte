<script lang="ts">
	import Swal from 'sweetalert2';
	import { Accordion, AccordionItem, Button, Card, Label, Textarea } from 'flowbite-svelte';
	import { type ClarificationDetail } from '@/apps/contest/domain/clarification.js';
	import { datetimeString } from '$lib';
	import { goto } from '$app/navigation';
	import { useClarificationApp } from '@/apps/contest/application/clarification.js';

	const { data } = $props();
	if (!data.clarification) {
		Swal.fire({
			icon: 'error',
			title: '錯誤',
			text: '找不到該提問'
		}).then(() => {
			goto('/clarifications');
		});
	}

	const clarificationApp = useClarificationApp();

	function flattenReplies(reply: ClarificationDetail): ClarificationDetail[] {
		let replies: ClarificationDetail[] = [];

		if (reply.replyClarification) {
			replies.push(reply.replyClarification);
			replies = replies.concat(flattenReplies(reply.replyClarification));
		}

		return replies;
	}

	let replies = $derived(flattenReplies(data.clarification));
	let content = $state('');
</script>

{#if !data.clarification}
	<p>Loading...</p>
{:else}
	<div class="p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/clarifications"
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
				<p>{data.clarification.subject}</p>
			</div>
		</h1>

		<p class="my-4 flex items-center gap-2 text-gray-500 dark:text-gray-400">
			<span>建立時間</span>
			<span>{datetimeString(data.clarification.createdAt)}</span>
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
				<pre>{data.clarification.content}</pre>
			</AccordionItem>

			<AccordionItem open class="pointer-events-none">
				<span slot="header">回覆</span>
				<section>
					<form
						id="submit-form"
						class="space-y-6"
						onsubmit={(e) => {
							e.preventDefault();
							clarificationApp
								.replyClarification(data.clarification.id, content)
								.then((id) => {
									Swal.fire({
										icon: 'success',
										title: '回覆成功'
									});
									content = '';
									goto(`/clarifications/${id}`);
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
							<Label for="content-input">
								<span class="align-middle text-red-500">*</span>
								內容
							</Label>

							<Textarea
								id="content-input"
								name="content"
								class="h-32"
								required
								bind:value={content}
							/>
						</div>

						<Button color="green" type="submit" form="submit-form" class="w-full">回覆</Button>
					</form>
				</section>
			</AccordionItem>
		</Accordion>
	</div>
{/if}
