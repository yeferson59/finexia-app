import type { PageServerLoad } from './$types';
import * as transactions from '$lib/api/transactions';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const res = await transactions.getRecent({ cookies, fetch });

	return {
		transactions: res.ok && res.success && Array.isArray(res.data) ? res.data : []
	};
};
