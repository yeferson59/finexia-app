import type { LayoutServerLoad } from './$types';
import { authedFetch } from '$lib/server/api';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const response = await authedFetch({ cookies, fetch }, '/portfolios/risks');

	if (!response.ok) {
		return { risks: [], success: false };
	}

	const { data, success } = await response.json();

	if (!success) {
		return { risks: [], success: false };
	}

	return { risks: data, success: true };
};
