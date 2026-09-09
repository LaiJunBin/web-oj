import { useProjectApp } from '@/apps/admin/application/project';
import type { PageLoad } from './$types';
import { useGradingItemApp } from '@/apps/admin/application/grading-item';

export const load: PageLoad = async ({ params }) => {
	const { id, gid } = params;
	await Promise.all([
		useProjectApp().fetchProject(Number(id)),
		useGradingItemApp().fetchGradingItem(Number(gid))
	]);
};
