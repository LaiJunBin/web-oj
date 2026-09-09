<script>
	import { Button } from 'flowbite-svelte';
</script>

<section class="flex flex-grow items-center justify-center">
	<div class="mx-auto max-w-screen-xl px-4 py-8 lg:px-6 lg:py-16">
		<div class="mx-auto max-w-screen-sm text-center">
			<h1
				class="mb-4 text-7xl font-bold tracking-tight text-gray-900 md:text-7xl lg:text-9xl dark:text-white"
			>
				404
			</h1>
			<p class="mb-4 text-3xl font-bold tracking-tight text-gray-900 md:text-4xl dark:text-white">
				找不到頁面
			</p>
			<p class="mb-4 text-lg font-light text-gray-500 dark:text-gray-400">
				抱歉，我們找不到您要的頁面。
			</p>

			<Button href="/" size="lg" color="alternative">回首頁</Button>
		</div>
	</div>
</section>
