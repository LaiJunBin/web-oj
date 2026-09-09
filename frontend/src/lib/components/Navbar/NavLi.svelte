<script>
	import { NavLi } from 'flowbite-svelte';

	const { children, href = '' } = $props();
</script>

<NavLi
	class="block cursor-pointer rounded-sm text-gray-700 hover:bg-gray-100 md:py-2 md:ps-3 md:pe-4 md:hover:text-blue-700 lg:border-0 lg:p-0 lg:hover:bg-transparent dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-white lg:dark:hover:bg-transparent lg:dark:hover:text-white"
	{href}
>
	{@render children()}
</NavLi>
