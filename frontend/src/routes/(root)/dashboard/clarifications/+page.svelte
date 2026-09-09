<script lang="ts">
	import { adminCurrentProject } from '$lib/adminCurrentProject.svelte';
	import StateMessage from '$lib/components/StateMessage/StateMessage.svelte';
	import ClarificationsPanel from '$lib/components/admin/ClarificationsPanel.svelte';
</script>

{#if adminCurrentProject.projectId == null}
	<StateMessage
		title="尚未選擇競賽專案"
		description="請從左側的「競賽過程」選擇比賽與專案。"
	/>
{:else}
	<div class="p-4">
		<ClarificationsPanel
			projectId={adminCurrentProject.projectId}
			projectName={adminCurrentProject.projectName ?? ''}
		/>
	</div>
{/if}
