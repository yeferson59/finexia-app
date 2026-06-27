import { env } from '$env/dynamic/private';
import { z } from 'zod';
import { authedFetch, authedFetchSafe } from '$lib/server/api';
import type { PageServerLoad, Actions } from './$types';

interface Entry {
	id: string;
	assetId: string;
	ticker: string;
	name: string;
	assetType: string;
	exchange: string;
	currency: string;
	quantity: string;
	price: string;
	marketPrice: string;
	costCurrency: string;
	category: string;
	entryDate: string;
	notes: string;
}

export interface Transaction {
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
}

export const load: PageServerLoad = async ({ cookies, fetch, params }) => {
	const event = { cookies, fetch };

	const response = await authedFetch(event, `/portfolios/${params.id}`);

	if (!response.ok) {
		return { entries: [] as Entry[], transactions: [] as Transaction[], portfolioTotalValue: 0 };
	}

	const { data, success } = await response.json();

	if (!success || !data) {
		return { entries: [] as Entry[], transactions: [] as Transaction[], portfolioTotalValue: 0 };
	}

	const allHoldings: Entry[] = data.holdings ?? [];
	const entries = allHoldings.filter((h) => h.ticker === params.symbol);

	const portfolioTotalValue = allHoldings.reduce((sum, h) => {
		const qty = parseFloat(h.quantity) || 0;
		const mp = parseFloat(h.marketPrice) || parseFloat(h.price) || 0;
		return sum + qty * mp;
	}, 0);

	// Fetch transactions for each entry in parallel
	const txnResponses = await Promise.all(
		entries.map((entry) =>
			authedFetchSafe(event, `/portfolios/entries/${entry.id}/transactions`).then((r) =>
				r ? r.json() : { success: false, data: [] }
			)
		)
	);

	const transactions: Transaction[] = txnResponses
		.filter((r) => r.success)
		.flatMap((r) => r.data ?? []);

	return { entries, transactions, portfolioTotalValue };
};

export const actions: Actions = {
	default: async ({ request, fetch, cookies }) => {
		const accessToken = cookies.get('access_token_finexia');
		const formData = await request.formData();

		const { success, error, data } = await z
			.object({
				entryId: z.uuid(),
				type: z.string().min(1),
				quantity: z.coerce.number().positive(),
				price: z.coerce.number().positive(),
				currency: z.string().default('USD'),
				fees: z.coerce.number().min(0).default(0),
				transactionDate: z.coerce.date(),
				notes: z.string().optional()
			})
			.safeParseAsync({
				entryId: formData.get('entryId'),
				type: formData.get('type'),
				quantity: formData.get('quantity'),
				price: formData.get('price'),
				currency: formData.get('currency') || 'USD',
				fees: formData.get('fees') || 0,
				transactionDate: formData.get('transactionDate'),
				notes: formData.get('notes')
			});

		if (!success) {
			return { success: false, error: error.message };
		}

		const response = await fetch(
			`${env.BASE_API}/portfolios/entries/${data.entryId}/transactions`,
			{
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${accessToken}`
				},
				body: JSON.stringify({
					type: data.type,
					quantity: data.quantity,
					price: data.price,
					currency: data.currency,
					fees: data.fees,
					transactionDate: data.transactionDate,
					notes: data.notes ?? ''
				})
			}
		);

		if (!response.ok) {
			return { success: false };
		}

		const json = await response.json();
		return { success: json.success ?? false };
	}
};
