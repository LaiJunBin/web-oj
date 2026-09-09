import { useProjectApp } from '@/apps/admin/application/project';
import type { PageLoad } from './$types';
import { useImageApp } from '@/apps/admin/application/image';

export const load: PageLoad = async ({ params }) => {
	const { id } = params;
	await Promise.all([useImageApp().fetchImages(), useProjectApp().fetchProject(Number(id))]);
};
