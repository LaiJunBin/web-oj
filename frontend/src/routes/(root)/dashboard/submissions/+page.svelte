<script lang="ts">
	import { goto } from '$app/navigation';
	import { useSubmissionApp } from '@/apps/admin/application/submission';
	import { useSubmissionsStorage } from '@/apps/admin/domain/submission';
	import { adminCurrentProject } from '$lib/adminCurrentProject.svelte';
	import StateMessage from '$lib/components/StateMessage/StateMessage.svelte';
	import { streamContestProjectSubmissions } from '@/lib/api/stream';
	import { datetimeString } from '@/lib';
	import {
		Badge,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';

	const submissionApp = useSubmissionApp();
	const submissionsStorage = useSubmissionsStorage();
	const submissions = $derived(submissionsStorage.get());
	const partialSubmissions = $derived(
		submissions?.filter((submission) => submission.isPartial) ?? []
	);
	const fullSubmissions = $derived(
		submissions?.filter((submission) => !submission.isPartial) ?? []
	);

	const project = $derived(
		adminCurrentProject.projectId != null
			? { id: adminCurrentProject.projectId, name: adminCurrentProject.projectName ?? '' }
			: null
	);
	$effect(() => {
		if (project) {
			submissionApp.withLoading().fetchSubmissions(project.id);
		}
	});

	const projectId = $derived(project?.id);
	$effect(() => {
		if (!projectId) return;
		const sse = streamContestProjectSubmissions(projectId);
		sse.onmessage = (event) => {
			const submission = JSON.parse(event.data);
			const updated = [...(submissionsStorage.get() ?? [])];
			const idx = updated.findIndex((s) => s.id === submission.id);
			if (idx >= 0) {
				updated[idx] = submission;
			} else {
				const insertAt = updated.findIndex(
					(s) => new Date(s.createdAt).getTime() < new Date(submission.createdAt).getTime()
				);
				if (insertAt === -1) {
					updated.push(submission);
				} else {
					updated.splice(insertAt, 0, submission);
				}
			}
			submissionsStorage.set(updated);
		};
		return () => sse.close();
	});
	const statusColorMap: Record<string, 'dark' | 'red' | 'yellow' | 'green' | 'blue'> = {
		Created: 'dark',
		Queued: 'dark',
		Judging: 'blue',
		Judged: 'green',
		Cancelled: 'yellow',
		Error: 'red'
	};
</script>

{#if !project}
	<StateMessage
		title="尚未選擇競賽專案"
		description="請從左側的「競賽過程」選擇比賽與專案。"
	/>
{:else if !submissions}
	<div>Loading...</div>
{:else}
	<div class="flex flex-col p-4">
		<h2 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			{project.name} 的部分提交 ({partialSubmissions.length})
		</h2>
		<Table>
			<TableHead>
				<TableHeadCell>編號</TableHeadCell>
				<TableHeadCell>狀態</TableHeadCell>
				<TableHeadCell>使用者</TableHeadCell>
				<TableHeadCell>得分/總分</TableHeadCell>
				<TableHeadCell>提交時間</TableHeadCell>
			</TableHead>
			<TableBody>
				{#each partialSubmissions as submission, index}
					<TableBodyRow
						class="cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
						onclick={(e) => {
							const target = e.target as HTMLElement;
							if (target.tagName === 'TD') {
								goto(`/dashboard/submissions/${submission.id}`);
							}
						}}
					>
						<TableBodyCell>{partialSubmissions.length - index}</TableBodyCell>
						<TableBodyCell>
							<Badge color={statusColorMap[submission.status]}>{submission.status}</Badge>
						</TableBodyCell>
						<TableBodyCell>
							<a href={`/dashboard/users/${submission.user.id}`}>
								{submission.user.name}({submission.user.username})
							</a>
						</TableBodyCell>
						<TableBodyCell>{submission.score} / {submission.totalScore}</TableBodyCell>
						<TableBodyCell>{datetimeString(submission.createdAt)}</TableBodyCell>
					</TableBodyRow>
				{:else}
					<TableBodyRow>
						<TableBodyCell colspan={5} class="text-center">沒有部分提交</TableBodyCell>
					</TableBodyRow>
				{/each}
			</TableBody>
		</Table>

		<hr class="my-4" />

		<h2 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			{project.name} 的完整提交 ({fullSubmissions.length})
		</h2>

		<Table>
			<TableHead>
				<TableHeadCell>編號</TableHeadCell>
				<TableHeadCell>狀態</TableHeadCell>
				<TableHeadCell>使用者</TableHeadCell>
				<TableHeadCell>得分/總分</TableHeadCell>
				<TableHeadCell>提交時間</TableHeadCell>
				<TableHeadCell></TableHeadCell>
			</TableHead>
			<TableBody>
				{#each fullSubmissions as submission, index}
					<TableBodyRow
						class="cursor-pointer {submission.ignored
							? 'bg-yellow-100 hover:bg-yellow-200 dark:bg-yellow-700 dark:hover:bg-yellow-800'
							: 'hover:bg-gray-100 dark:hover:bg-gray-700'}"
						onclick={() => {
							goto(`/dashboard/submissions/${submission.id}`);
						}}
					>
						<TableBodyCell>{fullSubmissions.length - index}</TableBodyCell>
						<TableBodyCell>
							<Badge color={statusColorMap[submission.status]}>{submission.status}</Badge>
						</TableBodyCell>
						<TableBodyCell>
							<a href={`/dashboard/users/${submission.user.id}`}>
								{submission.user.name}({submission.user.username})
							</a>
						</TableBodyCell>
						<TableBodyCell>{submission.score} / {submission.totalScore}</TableBodyCell>
						<TableBodyCell>{datetimeString(submission.createdAt)}</TableBodyCell>
						<TableBodyCell>
							{#if submission.ignored}
								<Badge color="red">忽略</Badge>
							{/if}
						</TableBodyCell>
					</TableBodyRow>
				{:else}
					<TableBodyRow>
						<TableBodyCell colspan={6} class="text-center">沒有部分提交</TableBodyCell>
					</TableBodyRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
