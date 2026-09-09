<script lang="ts">
	import Loading from '$lib/components/Loading/Loading.svelte';
	import '../app.css';
	let { children } = $props();
</script>

<Loading />
{@render children()}
