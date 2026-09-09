import { useGradingItemApp } from '@/apps/admin/application/grading-item';
import type { PageLoad } from './$types';
import { useProjectApp } from '@/apps/admin/application/project';

export const load: PageLoad = async ({ params }) => {
	const { id } = params;
	await Promise.all([
		useProjectApp().fetchProject(Number(id)),
		useGradingItemApp().fetchGradingItems(Number(id))
	]);
};
