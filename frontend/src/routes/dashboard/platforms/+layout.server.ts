import { env } from '$env/dynamic/private';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const accessToken = cookies.get('access_token_finexia');

	const res = await fetch(`${env.BASE_API}/portfolios/sources`, {
		headers: {
			Authorization: `Bearer ${accessToken}`
		}
	});

	const { data, success } = await res.json();

	if (!success) {
		return { platforms: [] };
	}

	console.log(data[0].portfolioEntries);
	return { platforms: data };
};
