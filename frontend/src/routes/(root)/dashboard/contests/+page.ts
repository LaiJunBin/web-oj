import { useContestApp } from '@/apps/admin/application/contest';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	await useContestApp().fetchContests();
};
