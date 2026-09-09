<script lang="ts">
	import { writable } from 'svelte/store';

	export let files: string[] | Record<string, any> = [];
	export let selectedFilePath: string = '';
	export let onFileClick: (filePath: string) => void;
	export let parentPath: string = ''; // Track the full path of the current folder or file

	let treeData: Record<string, any> = {};
	let collapsed = writable<Record<string, boolean>>({});

	// Initialize Tree View data
	$: treeData = Array.isArray(files) ? buildTree(files) : files;

	// Build Tree View data structure
	function buildTree(files: string[]): Record<string, any> {
		const tree: Record<string, any> = {};
		files.forEach((file) => {
			const parts = file.split('/');
			let current = tree;
			parts.forEach((part, index) => {
				if (!current[part]) {
					current[part] = index === parts.length - 1 ? null : {};
				}
				current = current[part];
			});
		});
		return tree;
	}

	// Toggle folder collapse
	function toggleCollapse(path: string): void {
		collapsed.update((state) => {
			state[path] = !state[path];
			return state;
		});
	}

	// Sort entries: directories first, then files
	function sortEntries(entries: [string, any][]): [string, any][] {
		return entries.sort(([keyA, valueA], [keyB, valueB]) => {
			if (valueA === null && valueB !== null) return 1; // Files go below directories
			if (valueA !== null && valueB === null) return -1; // Directories go above files
			return keyA.localeCompare(keyB); // Alphabetical order
		});
	}

	// Ensure folders are collapsed by default
	$: collapsed.update((state) => {
		Object.keys(treeData).forEach((key) => {
			const fullPath = parentPath ? `${parentPath}/${key}` : key;
			if (treeData[key] !== null && !(fullPath in state)) {
				state[fullPath] = true; // Default to collapsed
			}
		});
		return state;
	});
</script>

<div class="tree-container">
	{#each sortEntries(Object.entries(treeData)) as [key, value]}
		{#if value === null}
			<!-- File -->
			<button
				type="button"
				class={{
					'tree-node': true,
					'bg-blue-100 text-blue-500':
						selectedFilePath === (parentPath ? `${parentPath}/${key}` : key)
				}}
				on:click={() => onFileClick(parentPath ? `${parentPath}/${key}` : key)}
				aria-label={`File: ${key}`}
			>
				📄 {key}
			</button>
		{:else}
			<!-- Folder -->
			<div>
				<button
					type="button"
					class="tree-node"
					on:click={() => toggleCollapse(parentPath ? `${parentPath}/${key}` : key)}
					on:keydown={(e) =>
						e.key === 'Enter' || e.key === ' '
							? toggleCollapse(parentPath ? `${parentPath}/${key}` : key)
							: null}
					aria-expanded={$collapsed[parentPath ? `${parentPath}/${key}` : key] !== true}
					aria-label={`Folder: ${key}`}
				>
					📁 {key}
				</button>
				{#if $collapsed[parentPath ? `${parentPath}/${key}` : key] !== true}
					<div class="ml-4">
						<svelte:self
							files={value}
							{selectedFilePath}
							{onFileClick}
							parentPath={parentPath ? `${parentPath}/${key}` : key}
						/>
					</div>
				{/if}
			</div>
		{/if}
	{/each}
</div>

<style>
	.tree-node {
		display: block;
		cursor: pointer;
		margin-left: 1rem;
		white-space: nowrap;
		background: none;
		border: none;
		text-align: left;
		padding: 0;
	}
</style>
