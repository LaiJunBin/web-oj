import { useContestProjectApp } from '@/apps/admin/application/contest-project';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const { pid } = params;
	await useContestProjectApp().fetchJudgmentAnalytics(Number(pid));
};
