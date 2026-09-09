<script lang="ts">
	import { goto } from '$app/navigation';
	import { useTestcaseApp } from '@/apps/admin/application/testcase';
	import { useGradingItemStorage } from '@/apps/admin/domain/grading-item';
	import { useProjectStorage } from '@/apps/admin/domain/project';
	import { useTestcaseStorage } from '@/apps/admin/domain/testcase';
	import {
		Button,
		ButtonGroup,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell
	} from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const testcaseApp = useTestcaseApp();
	const gradingItemStorage = useGradingItemStorage();
	const testcaseStorage = useTestcaseStorage();
	const gradingItem = $derived(gradingItemStorage.get());
	const project = $derived(useProjectStorage().get());
	const testcase = $derived(testcaseStorage.get());

	const formatGradingItemName = (item: { label?: string; name: string }) =>
		item.label ? item.label + ": " + item.name : item.name;
</script>

{#if !project || !gradingItem || !testcase}
	<div>Loading...</div>
{:else}
	<div class="flex flex-col p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			<div class="flex items-center gap-2">
				<Button
					href="/dashboard/projects/{project.id}/grading-items/{gradingItem.id}"
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
				<p>評分項目 {formatGradingItemName(gradingItem)} 的測試案例 {testcase.name}</p>
				<ButtonGroup class="ml-auto">
					<Button
						color="primary"
						href={`/dashboard/projects/${project.id}/grading-items/${gradingItem.id}/testcases/${testcase.id}/edit`}
						>編輯</Button
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
										await testcaseApp.deleteTestcase(testcase.id).then(() => {
											Swal.fire({
												title: '刪除成功',
												icon: 'success',
												showConfirmButton: false,
												timer: 1500
											});
										});
										goto(`/dashboard/projects/${project.id}/grading-items/${gradingItem.id}`);
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

		<div class="grid grid-cols-2 gap-4">
			<div>
				<h2 class="text-lg font-bold">描述訊息</h2>
				<p>{gradingItem.description || '無'}</p>
			</div>
			<div>
				<h2 class="text-lg font-bold">分數</h2>
				<p>{gradingItem.score}</p>
			</div>
			<div>
				<h2 class="text-lg font-bold">指令</h2>
				<p>{testcase.command}</p>
			</div>
			<div>
				<h2 class="text-lg font-bold">是否為範例測資</h2>
				<p>{testcase.isSample ? '是' : '否'}</p>
			</div>
		</div>

		<hr class="my-4" />

		<h2 class="text-lg font-bold">依賴的測試案例</h2>
		<div>
			<Table hoverable>
				<TableHead>
					<TableHeadCell>#</TableHeadCell>
					<TableHeadCell>Grading Item Name</TableHeadCell>
					<TableHeadCell>Testcase Name</TableHeadCell>
				</TableHead>
				<TableBody tableBodyClass="divide-y">
					{#each testcase.depends as tc, index}
						<TableBodyRow
							class="cursor-pointer"
							onclick={() => {
								goto(
									`/dashboard/projects/${project.id}/grading-items/${tc.gradingItem.id}/testcases/${tc.id}`
								);
							}}
						>
							<TableBodyCell>
								{index + 1}
							</TableBodyCell>
							<TableBodyCell>
								{formatGradingItemName(tc.gradingItem)}
							</TableBodyCell>
							<TableBodyCell>
								{tc.name}
							</TableBodyCell>
						</TableBodyRow>
					{:else}
						<TableBodyRow>
							<TableBodyCell colspan={3} class="text-center">無</TableBodyCell>
						</TableBodyRow>
					{/each}
				</TableBody>
			</Table>
		</div>
	</div>
{/if}
