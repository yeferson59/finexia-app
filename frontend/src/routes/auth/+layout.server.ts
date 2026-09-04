import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';
import { takeReturnTo } from '$lib/server/session';

export const load: LayoutServerLoad = async ({ locals, cookies }) => {
	if (locals.session && locals.user) {
		// Already signed in: honour a pending return target (the OAuth consent
		// screen sets one) before falling back to the dashboard.
		return redirect(303, takeReturnTo(cookies) ?? '/dashboard');
	}
};
