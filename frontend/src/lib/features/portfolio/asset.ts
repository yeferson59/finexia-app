/**
 * Helpers puros, constantes y tipos del detalle de un activo dentro de un
 * portafolio (`routes/dashboard/portfolios/[id]/assets/[symbol]`). Sin
 * dependencias de Svelte ni de red; los contratos vienen de `$lib/api/types`.
 */
import type { Holding, Transaction } from '$lib/api/types';
import { dayGap } from '$lib/shared/finance/returns';
import { formatCalendarDate } from '$lib/shared/format/date';
import { formatCurrency } from '$lib/shared/format/money';
import { formatSignedPercent } from '$lib/shared/format/percent';

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
	cash_interest: 'type-interest',
	cash_dividend: 'type-dividend',
	cash_sale: 'type-sell',
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
	cash_interest: 'Interés',
	cash_dividend: 'Dividendo',
	cash_sale: 'Venta',
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
 * Cómo se cuenta cada clase de activo, en singular y plural.
 *
 * La ficha decía «42 AAPL», que es el ticker haciendo de unidad de medida. Lo
 * que se tiene son 42 *acciones*, y de un ETF, participaciones. Es la palabra
 * que usa quien lo compró, no la que usa la base de datos.
 */
const UNIT_NOUNS: Record<string, [string, string]> = {
	stock: ['acción', 'acciones'],
	etf: ['participación', 'participaciones'],
	bond: ['título', 'títulos']
};

/**
 * El nombre de la unidad de una clase de activo, en singular o en plural.
 *
 * Cripto, efectivo y las clases que el backend añada se quedan en «unidad»:
 * inventarles un nombre propio arriesga más de lo que aclara.
 */
/**
 * Un sitio de donde puede salir el dinero de una compra: la cuenta principal de
 * una plataforma o uno de sus bolsillos, con lo que hay en él.
 *
 * `id` vacío es la cuenta principal, que es la convención del backend —ahí no
 * hay bolsillo que nombrar— y lo que manda el formulario cuando no se elige
 * otra cosa. Los depósitos a plazo nunca entran en esta lista: están cerrados
 * hasta que vencen.
 */
export interface CashSource {
	id: string;
	name: string;
	balance: number;
}

/**
 * Un importe en la moneda de la cuenta, tal como lo escriben los formularios
 * que liquidan contra el efectivo: el total convertido, lo que se abona y lo
 * que se paga.
 *
 * Fija `es-CO` a propósito, que es el locale del panel; la copia que había en
 * cada uno de esos formularios hacía lo mismo por separado.
 */
export function formatSettled(value: number, currency: string): string {
	return new Intl.NumberFormat('es-CO', {
		style: 'currency',
		currency,
		minimumFractionDigits: 2
	}).format(value);
}

export function unitNoun(assetType: string, count = 2): string {
	const [one, many] = UNIT_NOUNS[assetType] ?? ['unidad', 'unidades'];
	return count === 1 ? one : many;
}

/** La cantidad con el nombre de su unidad: «42 acciones», «0,15 unidades». */
export function formatUnits(quantity: number, assetType: string): string {
	const amount = quantity.toLocaleString('es-CO', { maximumFractionDigits: 8 });
	return `${amount} ${unitNoun(assetType, quantity)}`;
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

/**
 * Clases de activo que solo se negocian con la bolsa abierta. Cripto cotiza
 * todos los días, y del resto —efectivo, inmuebles, materias primas, «otro»— no
 * se puede afirmar nada, así que no se avisa.
 */
const EXCHANGE_TRADED = new Set(['stock', 'etf', 'bond']);

/** Días hacia atrás en los que una fecha todavía se lee como «la de hoy». */
const RECENT_DAYS = 7;

/** Distancia al precio de mercado a partir de la cual el precio no parece de esa fecha. */
const PRICE_GAP = 0.2;

/** Lo que hace falta para juzgar si la fecha de una operación es la suya. */
export interface TradeDateCheck {
	/** Fecha de la operación, `YYYY-MM-DD`. */
	date: string;
	/** Hoy en el calendario de quien registra, `YYYY-MM-DD`. */
	today: string;
	ticker: string;
	assetType: string;
	/** Precio unitario escrito, en `currency`. */
	price: number;
	currency: string;
	/** Precio de mercado que conoce la app; `null` si no conoce ninguno. */
	marketPrice: number | null;
	marketCurrency: string;
}

/**
 * Por qué la fecha de una compra o venta parece no ser la suya.
 *
 * Los formularios proponen la fecha de hoy, y quien carga una cartera que ya
 * tenía copia el precio de una confirmación vieja y se olvida de cambiarla. Nada
 * lo delataba: ADBE entró comprada «ayer» a 483,57 cuando valía 252,23, y TSM,
 * SPCX y VST con fecha de un sábado. La serie de crecimiento trata una operación
 * de esa fecha como dinero nuevo, y la ganancia que la posición ya traía se
 * contaba como rentabilidad del día en que se registró.
 *
 * Son avisos y no errores —una acción que se movió un 25 % esta semana existe—,
 * así que el formulario los enseña y deja guardar. Dos señales, cada una con su
 * frase:
 * - la fecha cae en fin de semana y el activo solo se negocia con la bolsa abierta;
 * - la fecha es de los últimos siete días y el precio está a más de un 20 % del
 *   de mercado, comparados en la misma moneda.
 */
export function tradeDateWarnings(check: TradeDateCheck): string[] {
	const warnings: string[] = [];
	if (!/^\d{4}-\d{2}-\d{2}$/.test(check.date)) return warnings;

	const [year, month, day] = check.date.split('-').map(Number);
	const weekday = new Date(Date.UTC(year, month - 1, day)).getUTCDay();
	if (EXCHANGE_TRADED.has(check.assetType) && (weekday === 0 || weekday === 6)) {
		const longDate = formatCalendarDate(check.date, {
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		});
		warnings.push(
			`El ${longDate} fue ${weekday === 6 ? 'sábado' : 'domingo'} y la bolsa no abrió. ` +
				'Si es una compra antigua, pon la fecha de la confirmación del bróker.'
		);
	}

	const age = dayGap(check.date, check.today);
	const currency = check.currency.trim().toUpperCase();
	const market = check.marketPrice ?? 0;
	const comparable =
		currency !== '' && currency === check.marketCurrency.trim().toUpperCase() && market > 0;

	if (age >= 0 && age <= RECENT_DAYS && comparable && check.price > 0) {
		const gap = check.price / market - 1;
		if (Math.abs(gap) > PRICE_GAP) {
			warnings.push(
				`${check.ticker} vale hoy ${formatCurrency(market, currency)} y escribiste ` +
					`${formatCurrency(check.price, currency)} (${formatSignedPercent(gap * 100, 0)}). ` +
					'Con una fecha tan reciente suele ser una compra antigua con la fecha de hoy: revisa la fecha.'
			);
		}
	}

	return warnings;
}

/**
 * Las transacciones que necesitan tasa para liquidar en `costCurrency`: las que
 * cotizan en otra moneda. Un split no mueve dinero y no la pide.
 */
export function transactionsNeedingRate<T extends { currency: string; type: string }>(
	transactions: T[],
	costCurrency: string
): T[] {
	const target = costCurrency.trim().toUpperCase();

	return transactions.filter(
		(txn) => txn.type !== 'split' && txn.currency.trim().toUpperCase() !== target
	);
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

/**
 * La fila de una transacción tal como la dejará el servidor, sacada de lo que
 * manda su formulario: es lo que la tabla pinta mientras el envío no se
 * confirma. El backend no recalcula nada de la fila —solo de la posición—, así
 * que coincide con la que llega después.
 *
 * `base` es la transacción que se edita, o lo que ya se sabe de una nueva (su
 * id provisional y la moneda de la cuenta). Un campo que el formulario no manda
 * se queda como estaba; las casillas, que desmarcadas no viajan, se leen de su
 * presencia.
 */
export function draftTransaction(
	formData: FormData,
	base: Pick<Transaction, 'id'> & Partial<Transaction>
): Transaction {
	const field = (name: string, fallback: string) => {
		const value = formData.get(name);
		return typeof value === 'string' && value.trim() !== '' ? value.trim() : fallback;
	};

	const currency = field('currency', base.currency ?? 'USD');
	const cashPaid = formData.has('payFromCash');

	return {
		id: base.id,
		entryId: field('entryId', base.entryId ?? ''),
		type: field('type', base.type ?? 'buy'),
		quantity: field('quantity', base.quantity ?? '0'),
		price: field('price', base.price ?? '0'),
		currency,
		fxRate: field('fxRate', base.fxRate ?? '1'),
		costCurrency: base.costCurrency,
		fees: field('fees', '0'),
		feesCurrency: field('feesCurrency', currency).toUpperCase(),
		transactionDate: field('transactionDate', base.transactionDate ?? ''),
		notes: field('notes', ''),
		cashCredited: formData.has('creditCash'),
		cashPaid,
		cashPocketId: cashPaid ? field('payFromPocketId', '') || null : null,
		createdAt: base.createdAt ?? new Date().toISOString()
	};
}
