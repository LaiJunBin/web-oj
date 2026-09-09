import type { LayoutLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: LayoutLoad = async ({ parent }) => {
	const { user } = await parent();
	// Signed out → login. Signed in without the admin role → home. The second
	// case is also where an admin lands after removing their own admin role:
	// invalidating the profile re-runs this load, and SvelteKit follows the
	// redirect in place — no reload, no blank page, no loop (home is not gated).
	if (!user) {
		redirect(302, '/login');
	}
	if (!user.isAdmin) {
		redirect(302, '/');
	}
};
