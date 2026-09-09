import { useUserApp } from '@/apps/admin/application/user';
import type { PageLoad } from './$types';
import { useRoleApp } from '@/apps/admin/application/role';

export const load: PageLoad = async (event) => {
	const { id } = event.params;
	const user = await useUserApp().fetchUser(Number(id));
	const roleApp = useRoleApp();
	await roleApp.fetchRoles();
	return { props: { user } };
};
