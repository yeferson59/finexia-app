/**
 * Helpers puros del efectivo: los saldos que guarda cada plataforma y los
 * movimientos que los cambian. Sin dependencias de Svelte ni de red; los
 * contratos vienen de `$lib/api/types`.
 */

import type { CashBalance, CashMovement } from '$lib/api/types';

export type { CashBalance, CashMovement };

/** Lo que se puede registrar desde esta pantalla. */
export type CashKind = 'deposit' | 'withdrawal' | 'interest';

/**
 * Los tres movimientos, con la frase que dice qué hacen con la rentabilidad.
 *
 * Es la duda que importa al elegir: un depósito sube el saldo igual que unos
 * intereses, y solo uno de los dos es ganancia. Decirlo al lado de la opción
 * evita que alguien anote la nómina como intereses y vea un rendimiento que
 * nunca tuvo.
 */
export const CASH_KIND_OPTIONS: { value: CashKind; label: string; hint: string }[] = [
	{
		value: 'deposit',
		label: 'Depósito',
		hint: 'Dinero que metes en la cuenta. Sube el saldo y cuenta como aporte, no como ganancia.'
	},
	{
		value: 'withdrawal',
		label: 'Retiro',
		hint: 'Dinero que sacas de la cuenta. Baja el saldo y cuenta como retiro, no como pérdida.'
	},
	{
		value: 'interest',
		label: 'Intereses',
		hint: 'Lo que te abonó la cuenta. Sube el saldo y cuenta como rendimiento. Anótalos netos, ya descontada la retención.'
	}
];

const KIND_LABELS: Record<string, string> = {
	deposit: 'Depósito',
	withdrawal: 'Retiro',
	interest: 'Intereses',
	other: 'Otro movimiento'
};

/** Nombre de un movimiento; el crudo si el backend manda uno que no se conoce. */
export function formatCashKind(kind: string): string {
	return KIND_LABELS[kind] ?? kind;
}

/** Si el movimiento suma al saldo (1), lo resta (−1) o no lo toca (0). */
export function cashKindSign(kind: string): 1 | -1 | 0 {
	if (kind === 'deposit' || kind === 'interest') return 1;
	if (kind === 'withdrawal') return -1;
	return 0;
}

/** Plataforma y portafolio de un saldo o un movimiento, en una línea. */
export function cashAccountLabel(row: { sourceName: string; portfolioName: string }): string {
	return row.sourceName ? `${row.sourceName} · ${row.portfolioName}` : row.portfolioName;
}

/** Lo que suma el efectivo en una moneda, entre todas sus cuentas. */
export interface CashCurrencyTotal {
	currency: string;
	/** En la propia moneda: esto sí se suma sin tasas. */
	balance: number;
	/** En la moneda de la pantalla, solo lo que se pudo convertir. */
	value: number;
	accounts: number;
}

export interface CashSummary {
	/** Lo que hay en total, en `currency`. */
	total: number;
	currency: string;
	/**
	 * Saldos con dinero que no tenían tasa a `currency` y quedaron fuera de
	 * `total`: sumarlos a valor nominal mezclaría monedas.
	 */
	unconverted: number;
	/** Cuentas con saldo, de todas las listadas. */
	funded: number;
	/** Por moneda, de la que más pesa a la que menos. */
	byCurrency: CashCurrencyTotal[];
}

/**
 * El total del efectivo y su reparto por moneda.
 *
 * El backend convierte cada saldo por separado y marca los que no pudo. Esos se
 * quedan fuera del total —es la regla del listado de portafolios— pero siguen
 * en el reparto por moneda con su importe propio, que es donde sí significan
 * algo.
 */
export function summarizeCash(balances: CashBalance[], fallbackCurrency: string): CashSummary {
	let total = 0;
	let unconverted = 0;
	let funded = 0;
	const groups = new Map<string, CashCurrencyTotal>();

	for (const row of balances) {
		const balance = parseFloat(row.balance) || 0;
		const value = parseFloat(row.value) || 0;

		if (balance !== 0) funded += 1;

		if (row.fxConverted) total += value;
		else if (balance !== 0) unconverted += 1;

		const group = groups.get(row.currency) ?? {
			currency: row.currency,
			balance: 0,
			value: 0,
			accounts: 0
		};
		group.balance += balance;
		group.value += row.fxConverted ? value : 0;
		group.accounts += 1;
		groups.set(row.currency, group);
	}

	const byCurrency = [...groups.values()].sort(
		(a, b) => b.value - a.value || b.balance - a.balance || a.currency.localeCompare(b.currency)
	);

	return {
		total,
		currency: balances[0]?.displayCurrency || fallbackCurrency,
		unconverted,
		funded,
		byCurrency
	};
}

/**
 * El motivo de un movimiento rechazado, en la frase que ve el usuario.
 *
 * El backend lo explica en `details`, pero en inglés y pensando en quien
 * integra la API. Aquí se traducen los casos que puede provocar el formulario;
 * lo demás es un fallo que el usuario no puede arreglar cambiando un campo.
 */
export function cashErrorMessage(status: number, details = ''): string {
	if (details.includes('portfolio or source not found')) {
		return 'No encontramos ese portafolio o esa plataforma. Recarga la página y vuelve a elegirlos.';
	}
	if (status === 409) {
		return 'El saldo no alcanza: el movimiento lo dejaría en negativo. Revisa el importe o registra antes el depósito que lo cubre.';
	}
	if (status === 404) {
		return 'Ese movimiento ya no existe. Recarga la página.';
	}
	if (details.includes('record movements on it from the position')) {
		return 'Esa plataforma ya guarda esta moneda como una posición comprada con otra moneda. Registra el movimiento desde esa posición.';
	}
	if (details.includes('cannot be edited as a cash movement')) {
		return 'Este movimiento no se anotó como efectivo, así que se edita desde la posición del activo.';
	}
	if (details.includes('net of fees')) {
		return 'Los intereses se anotan netos, sin comisión.';
	}
	if (details.includes('cannot exceed its amount')) {
		return 'La comisión no puede ser mayor que el retiro.';
	}
	if (details.includes('currency must be one of')) {
		return 'Esa moneda todavía no se puede convertir. Elige una de la lista.';
	}

	return 'No pudimos guardar el movimiento. Vuelve a intentarlo en un momento.';
}
