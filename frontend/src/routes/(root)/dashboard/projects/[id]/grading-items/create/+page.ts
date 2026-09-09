import { useProjectApp } from '@/apps/admin/application/project';
import type { PageLoad } from './$types';
import { useTestcaseApp } from '@/apps/admin/application/testcase';

export const load: PageLoad = async ({ params }) => {
	const { id } = params;
	await Promise.all([
		useProjectApp().fetchProject(Number(id)),
		useTestcaseApp().fetchProjectTestcases(Number(id))
	]);
};
