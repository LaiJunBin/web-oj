import { useSubmissionApp } from '@/apps/admin/application/submission';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const { id } = params;
	const submissionApp = useSubmissionApp();
	await submissionApp.fetchSubmission(id);
};
