import { env } from '$env/dynamic/private';
import type { PageServerLoad } from './$types';

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
	const accessToken = cookies.get('access_token_finexia');

	const response = await fetch(`${env.BASE_API}/portfolios/transactions`, {
		headers: { Authorization: `Bearer ${accessToken}` }
	}).catch(() => null);

	if (!response?.ok) {
		return { transactions: [] as UserTransaction[] };
	}

	const { data, success } = await response.json();

	return {
		transactions: (success && Array.isArray(data) ? data : []) as UserTransaction[]
	};
};
