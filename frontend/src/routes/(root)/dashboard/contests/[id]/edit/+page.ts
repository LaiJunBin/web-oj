import { useProjectApp } from '@/apps/admin/application/project';
import type { PageLoad } from './$types';
import { useJudgeWorkerApp } from '@/apps/admin/application/judge-worker';
import { useContestApp } from '@/apps/admin/application/contest';

export const load: PageLoad = async ({ params }) => {
	const { id } = params;
	await Promise.all([
		useProjectApp().fetchProjects(),
		useJudgeWorkerApp().fetchJudgePriorityStrategies(),
		// 專案欄位留空時套用的伺服器設定值,用來把 placeholder 的「預設」講成實際數字
		useJudgeWorkerApp().fetchSystemDefaults(),
		useContestApp().fetchContest(Number(id))
	]);
};
