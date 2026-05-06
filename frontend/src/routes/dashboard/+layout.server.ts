import { env } from '$/config/env';
import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const accessToken = cookies.get('access_token_finexia');

	if (!accessToken) {
		return redirect(303, '/auth');
	}

	const response = await fetch(`${env.baseApi}/auth/session`, {
		headers: {
			Authorization: `Bearer ${accessToken}`
		}
	});

	if (!response.ok) {
		return redirect(303, '/auth');
	}

	const { data, success } = await response.json();

	if (!success) {
		return redirect(303, '/auth');
	}

	return {
		id: data.id,
		name: data.name,
		email: data.email
	};
};
