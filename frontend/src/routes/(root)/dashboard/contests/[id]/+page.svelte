<script lang="ts">
	import { goto } from '$app/navigation';
	import { refreshContestProject, resetContestState } from '@/lib/refreshContestState';
	import { useContestApp } from '@/apps/admin/application/contest';
	import { useBundleApp } from '@/apps/admin/application/bundle';
	import { askExportMode, showExporting, showExportError } from '@/lib/exportMode';
	import { useContestParticipantsStorage, useContestStorage } from '@/apps/admin/domain/contest';
	import { useGroupsStorage } from '@/apps/admin/domain/group';
	import { useUsersStorage } from '@/apps/admin/domain/user';
	import { datetimeString } from '@/lib';
	import {
		Badge,
		Button,
		ButtonGroup,
		MultiSelect,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell,
		Timeline,
		TimelineItem
	} from 'flowbite-svelte';
	import { StarSolid } from 'flowbite-svelte-icons';
	import Swal from 'sweetalert2';

	const contestApp = useContestApp();
	const bundleApp = useBundleApp();
	const contestStorage = useContestStorage();
	const contest = $derived(contestStorage.get()!);

	/**
	 * 完整匯出要先問清楚映像用哪種形態隨行——兩者差幾百 MB,而且 archive 有
	 * 「所有映像都得先建置完成」的前提,選錯要重來一次。與其猜,不如把差別
	 * 講在選擇畫面上。
	 */
	const exportFullBundle = async () => {
		const mode = await askExportMode(
			'完整匯出',
			'包含這場競賽、它的所有題目,以及題目用到的映像。要選映像以哪種形態隨行:'
		);
		if (!mode) return;
		const done = showExporting();
		bundleApp
			.exportContestBundle(contest.id, mode)
			.then(done)
			// archive 模式失敗時,後端會把還沒建置的映像 tag 列在訊息裡
			.catch((error) => showExportError(error));
	};

	const participants = $derived(useContestParticipantsStorage().get()!);
	const users = $derived(useUsersStorage().get()!);
	const groups = $derived(useGroupsStorage().get()!);

	const participantUsers = $derived(users.filter((user) => participants.userIds.includes(user.id)));
	const participantGroups = $derived(
		groups.filter((group) => participants.groupIds.includes(group.id))
	);
	const candidateUsers = $derived(users.filter((user) => !participants.userIds.includes(user.id)));
	const candidateGroups = $derived(
		groups.filter((group) => !participants.groupIds.includes(group.id))
	);

	let selectedUserIds = $state<number[]>([]);
	let selectedGroupIds = $state<number[]>([]);

	const refreshParticipants = () => {
		// Also refresh the contestant-side project so the shared Navbar's submit
		// button (gated on isParticipant) updates immediately after a participant
		// change, without leaving the dashboard.
		refreshContestProject();
		return contestApp.fetchContestParticipants(contest.id);
	};

	const showError = (title: string) => (error: unknown) => {
		console.error(error);
		Swal.fire({
			title,
			icon: 'error',
			confirmButtonText: '確定'
		});
	};

	const getDuration = (startTime: string, endTime: string) => {
		const total =
			new Date(new Date(endTime).getTime() - new Date(startTime).getTime()).getTime() / 1000;
		const day = Math.floor(total / 86400).toString();
		const hour = Math.floor((total % 86400) / 3600)
			.toString()
			.padStart(2, '0');
		const minute = Math.floor((total % 3600) / 60)
			.toString()
			.padStart(2, '0');
		const second = Math.floor(total % 60).toString().padStart(2, '0');

		if (day !== '0') {
			return `${day} 天 ${hour} 小時 ${minute} 分鐘 ${second} 秒`;
		}

		if (hour === '00') {
			if (minute === '00') {
				return `${second} 秒`;
			}
			return `${minute} 分鐘 ${second} 秒`;
		}

		return `${hour} 小時 ${minute} 分鐘 ${second} 秒`;
	};
</script>

{#if !contest}
	<div>Loading...</div>
{:else}
	<div class="flex flex-col p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/dashboard/contests"
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
				<p>{contest.name}</p>
				<Badge color={contest.isPublic ? 'green' : 'red'} class="ml-2"
					>{contest.isPublic ? '公開' : '非公開'}</Badge
				>
				<ButtonGroup class="ml-auto">
					<Button color="primary" href={`/dashboard/contests/${contest.id}/edit`}>編輯</Button>
					<Button
						color="alternative"
						class="cursor-pointer"
						onclick={() => {
							contestApp
								.exportContest(contest.id)
								.catch((error) =>
									Swal.fire({ icon: 'error', title: '匯出失敗', text: error.message })
								);
						}}>匯出</Button
					>
					<Button color="alternative" class="cursor-pointer" onclick={exportFullBundle}
						>完整匯出</Button
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
										await contestApp.deleteContest(contest.id).then(() => {
											resetContestState();
											Swal.fire({
												title: '刪除成功',
												icon: 'success',
												showConfirmButton: false,
												timer: 1500
											});
										});
										goto('/dashboard/contests');
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

		<div class="my-2">
			<h3 class="text-lg font-bold">描述訊息</h3>
			<p>{contest.description || '無'}</p>
		</div>

		<div class="my-2">
			<h3 class="text-lg font-bold">起訖時間</h3>
			<p>{datetimeString(contest.activateTime)} ~ {datetimeString(contest.deactivateTime)}</p>
		</div>

		<hr class="my-4" />

		<div class="my-2">
			<h2 class="mb-4 flex items-center gap-x-2 text-2xl font-bold">
				<span>參賽者</span>
				{#if contest.isPublic}
					<Badge color="yellow">公開競賽所有登入使用者皆可瀏覽，繳交仍以下方名單為準</Badge>
				{/if}
			</h2>

			<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
				<div>
					<h3 class="mb-2 text-lg font-bold">參賽使用者</h3>
					<div class="mb-4 flex items-end gap-2">
						<div class="grow">
							<MultiSelect
								id="participant-user-select"
								items={candidateUsers.map((user) => ({
									name: `${user.name} (${user.username})`,
									value: user.id
								}))}
								bind:value={selectedUserIds}
							></MultiSelect>
						</div>
						<Button
							color="green"
							class="cursor-pointer"
							disabled={selectedUserIds.length === 0}
							onclick={() => {
								contestApp
									.addUsersToContest(contest.id, selectedUserIds)
									.then(() => {
										selectedUserIds = [];
										refreshParticipants();
									})
									.catch(showError('加入失敗'));
							}}>加入</Button
						>
					</div>
					<Table striped hoverable shadow>
						<TableHead>
							<TableHeadCell>使用者名稱</TableHeadCell>
							<TableHeadCell>名字</TableHeadCell>
							<TableHeadCell>
								<span class="sr-only">Actions</span>
							</TableHeadCell>
						</TableHead>
						<TableBody tableBodyClass="divide-y">
							{#each participantUsers as user}
								<TableBodyRow>
									<TableBodyCell>{user.username}</TableBodyCell>
									<TableBodyCell>{user.name}</TableBodyCell>
									<TableBodyCell>
										<Button
											pill
											color="red"
											class="cursor-pointer"
											onclick={() => {
												contestApp
													.removeUsersFromContest(contest.id, [user.id])
													.then(refreshParticipants)
													.catch(showError('移除失敗'));
											}}>移除</Button
										>
									</TableBodyCell>
								</TableBodyRow>
							{:else}
								<TableBodyRow>
									<TableBodyCell colspan={3} class="text-center">尚無參賽使用者</TableBodyCell>
								</TableBodyRow>
							{/each}
						</TableBody>
					</Table>
				</div>

				<div>
					<h3 class="mb-2 text-lg font-bold">參賽群組</h3>
					<div class="mb-4 flex items-end gap-2">
						<div class="grow">
							<MultiSelect
								id="participant-group-select"
								items={candidateGroups.map((group) => ({
									name: group.name,
									value: group.id
								}))}
								bind:value={selectedGroupIds}
							></MultiSelect>
						</div>
						<Button
							color="green"
							class="cursor-pointer"
							disabled={selectedGroupIds.length === 0}
							onclick={() => {
								contestApp
									.addGroupsToContest(contest.id, selectedGroupIds)
									.then(() => {
										selectedGroupIds = [];
										refreshParticipants();
									})
									.catch(showError('加入失敗'));
							}}>加入</Button
						>
					</div>
					<Table striped hoverable shadow>
						<TableHead>
							<TableHeadCell>名稱</TableHeadCell>
							<TableHeadCell>描述</TableHeadCell>
							<TableHeadCell>
								<span class="sr-only">Actions</span>
							</TableHeadCell>
						</TableHead>
						<TableBody tableBodyClass="divide-y">
							{#each participantGroups as group}
								<TableBodyRow>
									<TableBodyCell>
										<a class="text-primary-600 hover:underline" href={`/dashboard/groups/${group.id}`}
											>{group.name}</a
										>
									</TableBodyCell>
									<TableBodyCell>{group.description || '無'}</TableBodyCell>
									<TableBodyCell>
										<Button
											pill
											color="red"
											class="cursor-pointer"
											onclick={() => {
												contestApp
													.removeGroupsFromContest(contest.id, [group.id])
													.then(refreshParticipants)
													.catch(showError('移除失敗'));
											}}>移除</Button
										>
									</TableBodyCell>
								</TableBodyRow>
							{:else}
								<TableBodyRow>
									<TableBodyCell colspan={3} class="text-center">尚無參賽群組</TableBodyCell>
								</TableBodyRow>
							{/each}
						</TableBody>
					</Table>
				</div>
			</div>
		</div>

		<hr class="my-4" />

		<div class="flex flex-col items-center justify-center">
			<h2 class="mb-4 flex items-center gap-x-2 text-2xl font-bold">
				<span>競賽專案</span>
			</h2>

			<Timeline
				order="vertical"
				onclick={function (e) {
					let element = e.target as HTMLElement;
					while (element !== (this as unknown as HTMLOListElement)) {
						if (element.tagName === 'LI') {
							const index = Array.from(this.children).indexOf(element);
							const project = contest.projects[index];
							goto(`/dashboard/contests/${contest.id}/projects/${project.id}`);
							break;
						}
						element = element.parentElement!;
					}
				}}
			>
				{#each contest.projects as project}
					<TimelineItem
						classLi="border p-2 rounded-lg border-gray-200 dark:border-gray-700 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700"
						classTime="text-lg"
						classH3="text-2xl mb-2"
						classDiv="mt-1"
						title={`${project.name} (競賽時間: ${getDuration(project.startTime, project.endTime)})`}
						date={`${datetimeString(project.startTime) || ''} - ${datetimeString(project.endTime) || ''}`}
					>
						<svelte:fragment slot="icon">
							<span
								class="bg-primary-200 dark:bg-primary-900 absolute -start-3 mt-1 flex h-6 w-6 items-center justify-center rounded-full ring-8 ring-white dark:ring-gray-900"
							>
								<StarSolid class="h-4 w-4" />
							</span>
						</svelte:fragment>
						<p class="mb-4 text-lg font-normal text-gray-500 dark:text-gray-400">
							{project.description}
						</p>
					</TimelineItem>
				{:else}
					<p class="text-center text-2xl text-gray-500 dark:text-gray-400">尚未有競賽日程</p>
				{/each}
			</Timeline>
		</div>
	</div>
{/if}
