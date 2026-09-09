import { useUserApp } from '@/apps/admin/application/user';
import type { PageLoad } from './$types';

export const load: PageLoad = async (event) => {
	const { id } = event.params;
	await useUserApp().fetchUser(Number(id));
};
