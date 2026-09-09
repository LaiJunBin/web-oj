import { useProjectApp } from '@/apps/admin/application/project';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	await useProjectApp().fetchProjects();
};
