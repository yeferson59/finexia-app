/**
 * Helpers puros de los fondos de inversión: cuánto hace que no se actualizan,
 * lo que ganan y cuánto se movió el valor de unidad entre dos marcas.
 *
 * Un fondo no rinde a una tasa: la entidad publica el valor de la unidad y la
 * rentabilidad es cómo se movió. Finexia no estima entre marcas, así que todo
 * lo que se calcula aquí sale de marcas escritas por el dueño.
 */
import type { Fund, FundMark, PublicFund } from '$lib/api/types';

export type {
	Fund,
	FundMark,
	FundMovement,
	FundPosition,
	PublicFund,
	PublicFundLink
} from '$lib/api/types';

/**
 * Pasados estos días sin marca, el valor se avisa como viejo. Un extracto
 * mensual llega unos días después del cierre: 35 deja pasar un mes y su
 * extracto sin avisar.
 */
export const FUND_STALE_DAYS = 35;

const DAY_MS = 86_400_000;

/** Días de calendario entre dos fechas `YYYY-MM-DD` (o ISO a medianoche UTC). */
export function calendarDaysBetween(from: string, to: string): number {
	const start = Date.parse(`${from.slice(0, 10)}T00:00:00Z`);
	const end = Date.parse(`${to.slice(0, 10)}T00:00:00Z`);

	return Math.round((end - start) / DAY_MS);
}

/** Cuántos días lleva el fondo con la misma marca; `null` si no tiene ninguna. */
export function daysSinceMark(fund: Pick<Fund, 'valuedOn'>, today: string): number | null {
	return fund.valuedOn ? Math.max(0, calendarDaysBetween(fund.valuedOn, today)) : null;
}

/** Si la última marca es tan vieja que el valor ya no se puede creer. */
export function isStale(fund: Pick<Fund, 'valuedOn'>, today: string): boolean {
	const days = daysSinceMark(fund, today);

	return days !== null && days > FUND_STALE_DAYS;
}

/**
 * Lo que gana el fondo sobre lo que costó, en su moneda y en porcentaje.
 *
 * `null` sin marca: el fondo vale lo que costó por definición, y un «0 %» diría
 * que no se movió cuando lo que pasa es que nadie dijo cuánto vale.
 */
export function fundGain(
	fund: Pick<Fund, 'value' | 'cost' | 'pricedAtCost'>
): { amount: number; pct: number | null } | null {
	if (fund.pricedAtCost) return null;

	const value = parseFloat(fund.value) || 0;
	const cost = parseFloat(fund.cost) || 0;
	const amount = value - cost;

	return { amount, pct: cost > 0 ? (amount / cost) * 100 : null };
}

/** Cómo se movió el valor de unidad de una marca a otra. */
export interface MarkChange {
	/** En porcentaje: `+0,69` es que la unidad vale 0,69 % más. */
	pct: number;
	days: number;
}

/** El cambio de `from` a `to`, o `null` si no se puede medir. */
export function markChange(
	from: Pick<FundMark, 'date' | 'unitValue'>,
	to: Pick<FundMark, 'date' | 'unitValue'>
): MarkChange | null {
	const before = parseFloat(from.unitValue);
	const after = parseFloat(to.unitValue);
	const days = calendarDaysBetween(from.date, to.date);

	if (!(before > 0) || !(after > 0) || days <= 0) return null;

	return { pct: (after / before - 1) * 100, days };
}

/**
 * Las marcas de la más reciente a la más vieja, cada una con su cambio sobre la
 * anterior. La más vieja no tiene contra qué compararse.
 */
export function marksWithChange(
	marks: FundMark[]
): { mark: FundMark; change: MarkChange | null }[] {
	const sorted = [...marks].sort((a, b) => b.date.localeCompare(a.date));

	return sorted.map((mark, i) => ({
		mark,
		change: i + 1 < sorted.length ? markChange(sorted[i + 1], mark) : null
	}));
}

/**
 * La marca anterior a un día: contra la que se compara una marca nueva antes de
 * guardarla. Una del mismo día no cuenta, porque la nueva la reemplaza.
 */
export function markBefore(marks: FundMark[], date: string): FundMark | null {
	const day = date.slice(0, 10);

	return (
		[...marks]
			.filter((m) => m.date.slice(0, 10) < day)
			.sort((a, b) => b.date.localeCompare(a.date))[0] ?? null
	);
}

/** Las plataformas en que está el fondo, sin repetir, para la línea de la tarjeta. */
export function fundPlatforms(fund: Pick<Fund, 'positions'>): string {
	return [...new Set(fund.positions.map((p) => p.sourceName || 'Sin plataforma'))].join(' · ');
}

/** Palabras que en un nombre propio van en minúscula. */
const SMALL_WORDS = new Set([
	'de',
	'del',
	'la',
	'las',
	'los',
	'el',
	'y',
	'e',
	'en',
	'con',
	'para',
	'por',
	'a'
]);

/** Siglas que se quedan en mayúsculas aunque sean largas. */
const ACRONYMS = new Set([
	'BBVA',
	'USD',
	'COP',
	'ESG',
	'TES',
	'CDT',
	'SURA',
	'BTG',
	'ETF',
	'S.A.',
	'SA'
]);

/**
 * El nombre de un fondo de la Superfinanciera como lo dice una persona. El
 * catálogo lo trae en mayúsculas y con la figura legal delante («FONDO DE
 * INVERSIÓN COLECTIVA ABIERTO FIDUCUENTA»); aquí queda «Fiducuenta».
 */
export function shortFundName(name: string): string {
	const stripped = name
		.trim()
		.replace(/^fondos? de inversi[oó]n(es)? colectiva(s)?\s*/i, '')
		.replace(/^fic\s+/i, '')
		.replace(/^(abierto|cerrado)\s+/i, '')
		.trim();

	const words = (stripped || name.trim()).split(/\s+/);

	return words
		.map((word, i) => {
			const lower = word.toLocaleLowerCase('es');
			if (i > 0 && SMALL_WORDS.has(lower)) return lower;
			if (ACRONYMS.has(word.toUpperCase())) return word.toUpperCase();

			return lower.charAt(0).toLocaleUpperCase('es') + lower.slice(1);
		})
		.join(' ');
}

/** Si un fondo puede tomar el valor que publica la Superfinanciera. */
export function canLinkFund(fund: Pick<Fund, 'tracking' | 'currency'>): boolean {
	return fund.tracking === 'units' && fund.currency === 'COP';
}

/**
 * La línea que distingue un tipo de participación de otro del mismo fondo: cada
 * uno tiene su valor de unidad, y el del extracto delata cuál es el propio.
 */
export function publicFundDetail(fund: Pick<PublicFund, 'participation' | 'investors'>): string {
	const investors = new Intl.NumberFormat('es-CO').format(fund.investors);

	return `Participación ${fund.participation} · ${investors} ${fund.investors === 1 ? 'inversionista' : 'inversionistas'}`;
}
