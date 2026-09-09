import tailwindcss from '@tailwindcss/vite';
import { svelteTesting } from '@testing-library/svelte/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { type Plugin } from 'vite';
import { defineConfig } from 'vitest/config';

function extractDepTypes(params: string): string[] {
	if (!params.trim()) return [];
	return params.split(',').flatMap((param) => {
		const trimmed = param.trim();
		if (!trimmed) return [];
		// Mirror the dev fallback: extract parameter name (strip modifiers, ignore type annotation)
		// and capitalize first letter. e.g. "private userStorage: UserStorageInterface" → "UserStorage"
		const nameMatch = /(?:(?:public|private|protected|readonly)\s+)*([\w$]+)/.exec(trimmed);
		if (!nameMatch) return [];
		const name = nameMatch[1];
		return [name.charAt(0).toUpperCase() + name.slice(1)];
	});
}

function diDepsPlugin(): Plugin {
	return {
		name: 'di-deps',
		enforce: 'pre',
		apply: 'build',
		transform(code, id) {
			if (!id.includes('/src/') || !/\.[jt]sx?$/.test(id)) return null;
			if (!code.includes('constructor(')) return null;
			let modified = false;
			const result = code.replace(
				/\bconstructor\s*\(([^)]*)\)\s*\{/g,
				(match, params) => {
					const types = extractDepTypes(params);
					if (types.length === 0) return match;
					modified = true;
					return `static __deps = ${JSON.stringify(types)};\n\t${match}`;
				}
			);
			return modified ? result : null;
		}
	};
}

export default defineConfig({
	plugins: [diDepsPlugin(), sveltekit(), tailwindcss()],

	test: {
		projects: [
			{
				extends: './vite.config.ts',
				plugins: [svelteTesting()],

				test: {
					name: 'client',
					environment: 'jsdom',
					env: { TZ: 'Asia/Taipei' },
					clearMocks: true,
					include: ['src/**/*.svelte.{test,spec}.{js,ts}'],
					exclude: ['src/lib/server/**'],
					setupFiles: ['./vitest-setup-client.ts']
				}
			},
			{
				extends: './vite.config.ts',

				test: {
					name: 'server',
					environment: 'jsdom',
					env: { TZ: 'Asia/Taipei' },
					include: ['src/**/*.{test,spec}.{js,ts}'],
					exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	},

	build: {
		sourcemap: true
	},

	esbuild: {
		keepNames: true
	},

	preview: {
		host: true,
		port: 4173
	},

	resolve: {
		alias: {
			$lib: '/src/lib',
			'@': '/src',
			'@/*': '/src/*'
		}
	}
});
