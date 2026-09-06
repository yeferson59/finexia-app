/**
 * Constantes y cálculos del dominio `transactions`. Sin dependencias de Svelte
 * ni de red; los contratos HTTP viven en `types.ts`.
 */

import type { UserTransaction } from '$lib/api/types';

/** Tipos de operación admitidos por el backend, con su etiqueta legible. */
export const TXN_TYPE_OPTIONS: { value: string; label: string }[] = [
	{ value: 'buy', label: 'Compra' },
	{ value: 'sell', label: 'Venta' },
	{ value: 'dividend', label: 'Dividendo' },
	{ value: 'interest', label: 'Interés' },
	{ value: 'transfer_in', label: 'Transferencia entrada' },
	{ value: 'transfer_out', label: 'Transferencia salida' },
	{ value: 'fee', label: 'Cargo' },
	{ value: 'split', label: 'División' }
];

/** Categorías de activo que puede fijar el import como valor por defecto. */
export const CATEGORY_OPTIONS: { value: string; label: string }[] = [
	{ value: 'stock', label: 'Acciones' },
	{ value: 'etf', label: 'ETFs' },
	{ value: 'crypto', label: 'Criptomonedas' },
	{ value: 'bond', label: 'Bonos' },
	{ value: 'cash', label: 'Efectivo' },
	{ value: 'real_estate', label: 'Bienes raíces' },
	{ value: 'commodity', label: 'Materias primas' },
	{ value: 'other', label: 'Otros' }
];

/** Índice `value → label` de `TXN_TYPE_OPTIONS`, para pintar filas del preview. */
export const TXN_TYPE_LABELS: Record<string, string> = Object.fromEntries(
	TXN_TYPE_OPTIONS.map((t) => [t.value, t.label])
);

/** Importe de una transacción, en la moneda que se liquidó. */
export interface TransactionTotal {
	/** Cantidad × precio, llevado a `currency` por la tasa del día. */
	amount: number;
	/** Moneda en la que la cuenta pagó o cobró. */
	currency: string;
	/** Moneda en la que cotizó la operación, si no es la misma. */
	quoteCurrency: string;
	/** La operación se liquidó en una moneda distinta de la que cotizó. */
	converted: boolean;
	/** Lo que valía una unidad de `quoteCurrency` en `currency` ese día. */
	rate: number;
}

/**
 * Lo que costó o produjo una transacción.
 *
 * El listado multiplicaba cantidad por precio y etiquetaba el resultado con la
 * moneda de la operación, sin mirar la tasa. Una compra de LVMH cotizada en
 * euros y liquidada en dólares se leía como si el bróker hubiera cobrado
 * 606,60 USD. Es el mismo cálculo que hace la tabla de movimientos de un
 * activo, y tiene que dar lo mismo en las dos pantallas.
 */
export function transactionTotal(txn: UserTransaction): TransactionTotal {
	const quantity = parseFloat(txn.quantity) || 0;
	const price = parseFloat(txn.price) || 0;
	const rate = parseFloat(txn.fxRate ?? '1') || 1;
	const currency = txn.costCurrency || txn.currency;

	return {
		amount: quantity * price * rate,
		currency,
		quoteCurrency: txn.currency,
		converted: currency !== txn.currency,
		rate
	};
}

/**
 * Las transacciones de la más reciente a la más antigua.
 *
 * El orden lo trae el backend, pero la página lo afirma en su subtítulo y
 * pagina sobre él: si algún día llega otro, la primera hoja enseñaría
 * cualquier cosa. Copia antes de ordenar; `sort` muta.
 */
export function sortByDateDesc(transactions: UserTransaction[]): UserTransaction[] {
	return [...transactions].sort((a, b) => b.transactionDate.localeCompare(a.transactionDate));
}
