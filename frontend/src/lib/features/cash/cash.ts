/**
 * Helpers puros del efectivo: los saldos que guarda cada plataforma y los
 * movimientos que los cambian. Sin dependencias de Svelte ni de red; los
 * contratos vienen de `$lib/api/types`.
 */

import type { CashBalance, CashMovement } from '$lib/api/types';
import { formatCalendarDate } from '$lib/shared/format/date';

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
	dividend: 'Dividendo',
	other: 'Otro movimiento'
};

/** Nombre de un movimiento; el crudo si el backend manda uno que no se conoce. */
export function formatCashKind(kind: string): string {
	return KIND_LABELS[kind] ?? kind;
}

/** Si el movimiento suma al saldo (1), lo resta (−1) o no lo toca (0). */
export function cashKindSign(kind: string): 1 | -1 | 0 {
	if (kind === 'deposit' || kind === 'interest' || kind === 'dividend') return 1;
	if (kind === 'withdrawal') return -1;
	return 0;
}

/**
 * Plataforma y portafolio de un saldo o un movimiento, en una línea.
 *
 * Sin `withPortfolio` solo se nombra la plataforma: con un único portafolio no
 * hay otro sitio donde el dinero pueda contar, y repetirlo en cada fila no dice
 * nada.
 */
export function cashAccountLabel(
	row: { sourceName: string; portfolioName: string },
	withPortfolio = true
): string {
	if (!withPortfolio) return row.sourceName || 'Sin plataforma';
	return row.sourceName ? `${row.sourceName} · ${row.portfolioName}` : row.portfolioName;
}

/**
 * Una cuenta: lo que guarda una plataforma en una moneda.
 *
 * Es la cifra del extracto del banco o del bróker. Por dentro cada portafolio
 * lleva su propio saldo en esa cuenta —así suma en su valor y en su
 * rentabilidad—, pero un único saldo real partido en varias filas sería
 * irreconocible.
 */
export interface CashAccount {
	/** Plataforma, moneda y bolsillo: lo que identifica la cuenta. */
	key: string;
	sourceId: string;
	sourceName: string;
	currency: string;
	/**
	 * El bolsillo, `null` en la cuenta principal. Un bolsillo es una subcuenta
	 * de la plataforma —su dinero suma en ella—, pero rinde su propia tasa, así
	 * que se agrupa y se mira aparte.
	 */
	pocketId: string | null;
	pocketName: string;
	pocketKind: string;
	displayCurrency: string;
	/** La suma de sus saldos, en `currency`. */
	balance: number;
	/** La suma en `displayCurrency` de los saldos que se pudieron convertir. */
	value: number;
	/** Si se pudieron convertir todos. */
	fxConverted: boolean;
	lastMovementDate: string | null;
	/** Un saldo por portafolio, del mayor al menor. */
	balances: CashBalance[];
	/**
	 * Los bolsillos de esta cuenta, en una cuenta principal; vacío en un
	 * bolsillo, que no tiene bolsillos propios.
	 */
	pockets: CashAccount[];
}

/** Los saldos agrupados por cuenta, de la que más vale a la que menos. */
export function groupCashAccounts(balances: CashBalance[]): CashAccount[] {
	const accounts = new Map<string, CashAccount>();

	for (const row of balances) {
		// La cuenta principal conserva la clave que siempre tuvo; un bolsillo
		// añade la suya. Así nada de lo que ya la usaba cambia.
		const key = row.pocketId
			? `${row.sourceId}:${row.currency}:${row.pocketId}`
			: `${row.sourceId}:${row.currency}`;
		const account = accounts.get(key) ?? {
			key,
			sourceId: row.sourceId,
			sourceName: row.sourceName,
			currency: row.currency,
			pocketId: row.pocketId,
			pocketName: row.pocketName,
			pocketKind: row.pocketKind,
			displayCurrency: row.displayCurrency,
			balance: 0,
			value: 0,
			fxConverted: true,
			lastMovementDate: null,
			balances: [],
			pockets: []
		};

		account.balance += parseFloat(row.balance) || 0;
		if (row.fxConverted) account.value += parseFloat(row.value) || 0;
		else account.fxConverted = false;

		if (
			row.lastMovementDate &&
			(!account.lastMovementDate ||
				Date.parse(row.lastMovementDate) > Date.parse(account.lastMovementDate))
		) {
			account.lastMovementDate = row.lastMovementDate;
		}

		account.balances.push(row);
		accounts.set(key, account);
	}

	const list = [...accounts.values()];
	for (const account of list) {
		account.balances.sort((a, b) => (parseFloat(b.balance) || 0) - (parseFloat(a.balance) || 0));
	}

	return list.sort(
		(a, b) =>
			b.value - a.value ||
			b.balance - a.balance ||
			a.sourceName.localeCompare(b.sourceName) ||
			a.currency.localeCompare(b.currency) ||
			a.pocketName.localeCompare(b.pocketName)
	);
}

/** Los movimientos de un mes. */
export interface CashMonth {
	/** `2026-09`. */
	key: string;
	/** «Septiembre de 2026». */
	label: string;
	movements: CashMovement[];
}

/**
 * Los movimientos agrupados por mes, en el orden en que aparece cada mes.
 *
 * La lista llega de la más reciente a la más antigua, así que los meses salen
 * en ese orden. El mes es el del día del calendario que guarda el backend, no
 * el de la hora local: un movimiento del 1 de septiembre no puede caer en
 * agosto en una zona al oeste de Greenwich.
 */
export function groupCashMovementsByMonth(movements: CashMovement[]): CashMonth[] {
	const months = new Map<string, CashMonth>();

	for (const movement of movements) {
		const key = movement.date.slice(0, 7);
		const month = months.get(key);

		if (month) {
			month.movements.push(movement);
			continue;
		}

		const label = formatCalendarDate(`${key}-01`, { month: 'long', year: 'numeric' });
		months.set(key, {
			key,
			label: label.charAt(0).toUpperCase() + label.slice(1),
			movements: [movement]
		});
	}

	return [...months.values()];
}

/**
 * El portafolio que se propone para un movimiento en una plataforma y una
 * moneda.
 *
 * Si la cuenta ya existe, el de su saldo mayor: un depósito cae sobre el dinero
 * que ya estaba, en lugar de abrir otro saldo en otro portafolio y partir la
 * cuenta en dos. Si no existe, se queda `fallback`, lo que ya estaba elegido.
 */
export function suggestCashPortfolio(
	balances: CashBalance[],
	sourceId: string,
	currency: string,
	fallback: string
): string {
	let best: CashBalance | undefined;

	for (const row of balances) {
		if (row.sourceId !== sourceId || row.currency !== currency) continue;
		if (!best || (parseFloat(row.balance) || 0) > (parseFloat(best.balance) || 0)) best = row;
	}

	return best?.portfolioId ?? fallback;
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
	 * Cuentas con dinero que no tenían tasa a `currency` y quedaron fuera de
	 * `total`: sumarlas a valor nominal mezclaría monedas.
	 */
	unconverted: number;
	/** Cuentas con saldo, de todas las listadas. */
	funded: number;
	/** Por moneda, de la que más pesa a la que menos. */
	byCurrency: CashCurrencyTotal[];
	/**
	 * Los intereses abonados este mes, en `currency`. Como el total, sin los
	 * saldos que no tenían tasa de cambio.
	 */
	interestThisMonth: number;
}

/**
 * El total del efectivo y su reparto por moneda.
 *
 * El backend convierte cada saldo por separado y marca los que no pudo. Esos se
 * quedan fuera del total —es la regla del listado de portafolios— pero siguen
 * en el reparto por moneda con su importe propio, que es donde sí significan
 * algo.
 *
 * Lo que se cuenta son cuentas, no saldos: una cuenta cuyo dinero suma en dos
 * portafolios sigue siendo una, que es como aparece en la tabla.
 */
export function summarizeCash(balances: CashBalance[], fallbackCurrency: string): CashSummary {
	let total = 0;
	let unconverted = 0;
	let funded = 0;
	const groups = new Map<string, CashCurrencyTotal>();

	for (const account of groupCashAccounts(balances)) {
		if (account.balance !== 0) funded += 1;

		total += account.value;
		if (!account.fxConverted && account.balance !== 0) unconverted += 1;

		const group = groups.get(account.currency) ?? {
			currency: account.currency,
			balance: 0,
			value: 0,
			accounts: 0
		};
		group.balance += account.balance;
		group.value += account.value;
		group.accounts += 1;
		groups.set(account.currency, group);
	}

	const byCurrency = [...groups.values()].sort(
		(a, b) => b.value - a.value || b.balance - a.balance || a.currency.localeCompare(b.currency)
	);

	// Por saldo, no por cuenta: cada saldo trae su parte del mes ya convertida.
	const interestThisMonth = balances.reduce(
		(sum, row) => (row.fxConverted ? sum + (parseFloat(row.interestThisMonthValue) || 0) : sum),
		0
	);

	return {
		total,
		currency: balances[0]?.displayCurrency || fallbackCurrency,
		unconverted,
		funded,
		byCurrency,
		interestThisMonth
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
