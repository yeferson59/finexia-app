import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: LayoutServerLoad = async ({ locals }) => {
	if (!locals.session || !locals.user) {
		return redirect(303, '/auth');
	}

	return {
		user: locals.user,
		session: locals.session
	};
};
