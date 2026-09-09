import { useContestApp } from '@/apps/admin/application/contest';
import type { PageLoad } from './$types';
import { useContestProjectApp } from '@/apps/admin/application/contest-project';

export const load: PageLoad = async ({ params }) => {
	const { id, pid } = params;
	await Promise.all([
		useContestApp().fetchContest(Number(id)),
		useContestProjectApp().fetchContestProject(Number(pid))
	]);
};
