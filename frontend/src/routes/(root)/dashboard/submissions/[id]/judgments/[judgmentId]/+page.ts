import { useSubmissionApp } from '@/apps/admin/application/submission';
import type { PageLoad } from './$types';

// A judgment has no standalone endpoint; its full data lives inside the
// submission detail, so deep links load the submission and pick the judgment.
export const load: PageLoad = async ({ params }) => {
	await useSubmissionApp().fetchSubmission(params.id);
};
