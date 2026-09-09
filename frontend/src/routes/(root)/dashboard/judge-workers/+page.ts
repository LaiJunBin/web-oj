import { useJudgeWorkerApp } from '@/apps/admin/application/judge-worker';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	await useJudgeWorkerApp().fetchJudgeWorkers();
};
