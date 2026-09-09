import { goto, invalidate } from '$app/navigation';
import Swal from 'sweetalert2';
import { useUserApp, useUserProfile } from '@/apps/contest/application/user';
import { refreshContestState } from '$lib/refreshContestState';

/**
 * The dependency the root layout's load registers for the signed-in profile.
 * Roles are loaded there and nowhere else, and client-side navigation reuses
 * the load's result — so an edit that changes the current user's own roles
 * has to `invalidate` this key to be seen without a reload.
 */
export const PROFILE_DEPENDENCY = 'app:profile';

/**
 * Whether an admin-side edit targets the account this browser is signed in
 * as. Username is the join key: the profile carries no id, and a username
 * cannot change (UpdateUserRequest has no such field), so the comparison is
 * stable across the edit itself.
 */
export function isSelf(username: string): boolean {
	return useUserProfile().username === username;
}

/**
 * After an admin edit that may have changed the signed-in user's own roles.
 * Re-runs the profile load through SvelteKit — which also re-runs the
 * dashboard gate, and SvelteKit follows its redirect without a reload — then
 * refetches the contest-side state, because that is fetched against the role
 * at fetch time: only a contestant is served their own contests and the
 * `isParticipant` flag the navbar's submit button is gated on.
 *
 * Returns whether the edit was a self-edit, so the caller can decide where to
 * navigate next (a demoted admin has already been sent home by the gate).
 */
export async function refreshProfileIfSelf(username: string): Promise<boolean> {
	if (!isSelf(username)) return false;
	await invalidate(PROFILE_DEPENDENCY);
	refreshContestState();
	return true;
}

/**
 * After the signed-in user deleted their own account. The backend already
 * refuses every request from this session (verified: profile answers 401),
 * but the JWT cookie is still in the browser, and the login page bounces
 * anyone carrying one back to the home page — without an explicit logout the
 * user could not sign in again until the cookie expired. `/auth/logout`
 * clears both cookies even while answering 401, which is why the failure is
 * swallowed here. The 401 also trips the global "session expired" toast;
 * the message below replaces it with what actually happened.
 */
export async function leaveAfterSelfDelete(): Promise<void> {
	await useUserApp()
		.logout()
		.catch(() => undefined);
	Swal.fire({
		icon: 'info',
		title: '已刪除你自己的帳號',
		text: '將返回登入頁',
		showConfirmButton: false,
		timer: 2000
	});
	await goto('/login');
}
