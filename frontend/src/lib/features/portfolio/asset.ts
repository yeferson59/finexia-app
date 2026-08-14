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

/**
 * Métricas agregadas de la posición en un activo.
 *
 * Conviven tres monedas y conviene no confundirlas: `averageCost` está en
 * `costCurrency` (la de la compra), `marketPrice` en `currency` (la de
 * cotización del activo) y los totales (`totalCost`, `totalValue`, `gainLoss`)
 * en `baseCurrency`, que es la única en la que se pueden sumar y restar.
 */
export interface AssetPosition {
	ticker: string;
	name: string;
	assetType: string;
	exchange: string;
	currency: string;
	costCurrency: string;
	baseCurrency: string;
	marketPrice: number;
	totalQty: number;
	totalCost: number;
	averageCost: number;
	totalValue: number;
	gainLoss: number;
	gainLossPercent: number;
	allocation: number;
	/** `false` si faltó la tasa: los totales están sin convertir. */
	fxConverted: boolean;
}

/**
 * Calcula la posición a partir de las entradas del activo. Usa los agregados
 * mantenidos por el backend, así que es exacta con independencia de la
 * paginación de transacciones.
 *
 * `portfolioTotalValue` tiene que venir en la misma moneda base que los
 * importes de las entradas; si no, la asignación compara peras con manzanas.
 */
export function computePosition(
	entries: Holding[],
	portfolioTotalValue: number,
	baseCurrency = 'USD'
): AssetPosition | null {
	if (entries.length === 0) return null;

	const first = entries[0];

	const totalQty = entries.reduce((s, e) => s + (parseFloat(e.quantity) || 0), 0);
	// El precio promedio es por unidad y se queda en la moneda de coste: es lo
	// que se compara contra el precio de mercado que muestra la ficha.
	const rawCost = entries.reduce(
		(s, e) => s + (parseFloat(e.quantity) || 0) * (parseFloat(e.price) || 0),
		0
	);
	const averageCost = totalQty > 0 ? rawCost / totalQty : 0;

	const marketPrice = parseFloat(first.marketPrice) || averageCost;

	// Los totales, en cambio, se toman ya convertidos a la moneda base: restar
	// un valor de mercado en EUR de un coste en USD daba un ROI inventado.
	const totalCost = sumBase(entries, 'costBasisBase', rawCost);
	const totalValue = sumBase(entries, 'marketValueBase', totalQty * marketPrice);
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
		baseCurrency,
		marketPrice,
		totalQty,
		totalCost,
		averageCost,
		totalValue,
		gainLoss,
		gainLossPercent,
		allocation,
		// Ausente (backend anterior) no es lo mismo que `false`: solo se avisa
		// cuando el backend afirma que faltó la tasa.
		fxConverted: entries.every((e) => e.fxConverted ?? true)
	};
}

/**
 * Suma un importe en moneda base de todas las entradas, volviendo al cálculo
 * nativo si el backend no envía el campo (versión anterior).
 */
function sumBase(
	entries: Holding[],
	field: 'costBasisBase' | 'marketValueBase',
	nativeFallback: number
): number {
	let total = 0;
	for (const entry of entries) {
		const raw = entry[field];
		if (raw === undefined || raw === '') return nativeFallback;
		const parsed = parseFloat(raw);
		if (!Number.isFinite(parsed)) return nativeFallback;
		total += parsed;
	}

	return total;
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
	deleted?: boolean;
	error?: string;
}
