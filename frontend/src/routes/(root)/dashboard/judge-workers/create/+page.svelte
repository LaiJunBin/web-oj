<script lang="ts">
	import { goto } from '$app/navigation';
	import { useJudgeWorkerApp } from '@/apps/admin/application/judge-worker';
	import {
		JudgeWorkerOrders,
		type CreateJudgeWorkerPayload
	} from '@/apps/admin/domain/judge-worker';
	import { Button, Label, Select } from 'flowbite-svelte';
	import Swal from 'sweetalert2';

	const judgeWorkerApp = useJudgeWorkerApp();
	const payload: CreateJudgeWorkerPayload = $state({
		judgeOrder: Array(Object.values(JudgeWorkerOrders).length).fill('')
	});
</script>

<div class="p-4">
	<h1
		class="mb-4 flex items-center space-x-2 border-b-2 border-gray-200 pb-4 text-2xl font-bold text-gray-900 dark:text-white"
	>
		<span>新增評分工作者</span>
	</h1>

	<form
		id="submit-form"
		class="space-y-6"
		onsubmit={async (e) => {
			e.preventDefault();
			judgeWorkerApp.createJudgeWorker(payload).then(() => {
				Swal.fire({
					title: '新增成功',
					icon: 'success',
					showConfirmButton: false,
					timer: 1500
				});
				goto(`/dashboard/judge-workers`);
			});
		}}
	>
		{#each Object.values(JudgeWorkerOrders) as order, i}
			<div id="judge-order-group">
				<Label for={`judge-order-select-${i}`}>
					<span class="align-middle text-red-500">*</span>
					評分順序 {i + 1}
				</Label>

				<Select
					id={`judge-order-select-${i}`}
					required
					placeholder=""
					bind:value={payload.judgeOrder[i]}
				>
					<option value="" disabled={!payload.judgeOrder[i]}>請選擇</option>
					{#each Object.values(JudgeWorkerOrders) as order}
						{#if payload.judgeOrder.includes(order)}
							<option value={order} disabled>{order}</option>
						{:else}
							<option value={order}>{order}</option>
						{/if}
					{/each}
				</Select>
			</div>
		{/each}
		<Button color="green" class="cursor-pointer" type="submit">新增</Button>
	</form>
</div>
