import type { LayoutServerLoad } from './$types';
import * as portfolio from '$lib/api/portfolio';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const res = await portfolio.getSummaries({ cookies, fetch });

	if (!res.success) {
		return { portfolios: [], success: false };
	}

	return { portfolios: res.data ?? [], success: true };
};
