import { z } from 'zod';
import * as portfolio from '$lib/api/portfolio';
import * as transactions from '$lib/api/transactions';
import type { PageServerLoad, Actions } from './$types';
import type { Holding, Transaction } from '$lib/api/types';

export interface TxnMeta {
	total: number;
	page: number;
	limit: number;
	totalPages: number;
}

const DEFAULT_META: TxnMeta = { total: 0, page: 1, limit: 20, totalPages: 0 };

export const load: PageServerLoad = async ({ cookies, fetch, params, url }) => {
	const event = { cookies, fetch };

	const page = Math.max(1, parseInt(url.searchParams.get('page') ?? '1', 10) || 1);
	const limit = (() => {
		const raw = parseInt(url.searchParams.get('limit') ?? '20', 10) || 20;
		return raw >= 1 && raw <= 100 ? raw : 20;
	})();

	const [response, txnRes] = await Promise.all([
		portfolio.getPortfolio(event, params.id),
		transactions.getAssetTransactions(event, params.id, params.symbol, page, limit)
	]);

	if (!response.ok || !response.success || !response.data) {
		return {
			entries: [] as Holding[],
			transactions: [] as Transaction[],
			portfolioTotalValue: 0,
			txnMeta: DEFAULT_META
		};
	}

	const allHoldings: Holding[] = response.data.holdings ?? [];
	const entries = allHoldings.filter((h) => h.ticker === params.symbol);

	const portfolioTotalValue = allHoldings.reduce((sum, h) => {
		const qty = parseFloat(h.quantity) || 0;
		const mp = parseFloat(h.marketPrice) || parseFloat(h.price) || 0;
		return sum + qty * mp;
	}, 0);

	const paged = txnRes.success ? txnRes.data : null;

	const transactionsList: Transaction[] = paged?.data ?? [];
	const txnMeta: TxnMeta = paged
		? {
				total: paged.total ?? 0,
				page: paged.page ?? page,
				limit: paged.limit ?? limit,
				totalPages: paged.totalPages ?? 0
			}
		: DEFAULT_META;

	return { entries, transactions: transactionsList, portfolioTotalValue, txnMeta };
};

const txnSchema = z.object({
	type: z.string().min(1),
	quantity: z.coerce.number().positive(),
	price: z.coerce.number().min(0),
	currency: z.string().default('USD'),
	fees: z.coerce.number().min(0).default(0),
	transactionDate: z.coerce.date(),
	notes: z.string().optional()
});

export const actions: Actions = {
	createTransaction: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const { success, error, data } = await txnSchema.extend({ entryId: z.uuid() }).safeParseAsync({
			entryId: formData.get('entryId'),
			type: formData.get('type'),
			quantity: formData.get('quantity'),
			price: formData.get('price'),
			currency: formData.get('currency') ?? 'USD',
			fees: formData.get('fees') ?? 0,
			transactionDate: formData.get('transactionDate'),
			notes: formData.get('notes')
		});

		if (!success) {
			return { success: false, error: error.message };
		}

		const response = await transactions.createTransaction({ cookies, fetch }, data.entryId, {
			type: data.type,
			quantity: data.quantity,
			price: data.price,
			currency: data.currency,
			fees: data.fees,
			transactionDate: data.transactionDate,
			notes: data.notes ?? ''
		});

		if (!response.ok) {
			return { success: false };
		}

		return { success: response.success };
	},

	editTransaction: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const { success, error, data } = await txnSchema.extend({ txnId: z.uuid() }).safeParseAsync({
			txnId: formData.get('txnId'),
			type: formData.get('type'),
			quantity: formData.get('quantity'),
			price: formData.get('price'),
			currency: formData.get('currency') ?? 'USD',
			fees: formData.get('fees') ?? 0,
			transactionDate: formData.get('transactionDate'),
			notes: formData.get('notes')
		});

		if (!success) {
			return { success: false, edited: true, error: error.message };
		}

		const response = await transactions.updateTransaction({ cookies, fetch }, data.txnId, {
			type: data.type,
			quantity: data.quantity,
			price: data.price,
			currency: data.currency,
			fees: data.fees,
			transactionDate: data.transactionDate,
			notes: data.notes ?? ''
		});

		if (!response.ok) {
			return { success: false, edited: true, error: response.message ?? response.action };
		}

		return { success: response.success, edited: true };
	}
};
