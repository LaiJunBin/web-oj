<script lang="ts">
	import { useUserStorage } from '@/apps/admin/domain/user';
	import {
		Badge,
		Button,
		ButtonGroup,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import { useUserApp } from '@/apps/admin/application/user';
	import { isSelf, leaveAfterSelfDelete } from '$lib/profileRefresh';
	import Swal from 'sweetalert2';
	import { goto } from '$app/navigation';
	import { datetimeString } from '@/lib';

	const userApp = useUserApp();
	const user = $derived(useUserStorage().get());
	const statusColorMap: Record<string, 'dark' | 'red' | 'yellow' | 'green' | 'blue'> = {
		Created: 'dark',
		Queued: 'dark',
		Judging: 'blue',
		Judged: 'green',
		Cancelled: 'yellow',
		Error: 'red'
	};

</script>

{#if !user}
	<div>Loading...</div>
{:else}
	<div class="flex h-full flex-col p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/dashboard/users"
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
				<p>使用者: {user.name} ({user.username})</p>
				<ButtonGroup class="ml-auto">
					<Button color="primary" href={`/dashboard/users/${user.id}/edit`}>編輯</Button>
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
										await userApp.deleteUser(user.id);
										// 刪的是自己:這個 session 已經死了,回列表只會 401。
										if (isSelf(user.username)) return leaveAfterSelfDelete();
										Swal.fire({
											title: '刪除成功',
											icon: 'success',
											showConfirmButton: false,
											timer: 1500
										});
										goto('/dashboard/users');
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

		<h2 class="my-2 text-lg font-bold">角色</h2>
		<div class="flex flex-wrap gap-2">
			{#each user.roles as role}
				<div class="rounded-lg bg-gray-100 px-4 py-2 dark:bg-gray-700">{role}</div>
			{/each}
		</div>

		<div class="mt-8 flex h-0 flex-grow flex-col">
			<h2 class="mb-2 flex items-center space-x-2 text-lg font-bold">
				<p>提交紀錄</p>
			</h2>

			<Table>
				<TableHead>
					<TableHeadCell>編號</TableHeadCell>
					<TableHeadCell>狀態</TableHeadCell>
					<TableHeadCell>競賽</TableHeadCell>
					<TableHeadCell>專案</TableHeadCell>
					<TableHeadCell>使用者</TableHeadCell>
					<TableHeadCell>得分/總分</TableHeadCell>
					<TableHeadCell>提交時間</TableHeadCell>
				</TableHead>
				<TableBody>
					{#each user.submissions as submission, index}
						<TableBodyRow
							class="cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
							onclick={(e) => {
								const target = e.target as HTMLElement;
								if (target.tagName === 'TD') {
									goto(`/dashboard/submissions/${submission.id}`);
								}
							}}
						>
							<TableBodyCell>{user.submissions.length - index}</TableBodyCell>
							<TableBodyCell>
								<Badge color={statusColorMap[submission.status]}>{submission.status}</Badge>
							</TableBodyCell>
							<TableBodyCell>
								<a href={`/dashboard/contests/${submission.contest.id}`} class="border-b">
									{submission.contest.name}
								</a>
							</TableBodyCell>
							<TableBodyCell>
								<a href={`/dashboard/projects/${submission.project.id}`} class="border-b">
									{submission.project.name}
								</a>
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
							<TableBodyCell colspan={7} class="text-center">沒有提交紀錄</TableBodyCell>
						</TableBodyRow>
					{/each}
				</TableBody>
			</Table>
		</div>
	</div>
{/if}
