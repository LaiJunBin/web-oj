import { useImageApp } from '@/apps/admin/application/image';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const { tag } = params;
	const imageApp = useImageApp();
	await imageApp.fetchImage(tag);
};
