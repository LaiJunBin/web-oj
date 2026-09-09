import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ route }) => {
	const { id } = route;
	if (id === '/') {
		redirect(302, '/overview');
	}
};
