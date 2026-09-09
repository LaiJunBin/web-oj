import { useProjectApp } from '@/apps/admin/application/project';
import type { PageLoad } from './$types';
import { useTestcaseApp } from '@/apps/admin/application/testcase';
import { useGradingItemApp } from '@/apps/admin/application/grading-item';

export const load: PageLoad = async ({ params }) => {
	const { id, gid } = params;
	await Promise.all([
		useProjectApp().fetchProject(Number(id)),
		useTestcaseApp().fetchProjectTestcases(Number(id)),
		useGradingItemApp().fetchGradingItem(Number(gid))
	]);
};
