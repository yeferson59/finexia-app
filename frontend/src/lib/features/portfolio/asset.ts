/**
 * Helpers puros, constantes y tipos del detalle de un activo dentro de un
 * portafolio (`routes/dashboard/portfolios/[id]/assets/[symbol]`). Sin
 * dependencias de Svelte ni de red; los contratos vienen de `$lib/api/types`.
 */
import type { Holding } from '$lib/api/types';

export const TRANSACTION_TYPES = [
	{ value: 'buy', label: 'Compra' },
	{ value: 'sell', label: 'Venta' },
	{ value: 'dividend', label: 'Dividendo' },
	{ value: 'transfer_in', label: 'Transferencia entrada' },
	{ value: 'transfer_out', label: 'Transferencia salida' },
	{ value: 'fee', label: 'Comisión' },
	{ value: 'interest', label: 'Interés' },
	{ value: 'split', label: 'Split' }
] as const;

export const TYPE_STYLE: Record<string, string> = {
	buy: 'type-buy',
	sell: 'type-sell',
	dividend: 'type-dividend',
	fee: 'type-fee',
	interest: 'type-interest',
	transfer_in: 'type-transfer',
	transfer_out: 'type-transfer',
	split: 'type-split'
};

export const TYPE_LABEL: Record<string, string> = {
	buy: 'Compra',
	sell: 'Venta',
	dividend: 'Dividendo',
	fee: 'Comisión',
	interest: 'Interés',
	transfer_in: 'T. Entrada',
	transfer_out: 'T. Salida',
	split: 'Split'
};

/**
 * Cómo se pinta el formulario según el tipo de transacción:
 * - `trade`  → cantidad + precio unitario + comisión (buy, sell, transfer_*)
 * - `amount` → solo monto (dividendo, comisión, interés); cantidad = 1 implícita
 * - `split`  → nuevas acciones recibidas, sin precio
 */
export type TxnMode = 'trade' | 'amount' | 'split';

export function txnModeFor(type: string): TxnMode {
	if (type === 'split') return 'split';
	if (['fee', 'dividend', 'interest'].includes(type)) return 'amount';
	return 'trade';
}

/** Etiqueta del campo de precio/monto, que depende del tipo. */
export function priceLabelFor(type: string): string {
	if (txnModeFor(type) !== 'amount') return 'Precio unitario';
	if (type === 'dividend') return 'Monto del dividendo';
	if (type === 'interest') return 'Monto del interés';
	return 'Monto de la comisión';
}

/** Métricas agregadas de la posición en un activo. */
export interface AssetPosition {
	ticker: string;
	name: string;
	assetType: string;
	exchange: string;
	currency: string;
	costCurrency: string;
	marketPrice: number;
	totalQty: number;
	totalCost: number;
	averageCost: number;
	totalValue: number;
	gainLoss: number;
	gainLossPercent: number;
	allocation: number;
}

/**
 * Calcula la posición a partir de las entradas del activo. Usa los agregados
 * mantenidos por el backend, así que es exacta con independencia de la
 * paginación de transacciones.
 */
export function computePosition(
	entries: Holding[],
	portfolioTotalValue: number
): AssetPosition | null {
	if (entries.length === 0) return null;

	const first = entries[0];

	const totalQty = entries.reduce((s, e) => s + (parseFloat(e.quantity) || 0), 0);
	const rawCost = entries.reduce(
		(s, e) => s + (parseFloat(e.quantity) || 0) * (parseFloat(e.price) || 0),
		0
	);
	const averageCost = totalQty > 0 ? rawCost / totalQty : 0;
	const totalCost = rawCost;

	const marketPrice = parseFloat(first.marketPrice) || averageCost;
	const totalValue = totalQty * marketPrice;
	const gainLoss = totalValue - totalCost;
	const gainLossPercent = totalCost > 0 ? (gainLoss / totalCost) * 100 : 0;
	const allocation = portfolioTotalValue > 0 ? (totalValue / portfolioTotalValue) * 100 : 0;

	return {
		ticker: first.ticker,
		name: first.name,
		assetType: first.assetType,
		exchange: first.exchange,
		currency: first.currency,
		costCurrency: first.costCurrency,
		marketPrice,
		totalQty,
		totalCost,
		averageCost,
		totalValue,
		gainLoss,
		gainLossPercent,
		allocation
	};
}

/** Metadatos de paginación de las transacciones del activo. */
export interface TxnMeta {
	total: number;
	page: number;
	limit: number;
	totalPages: number;
}

/** Resultado de las form actions de la página de activo. */
export interface AssetActionResult {
	success?: boolean;
	edited?: boolean;
	error?: string;
}
