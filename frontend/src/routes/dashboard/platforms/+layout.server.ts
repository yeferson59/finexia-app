import type { LayoutServerLoad } from './$types';
import { authedFetch } from '$lib/server/api';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const res = await authedFetch({ cookies, fetch }, '/portfolios/sources');

	const { data, success } = await res.json();

	if (!success) {
		return { platforms: [] };
	}

	return { platforms: data };
};
