import '@testing-library/jest-dom/vitest';
import { render, screen } from '@testing-library/svelte';
// import Loading from './Loading.svelte';
import { describe, expect, test, vi } from 'vitest';
import Page from '@/routes/+page.svelte';

// vi.mock('@/apps/common/domain/loading', () => ({
// 	useLoadingStorage: vi.fn(() => ({
// 		isLoading: true
// 	}))
// }));

describe('Loading.svelte', () => {
	test('should display loading spinner when isLoading is true', async () => {
		render(Page);
		// await waitFor(() => {
		// 	expect(container.querySelector('div[hidden=false]')).toBeInTheDocument();
		// });
	});

	// test('should hide loading spinner when isLoading is false', async () => {
	// 	vi.mock('@/apps/common/domain/loading', () => ({
	// 		useLoadingStorage: vi.fn(() => ({
	// 			isLoading: false
	// 		}))
	// 	}));

	// 	const { container } = render(Loading);
	// 	await waitFor(() => {
	// 		expect(container.querySelector('div[hidden=true]')).toBeInTheDocument();
	// 	});
	// });
});
