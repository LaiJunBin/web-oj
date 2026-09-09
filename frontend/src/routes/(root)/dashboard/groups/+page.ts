import { useGroupApp } from '@/apps/admin/application/group';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	const groupApp = useGroupApp();
	await groupApp.fetchGroups();
};
