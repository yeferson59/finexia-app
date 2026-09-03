import * as portfolio from '$lib/api/portfolio';
import * as transactions from '$lib/api/transactions';
import type { PageServerLoad, Actions } from './$types';
import type { Holding, Transaction } from '$lib/api/types';
import {
	transactionCreateSchema,
	transactionDeleteSchema,
	transactionUpdateSchema
} from '$lib/features/portfolio';

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
			baseCurrency: 'USD',
			txnMeta: DEFAULT_META
		};
	}

	const allHoldings: Holding[] = response.data.holdings ?? [];
	const entries = allHoldings.filter((h) => h.ticker === params.symbol);
	const baseCurrency = response.data.baseCurrency?.trim() || 'USD';

	// El denominador de la asignación se suma en moneda base; hacerlo con los
	// precios nativos mezclaba EUR con USD y daba porcentajes imposibles.
	const portfolioTotalValue = allHoldings.reduce((sum, h) => {
		if (h.marketValueBase !== undefined && h.marketValueBase !== '') {
			return sum + (parseFloat(h.marketValueBase) || 0);
		}
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

	return { entries, transactions: transactionsList, portfolioTotalValue, baseCurrency, txnMeta };
};

export const actions: Actions = {
	createTransaction: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const { success, error, data } = await transactionCreateSchema.safeParseAsync({
			entryId: formData.get('entryId'),
			type: formData.get('type'),
			quantity: formData.get('quantity'),
			price: formData.get('price'),
			currency: formData.get('currency') ?? 'USD',
			// Vacía significa tasa 1, que es lo correcto mientras la operación y
			// la posición estén en la misma moneda. Si difieren, el backend
			// rechaza la transacción en vez de inventar la tasa de hoy.
			fxRate: formData.get('fxRate') ?? 1,
			fees: formData.get('fees') ?? 0,
			feesCurrency: formData.get('feesCurrency') ?? '',
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
			fxRate: data.fxRate,
			fees: data.fees,
			// Se omite en vez de mandarse vacía: el backend decodifica la moneda
			// a un tipo que rechaza la cadena vacía con un error de formato, así
			// que «no la sé» tiene que ser la ausencia del campo, no un hueco.
			// Ausente significa la moneda de la operación.
			...(data.feesCurrency ? { feesCurrency: data.feesCurrency } : {}),
			transactionDate: data.transactionDate,
			notes: data.notes ?? ''
		});

		if (!response.ok) {
			return { success: false, error: response.details || response.message || response.action };
		}

		return { success: response.success };
	},

	editTransaction: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const { success, error, data } = await transactionUpdateSchema.safeParseAsync({
			txnId: formData.get('txnId'),
			type: formData.get('type'),
			quantity: formData.get('quantity'),
			price: formData.get('price'),
			currency: formData.get('currency') ?? 'USD',
			// Vacía significa tasa 1, que es lo correcto mientras la operación y
			// la posición estén en la misma moneda. Si difieren, el backend
			// rechaza la transacción en vez de inventar la tasa de hoy.
			fxRate: formData.get('fxRate') ?? 1,
			fees: formData.get('fees') ?? 0,
			feesCurrency: formData.get('feesCurrency') ?? '',
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
			fxRate: data.fxRate,
			fees: data.fees,
			// Se omite en vez de mandarse vacía: el backend decodifica la moneda
			// a un tipo que rechaza la cadena vacía con un error de formato, así
			// que «no la sé» tiene que ser la ausencia del campo, no un hueco.
			// Ausente significa la moneda de la operación.
			...(data.feesCurrency ? { feesCurrency: data.feesCurrency } : {}),
			transactionDate: data.transactionDate,
			notes: data.notes ?? ''
		});

		if (!response.ok) {
			// `details` primero: desde que FromDomain devuelve el texto del error en
			// los 4xx, es el campo que nombra el problema —«USD no se convierte en
			// sí misma a 1.0638»—, mientras que `message` es el titular genérico
			// del handler.
			return {
				success: false,
				edited: true,
				error: response.details || response.message || response.action
			};
		}

		return { success: response.success, edited: true };
	},

	// Borrar una transacción deja la posición recalculada por la base (en 0 si
	// era la última), así que basta con recargar la página al volver.
	deleteTransaction: async ({ request, fetch, cookies }) => {
		const formData = await request.formData();

		const { success, error, data } = await transactionDeleteSchema.safeParseAsync({
			txnId: formData.get('txnId')
		});

		if (!success) {
			return { success: false, deleted: true, error: error.message };
		}

		const response = await transactions.deleteTransaction({ cookies, fetch }, data.txnId);

		if (!response.ok) {
			return { success: false, deleted: true, error: response.message ?? response.action };
		}

		return { success: response.success, deleted: true };
	}
};
