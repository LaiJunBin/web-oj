import { useProjectApp } from '@/apps/admin/application/project';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const { id } = params;
	await useProjectApp().fetchProject(Number(id));
};
