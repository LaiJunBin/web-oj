import { useContestApp } from '@/apps/admin/application/contest';
import { useGroupApp } from '@/apps/admin/application/group';
import { useUserApp } from '@/apps/admin/application/user';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const { id } = params;
	await Promise.all([
		useContestApp().fetchContest(Number(id)),
		useContestApp().fetchContestParticipants(Number(id)),
		useUserApp().fetchUsers(),
		useGroupApp().fetchGroups()
	]);
};
