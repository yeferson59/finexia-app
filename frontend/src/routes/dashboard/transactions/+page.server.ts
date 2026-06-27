import type { PageServerLoad } from './$types';
import { authedFetchSafe } from '$lib/server/api';

export interface UserTransaction {
	id: string;
	entryId: string;
	type: string;
	quantity: string;
	price: string;
	currency: string;
	fees: string;
	transactionDate: string;
	notes: string;
	createdAt: string;
	assetTicker: string;
	assetName: string;
}

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const response = await authedFetchSafe({ cookies, fetch }, '/portfolios/transactions');

	if (!response?.ok) {
		return { transactions: [] as UserTransaction[] };
	}

	const { data, success } = await response.json();

	return {
		transactions: (success && Array.isArray(data) ? data : []) as UserTransaction[]
	};
};
