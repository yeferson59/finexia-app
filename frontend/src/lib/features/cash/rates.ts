/**
 * Helpers puros de la rentabilidad del efectivo: la tasa que rinde cada cuenta
 * y lo que ganaría con ella. Sin dependencias de Svelte ni de red; el contrato
 * `CashRate` viene de `$lib/api/types`.
 */

import type { CashRate } from '$lib/api/types';
import { formatCalendarDate } from '$lib/shared/format/date';

export type { CashRate };

/** Los días con que la E.A. se pasa a diaria, como hacen los bancos. */
const DAYS_PER_YEAR = 365;

/** «9,25% E.A.»: la tasa como la publica la entidad. */
export function formatAnnualRate(pct: string | number): string {
	const value = typeof pct === 'number' ? pct : parseFloat(pct) || 0;
	return `${value.toLocaleString('es-CO', { maximumFractionDigits: 4 })}% E.A.`;
}

/**
 * La tasa diaria equivalente a una efectiva anual, como fracción:
 * `(1 + EA)^(1/365) − 1`. Un 9 % E.A. da 0,000236131.
 */
export function dailyRateFromAnnual(annualPct: number): number {
	return Math.pow(1 + annualPct / 100, 1 / DAYS_PER_YEAR) - 1;
}

/**
 * Cada cuánto capitaliza una tasa nominal, para pasarla a efectiva anual.
 *
 * Una entidad que dice «12 % nominal mes vencido» no paga 12 % al año: paga
 * 1 % cada mes sobre el saldo con los intereses del mes anterior dentro, que
 * son 12,68 % E.A. El formulario guarda siempre la efectiva, así que esto es lo
 * que traduce lo que dice el folleto.
 */
export const NOMINAL_PERIODS: { value: number; label: string }[] = [
	{ value: 12, label: 'Mensual' },
	{ value: 4, label: 'Trimestral' },
	{ value: 2, label: 'Semestral' },
	{ value: 365, label: 'Diaria' }
];

/**
 * La efectiva anual equivalente a una nominal que capitaliza `periods` veces al
 * año: `(1 + nominal/periods)^periods − 1`, en porcentaje.
 *
 * Un 12 % nominal mensual da 12,6825 % E.A. Sin tasa o sin periodos da cero.
 */
export function annualFromNominal(nominalPct: number, periods: number): number {
	if (!(nominalPct > 0) || !(periods > 0)) return 0;

	return (Math.pow(1 + nominalPct / 100 / periods, periods) - 1) * 100;
}

/** Lo que rendiría un saldo en un día, en 30 y en un año. */
export interface InterestProjection {
	day: number;
	month: number;
	year: number;
}

/**
 * Lo que rendiría `balance` con una tasa, neto de retención y con los intereses
 * de cada día sumados al saldo del siguiente.
 *
 * Es una proyección sobre el saldo de hoy: no cuenta los depósitos ni los
 * retiros que vengan. Sin saldo o sin tasa rinde cero.
 *
 * Con tope, solo rinde la parte del saldo que cabe en él. El tope es de la
 * cuenta entera, así que la proyección de una cuenta repartida en varios
 * portafolios se calcula sobre la suma, como hace el backend.
 */
export function projectInterest(
	balance: number,
	annualPct: number,
	withholdingPct = 0,
	maxBalance: number | null = null
): InterestProjection {
	const earning = maxBalance !== null && maxBalance > 0 ? Math.min(balance, maxBalance) : balance;

	if (!(earning > 0) || !(annualPct > 0)) return { day: 0, month: 0, year: 0 };

	const kept = 1 - Math.min(Math.max(withholdingPct, 0), 100) / 100;
	const daily = dailyRateFromAnnual(annualPct) * kept;
	const earned = (days: number) => earning * (Math.pow(1 + daily, days) - 1);

	return { day: earned(1), month: earned(30), year: earned(DAYS_PER_YEAR) };
}

/** Las versiones de la tasa de una cuenta, leídas en un día. */
export interface CashAccountRate {
	/** La que rige ese día. */
	current: CashRate | null;
	/** La primera que empieza después, si ya está anotada. */
	upcoming: CashRate | null;
	/** La versión más reciente: la única que se corrige, pausa o borra. */
	latest: CashRate | null;
	/**
	 * El último día con intereses calculados en cualquiera de sus versiones;
	 * `null` si la cuenta nunca rindió. Es lo que dice si hay algo que
	 * recalcular.
	 */
	accruedThrough: string | null;
}

const calendarDay = (iso: string) => iso.slice(0, 10);

/**
 * La tasa de una cuenta —plataforma y moneda— en `today` (`2026-09-14`).
 *
 * Las fechas se comparan como días del calendario, que es como las guarda el
 * backend: una tasa desde el 15 no puede empezar el 14 en una zona al oeste de
 * Greenwich.
 */
export function cashAccountRate(
	rates: CashRate[],
	sourceId: string,
	currency: string,
	today: string
): CashAccountRate {
	const versions = rates.filter((r) => r.sourceId === sourceId && r.currency === currency);

	const current =
		versions.find(
			(r) =>
				calendarDay(r.effectiveFrom) <= today && (!r.endedOn || calendarDay(r.endedOn) >= today)
		) ?? null;

	const upcoming =
		versions
			.filter((r) => calendarDay(r.effectiveFrom) > today)
			.sort((a, b) => a.effectiveFrom.localeCompare(b.effectiveFrom))[0] ?? null;

	const accruedThrough =
		versions
			.map((r) => r.accruedThrough)
			.filter((day): day is string => !!day)
			.sort()
			.at(-1) ?? null;

	return { current, upcoming, latest: versions.find((r) => r.latest) ?? null, accruedThrough };
}

/** `2026-09-14` → `2026-09-15`, contado en UTC para que no lo mueva la zona. */
function nextCalendarDay(date: string): string {
	const next = new Date(`${calendarDay(date)}T00:00:00Z`);
	next.setUTCDate(next.getUTCDate() + 1);
	return next.toISOString().slice(0, 10);
}

const dayAndMonth = (iso: string) =>
	formatCalendarDate(calendarDay(iso), { day: 'numeric', month: 'short' });

/**
 * La tasa de una cuenta en una línea: la que rige y lo que cambia después, o
 * desde cuándo no rinde. `null` si nunca tuvo tasa.
 */
export function describeCashAccountRate({
	current,
	upcoming,
	latest
}: CashAccountRate): string | null {
	if (current) {
		const now =
			formatAnnualRate(current.annualRatePct) +
			(current.posting === 'monthly' ? ' · abono mensual' : '');
		if (upcoming) {
			return `${now} · ${formatAnnualRate(upcoming.annualRatePct)} desde el ${dayAndMonth(upcoming.effectiveFrom)}`;
		}
		if (current.endedOn) return `${now} hasta el ${dayAndMonth(current.endedOn)}`;
		return now;
	}

	if (upcoming) {
		return `${formatAnnualRate(upcoming.annualRatePct)} desde el ${dayAndMonth(upcoming.effectiveFrom)}`;
	}

	if (latest?.endedOn)
		return `Sin rentabilidad desde el ${dayAndMonth(nextCalendarDay(latest.endedOn))}`;

	return null;
}

/** Lo que está rindiendo el efectivo, en una cifra. */
export interface CashYield {
	/** La tasa media de los saldos que rinden, ponderada por lo que guarda cada uno. */
	pct: number;
	/** Cuántas cuentas con dinero no rinden nada. */
	idle: number;
}

/**
 * La tasa media del efectivo y cuántas cuentas están sin ella.
 *
 * La media deja fuera las cuentas sin tasa en vez de promediarles un cero: lo
 * que rinde el dinero que rinde y cuánto dinero está parado son dos preguntas
 * distintas, y esto responde las dos por separado. Se pondera por el valor en
 * la moneda de la pantalla, que es lo único que se puede sumar entre monedas;
 * una cuenta sin tasa de cambio no pesa, como en el total.
 *
 * `null` si no hay ninguna cuenta con dinero.
 */
export function cashYield(
	accounts: { sourceId: string; currency: string; balance: number; value: number }[],
	rates: CashRate[],
	today: string
): CashYield | null {
	let weighted = 0;
	let earning = 0;
	let idle = 0;
	let funded = 0;

	for (const account of accounts) {
		if (account.balance === 0) continue;
		funded += 1;

		const current = cashAccountRate(rates, account.sourceId, account.currency, today).current;
		if (!current) {
			idle += 1;
			continue;
		}

		earning += account.value;
		weighted += account.value * (parseFloat(current.annualRatePct) || 0);
	}

	if (funded === 0) return null;

	return { pct: earning > 0 ? weighted / earning : 0, idle };
}

/** Lo que se dice cuando el fallo no es de ningún campo en concreto. */
export const CASH_RATE_FALLBACK = 'No pudimos guardar la tasa. Vuelve a intentarlo en un momento.';
export const CASH_RECALCULATE_FALLBACK =
	'No pudimos recalcular los intereses. Vuelve a intentarlo en un momento.';

/**
 * El motivo de una tasa rechazada, en la frase que ve el usuario. Mismo criterio
 * que `cashErrorMessage`: se traducen los casos que el formulario puede
 * provocar, y lo demás es un fallo que no se arregla cambiando un campo.
 *
 * `fallback` es esa última frase. Las acciones sobre la tasa y el recálculo
 * comparten los motivos —la cuenta es la misma— pero no lo que se intentaba
 * hacer, así que cada una pone la suya.
 */
export function cashRateErrorMessage(
	status: number,
	details = '',
	fallback = CASH_RATE_FALLBACK
): string {
	if (details.includes('already earned interest')) {
		const from = details.match(/can (?:stop|start) from (\d{4}-\d{2}-\d{2})/)?.[1];
		return from
			? `Los intereses ya se calcularon con esta tasa hasta la víspera del ${dayAndMonth(from)}. Elige ese día o uno posterior.`
			: 'Esta tasa ya generó intereses, así que no se puede corregir ni borrar. Anota una tasa nueva o páusala.';
	}
	if (details.includes('only the latest version')) {
		return 'Solo se puede cambiar la versión más reciente de la tasa. Recarga la página.';
	}
	if (details.includes('already starts on or after')) {
		return 'Ya hay una tasa que empieza ese día o después. Corrígela o bórrala antes de anotar otra.';
	}
	if (details.includes('platform is inactive')) {
		return 'Esa plataforma está inactiva. Actívala para darle una tasa.';
	}
	if (details.includes('delete it instead')) {
		return 'La tasa todavía no ha empezado ese día. Si no la quieres, bórrala.';
	}
	if (details.includes('cannot be before')) {
		return 'Elige hoy o una fecha posterior: los días pasados no se recalculan.';
	}
	if (details.includes('annualRatePct takes at most')) {
		return 'Escribe la tasa con hasta cuatro decimales.';
	}
	if (details.includes('withholdingPct takes at most')) {
		return 'Escribe la retención con hasta dos decimales.';
	}
	if (details.includes('maxBalance takes at most')) {
		return 'Escribe el tope con hasta ocho decimales.';
	}
	if (details.includes('maxBalance must be greater than 0')) {
		return 'El tope tiene que ser mayor que cero. Déjalo vacío si la cuenta rinde sobre todo el saldo.';
	}
	if (details.includes('from cannot be in the future')) {
		return 'Elige hoy o un día pasado: no hay intereses que recalcular más adelante.';
	}
	if (details.includes('platform not found')) {
		return 'No encontramos esa plataforma. Recarga la página.';
	}
	if (status === 404) {
		return fallback === CASH_RATE_FALLBACK
			? 'Esa tasa ya no existe. Recarga la página.'
			: 'No encontramos esa cuenta. Recarga la página.';
	}

	return fallback;
}
