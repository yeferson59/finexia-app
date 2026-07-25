/**
 * Constantes del dominio `transactions`, compartidas por los pasos del wizard
 * de import. Sin dependencias de Svelte ni de red; los contratos HTTP viven en
 * `types.ts`.
 */

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
