import { env } from '$env/dynamic/private';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const accessToken = cookies.get('access_token_finexia');

	const response = await fetch(`${env.BASE_API}/portfolios/risks`, {
		headers: {
			Authorization: `Bearer ${accessToken}`
		}
	});

	if (!response.ok) {
		return { risks: [], success: false };
	}

	const { data, success } = await response.json();

	if (!success) {
		return { risks: [], success: false };
	}

	return { risks: data, success: true };
};
