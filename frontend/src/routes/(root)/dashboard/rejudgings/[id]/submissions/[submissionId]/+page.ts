import { useRejudgeApp } from '@/apps/admin/application/rejudge';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const { id, submissionId } = params;
	const rejudgeApp = useRejudgeApp();
	await rejudgeApp.fetchSubmissionComparison(id, submissionId);
};
