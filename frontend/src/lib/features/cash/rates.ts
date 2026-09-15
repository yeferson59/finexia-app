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
 */
export function projectInterest(
	balance: number,
	annualPct: number,
	withholdingPct = 0
): InterestProjection {
	if (!(balance > 0) || !(annualPct > 0)) return { day: 0, month: 0, year: 0 };

	const kept = 1 - Math.min(Math.max(withholdingPct, 0), 100) / 100;
	const daily = dailyRateFromAnnual(annualPct) * kept;
	const earned = (days: number) => balance * (Math.pow(1 + daily, days) - 1);

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

	return { current, upcoming, latest: versions.find((r) => r.latest) ?? null };
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
		const now = formatAnnualRate(current.annualRatePct);
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

/**
 * El motivo de una tasa rechazada, en la frase que ve el usuario. Mismo criterio
 * que `cashErrorMessage`: se traducen los casos que el formulario puede
 * provocar, y lo demás es un fallo que no se arregla cambiando un campo.
 */
export function cashRateErrorMessage(status: number, details = ''): string {
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
	if (details.includes('platform not found')) {
		return 'No encontramos esa plataforma. Recarga la página.';
	}
	if (status === 404) {
		return 'Esa tasa ya no existe. Recarga la página.';
	}

	return 'No pudimos guardar la tasa. Vuelve a intentarlo en un momento.';
}
