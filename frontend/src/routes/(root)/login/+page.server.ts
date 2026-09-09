import { redirect, type Actions } from '@sveltejs/kit';
import { apiClient } from '@/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
	if (cookies.get('access_token')) {
		redirect(302, '/');
	}

	return {};
};

export const actions = {
	default: async ({ request }) => {
		const data = await request.formData();
		const username = data.get('username') as string;
		const password = data.get('password') as string;
		const response = await apiClient.login({
			username,
			password
		});

		return response;
	}
} satisfies Actions;
