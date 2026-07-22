import type { LayoutServerLoad } from './$types';
import * as portfolio from '$lib/api/portfolio';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	const res = await portfolio.getRisks({ cookies, fetch });

	if (!res.ok || !res.success) {
		return { risks: [], success: false };
	}

	return { risks: res.data ?? [], success: true };
};
