import { useProjectApp } from '@/apps/admin/application/project';
import type { PageLoad } from './$types';
import { useJudgeWorkerApp } from '@/apps/admin/application/judge-worker';

export const load: PageLoad = async () => {
	await Promise.all([
		useProjectApp().fetchProjects(),
		useJudgeWorkerApp().fetchJudgePriorityStrategies(),
		// 專案欄位留空時套用的伺服器設定值,用來把 placeholder 的「預設」講成實際數字
		useJudgeWorkerApp().fetchSystemDefaults()
	]);
};
