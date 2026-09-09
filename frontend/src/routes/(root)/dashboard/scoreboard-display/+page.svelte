<script lang="ts">
	import { page } from '$app/state';
	import { adminCurrentProject } from '$lib/adminCurrentProject.svelte';
	import StateMessage from '$lib/components/StateMessage/StateMessage.svelte';
	import { Alert, Button, Input, Label } from 'flowbite-svelte';
	import { InfoCircleSolid } from 'flowbite-svelte-icons';

	// 為側欄選定的場次產生 kiosk 記分板連結,免去在場地機器手打網址。
	// 連結帶 contestId/projectId 鎖定場次:貼到任何機器都顯示同一場,
	// 場次不存在時 kiosk 端會直接顯示錯誤(不會改播別場)。
	let text = $state('');
	let footer = $state('');
	let copied = $state(false);

	const url = $derived.by(() => {
		if (adminCurrentProject.contestId == null || adminCurrentProject.projectId == null) {
			return '';
		}
		const params = new URLSearchParams();
		params.set('contestId', String(adminCurrentProject.contestId));
		params.set('projectId', String(adminCurrentProject.projectId));
		if (text.trim()) params.set('text', text.trim());
		if (footer.trim()) params.set('footer', footer.trim());
		return `${page.url.origin}/scoreboard?kiosk&${params.toString()}`;
	});

	let copiedTimer: ReturnType<typeof setTimeout>;
	async function copyUrl() {
		// clipboard API 在 http(場地內網常見)或未授權時「存在但會 reject」,
		// 所以任何失敗都要退回舊式複製,不能只看 API 存不存在。
		let done = false;
		try {
			if (navigator.clipboard?.writeText) {
				await navigator.clipboard.writeText(url);
				done = true;
			}
		} catch {
			// fall through to the legacy path below
		}
		if (!done) {
			try {
				const ta = document.createElement('textarea');
				ta.value = url;
				document.body.appendChild(ta);
				ta.select();
				done = document.execCommand('copy');
				ta.remove();
			} catch (e) {
				console.error('Failed to copy kiosk url', e);
			}
		}
		if (done) {
			copied = true;
			clearTimeout(copiedTimer);
			copiedTimer = setTimeout(() => (copied = false), 2000);
		}
	}
</script>

{#if adminCurrentProject.projectId == null}
	<StateMessage
		title="尚未選擇競賽專案"
		description="請從左側的「競賽過程」選擇比賽與專案。"
	/>
{:else}
	<div class="max-w-3xl p-4">
		<h1 class="mb-4 border-b-2 border-gray-200 pb-4 text-2xl font-bold">記分板展示</h1>
		<p class="mb-4 text-gray-600 dark:text-gray-400">
			為場地大螢幕產生記分板連結:複製後貼到場地機器的瀏覽器即可,無須手打網址。
		</p>

		<div class="mb-4 rounded-lg border border-gray-200 p-4 dark:border-gray-700">
			<p class="text-sm text-gray-500 dark:text-gray-400">將顯示的場次(由左側選單決定)</p>
			<p class="mt-1 font-bold text-gray-900 dark:text-white">
				{adminCurrentProject.contestName} — {adminCurrentProject.projectName}
			</p>
		</div>

		<div class="mb-4 flex flex-col gap-4">
			<div>
				<Label for="kiosk-text" class="mb-2">顯示文字(選填,顯示於標題下方)</Label>
				<Input id="kiosk-text" bind:value={text} placeholder="例:第一題:靜態網頁" />
			</div>
			<div>
				<Label for="kiosk-footer" class="mb-2">底部公告(選填,固定於畫面底部)</Label>
				<Input id="kiosk-footer" bind:value={footer} placeholder="例:請保持安靜,禁止交談" />
			</div>
		</div>

		<div class="mb-4">
			<p class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">產生的網址</p>
			<code
				class="block rounded-lg bg-gray-100 p-3 text-sm break-all text-gray-700 dark:bg-gray-900 dark:text-gray-300"
			>
				{url}
			</code>
		</div>

		<div class="mb-6 flex gap-2">
			<Button color="primary" class="cursor-pointer" onclick={copyUrl}>
				{copied ? '已複製!' : '複製連結'}
			</Button>
			<Button
				color="alternative"
				class="cursor-pointer"
				onclick={() => window.open(url, '_blank')}
			>
				在新分頁開啟
			</Button>
		</div>

		<Alert color="blue" class="dark:bg-blue-900 dark:text-blue-100">
			<InfoCircleSolid slot="icon" class="h-5 w-5" />
			公開比賽的大螢幕<strong>不需登入</strong>;貼上網址後按
			<strong>F11</strong> 全螢幕。若網址指定的場次不存在,畫面會直接顯示錯誤,不會改播其他場次。
		</Alert>
	</div>
{/if}
