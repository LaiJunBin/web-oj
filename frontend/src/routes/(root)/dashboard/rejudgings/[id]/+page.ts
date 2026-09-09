import { useRejudgeApp } from '@/apps/admin/application/rejudge';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const { id } = params;
	const rejudgeApp = useRejudgeApp();
	await rejudgeApp.fetchRejudge(id);
};
