import type { PageLoad } from './$types';
import { useImageApp } from '@/apps/admin/application/image';

export const load: PageLoad = async () => {
	await useImageApp().fetchImages();
};
