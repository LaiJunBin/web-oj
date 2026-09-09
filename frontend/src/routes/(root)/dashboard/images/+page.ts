import { useImageApp } from '@/apps/admin/application/image';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	const imageApp = useImageApp();
	await imageApp.fetchImages();
};
