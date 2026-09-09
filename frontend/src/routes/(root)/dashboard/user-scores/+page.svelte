<script lang="ts">
	import { adminCurrentProject } from '$lib/adminCurrentProject.svelte';
	import StateMessage from '$lib/components/StateMessage/StateMessage.svelte';
	import UserScoresPanel from '$lib/components/admin/UserScoresPanel.svelte';
</script>

{#if adminCurrentProject.projectId == null}
	<StateMessage
		title="尚未選擇競賽專案"
		description="請從左側的「競賽過程」選擇比賽與專案。"
	/>
{:else}
	<div class="flex flex-col p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">
			競賽: {adminCurrentProject.contestName}
		</h1>
		<UserScoresPanel
			projectId={adminCurrentProject.projectId}
			projectName={adminCurrentProject.projectName ?? ''}
		/>
	</div>
{/if}
