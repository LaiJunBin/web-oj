<script lang="ts">
	import { onMount } from 'svelte';
	import { sanitizeHtml } from '$lib/sanitize';

	let { html }: { html: string } = $props();

	let randomId = 'shadow-' + Math.random().toString(36).substring(7);
	let ref: HTMLElement | null = $state(null);
	let div: HTMLElement | null = $state(null);

	onMount(() => {
		if (!div) {
			div = document.createElement('div');
			div.id = randomId;
			div.attachShadow({ mode: 'open' });
		}

		if (ref && div) {
			// Sanitize, then inject. This component used to EXTRACT <script>
			// blocks from the html and re-inject them as live, executing <script>
			// elements — a stored-XSS sink the moment any user's output was shown
			// to another user. Now the markup is sanitized (no <script>, no
			// on*=, no javascript: URLs) and set once; the shadow root still
			// scopes styling, but nothing here executes.
			div.shadowRoot!.innerHTML = sanitizeHtml(html);
			ref.replaceWith(div);
		}
	});
</script>

<div bind:this={ref}></div>
