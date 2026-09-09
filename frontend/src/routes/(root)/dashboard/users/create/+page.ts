import { useRoleApp } from '@/apps/admin/application/role';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	const roleApp = useRoleApp();
	await roleApp.fetchRoles();
};
