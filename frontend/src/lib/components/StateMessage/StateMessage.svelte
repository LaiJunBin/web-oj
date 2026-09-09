<script lang="ts">
	import type { Snippet } from 'svelte';

	// A centered, theme- and kiosk-aware message for the empty / error / "nothing
	// to show yet" states that would otherwise sit on an infinite "Loading...".
	// Purely presentational: callers decide when to render it and what it says.
	const {
		title,
		description = '',
		kiosk = false,
		tone = 'empty',
		action
	}: {
		title: string;
		description?: string;
		kiosk?: boolean;
		tone?: 'empty' | 'error';
		action?: Snippet;
	} = $props();
</script>

<div class="flex h-full w-full flex-1 flex-col items-center justify-center p-8 text-center">
	<h2
		class={[
			'font-bold tracking-tight',
			tone === 'error' ? 'text-red-600 dark:text-red-400' : 'text-gray-900 dark:text-white',
			kiosk ? 'text-5xl' : 'text-2xl'
		]}
	>
		{title}
	</h2>
	{#if description}
		<p class={['mt-3 max-w-prose text-gray-500 dark:text-gray-400', kiosk ? 'text-2xl' : 'text-base']}>
			{description}
		</p>
	{/if}
	{#if action}
		<div class="mt-6">{@render action()}</div>
	{/if}
</div>
