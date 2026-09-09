import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
	const { id } = params;
	try {
		const response = await fetch(`${import.meta.env.VITE_API_URL}/user/clarifications/${id}`, {
			credentials: 'include'
		}).then((res) => res.json());
		return {
			clarification: response.data
		};
	} catch (error) {
		redirect(302, '/');
	}
};
