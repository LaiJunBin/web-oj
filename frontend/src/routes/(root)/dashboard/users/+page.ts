import { useUserApp } from '@/apps/admin/application/user';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	const userApp = useUserApp();
	await userApp.fetchUsers();
};
