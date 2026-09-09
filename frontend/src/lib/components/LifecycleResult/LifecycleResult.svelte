<script lang="ts">
	import { Badge } from 'flowbite-svelte';
	import { hasFailed, lineCount, type LifecycleRun } from '$lib/judgmentLifecycle';

	// One lifecycle step's outcome. A non-zero exit code is the whole story of a
	// broken judgment, so it is the loudest thing here; stdout/stderr are
	// collapsed — hundreds of steps fully expanded are unreadable — except
	// stderr of a failed step, which opens by default because it is the answer
	// to "what went wrong".
	let { run, title }: { run: LifecycleRun; title: string } = $props();

	const failed = $derived(hasFailed(run));
	const stderrLines = $derived(lineCount(run.result.stderr));
	const stdoutLines = $derived(lineCount(run.result.stdout));
	const countLabel = (n: number) => (n === 0 ? '(空)' : ` (${n} 行)`);
</script>

<div
	class={[
		'rounded-md border p-3',
		failed
			? 'border-red-300 bg-red-50 dark:border-red-700 dark:bg-red-950'
			: 'border-gray-200 dark:border-gray-700'
	]}
>
	<div class="flex flex-wrap items-center gap-3">
		<span class="font-bold">{title}</span>
		{#if failed}
			<Badge color="red" class="font-bold">失敗 · exit {run.result.exitCode}</Badge>
		{:else}
			<Badge color="green">exit 0</Badge>
		{/if}
		<span class="text-sm text-gray-500 dark:text-gray-400">{run.result.execTime} ms</span>
	</div>
	<div class="mt-2 flex items-start gap-2">
		<span class="shrink-0 font-bold">指令：</span>
		<code class="break-all">{run.command}</code>
	</div>
	<details class="mt-2" open={failed && stderrLines > 0}>
		<summary class="cursor-pointer font-bold select-none">stderr{countLabel(stderrLines)}</summary>
		<code
			class="mt-1 block max-h-60 overflow-auto rounded-md bg-gray-100 p-2 whitespace-pre text-gray-600 dark:bg-gray-800 dark:text-gray-300"
			>{run.result.stderr}</code
		>
	</details>
	<details class="mt-2">
		<summary class="cursor-pointer font-bold select-none">stdout{countLabel(stdoutLines)}</summary>
		<code
			class="mt-1 block max-h-60 overflow-auto rounded-md bg-gray-100 p-2 whitespace-pre text-gray-600 dark:bg-gray-800 dark:text-gray-300"
			>{run.result.stdout}</code
		>
	</details>
</div>
