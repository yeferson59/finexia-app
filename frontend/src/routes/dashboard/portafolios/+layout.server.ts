import { env } from '$env/dynamic/private';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const accessToken = cookies.get('access_token_finexia');

	const response = await fetch(`${env.BASE_API}/portfolios/id`, {
		headers: {
			Authorization: `Bearer ${accessToken}`
		}
	});

	const { data, success } = await response.json();

	if (!success) {
		return { portfolios: [], success: false };
	}

	return { portfolios: data, success: true };
};
