/**
 * Helpers puros de los depósitos a tasa fija (000048): lo que rinde uno a lo
 * largo de su plazo, el calendario de ese plazo, y cómo se lee en una línea.
 *
 * Un depósito es un bolsillo que conserva la tasa del día en que se abrió, así
 * que su tasa no tiene tramos ni versiones: la cuenta es la fórmula cerrada, y
 * no hace falta recorrer el año día a día como en `rates.ts`.
 */

import type { CashRate } from '$lib/api/types';
import { calendarDay, dailyRateFromAnnual, dayAndMonth, formatAnnualRate } from './rates';

/**
 * Lo que rinde `balance` en `days` días a una tasa sin tramos, neto de
 * retención: `saldo × ((1 + diaria × (1 − retención))^días − 1)`.
 *
 * Es lo que gana un depósito a tasa fija a lo largo de su plazo. Los intereses
 * de cada día cuentan en el saldo del siguiente aunque la entidad los abone al
 * vencer, así que la cuenta es la misma se abone como se abone: diez millones
 * al 10 % E.A. a 90 días con 4 % de retención rinden unos 228.176.
 */
export function projectInterestOverDays(
	balance: number,
	annualPct: number,
	days: number,
	withholdingPct = 0
): number {
	if (!(balance > 0) || !(annualPct > 0) || !(days > 0)) return 0;

	const kept = 1 - Math.min(Math.max(withholdingPct, 0), 100) / 100;
	const daily = dailyRateFromAnnual(annualPct) * kept;

	return balance * (Math.pow(1 + daily, days) - 1);
}

/** Cuántos días enteros van de `from` a `to`, contados en UTC. */
export function daysBetween(from: string, to: string): number {
	const start = Date.parse(`${from.slice(0, 10)}T00:00:00Z`);
	const end = Date.parse(`${to.slice(0, 10)}T00:00:00Z`);
	if (Number.isNaN(start) || Number.isNaN(end)) return 0;

	return Math.round((end - start) / 86_400_000);
}

/** `2026-09-01` + 90 → `2026-11-30`, contado en UTC. */
export function addCalendarDays(date: string, days: number): string {
	const at = new Date(`${calendarDay(date)}T00:00:00Z`);
	if (Number.isNaN(at.getTime())) return '';
	at.setUTCDate(at.getUTCDate() + days);

	return at.toISOString().slice(0, 10);
}

/**
 * Un depósito a tasa fija en una línea: «10 % E.A. · tasa fija · vence 30 nov».
 *
 * No lleva botón detrás, a diferencia de la tasa de una cuenta: la de un
 * depósito es la del día en que se abrió y no se cambia. Lo que se hace con él
 * es cancelarlo.
 */
export function describeFixedDeposit(
	pocket: { maturesOn: string | null; closedOn: string | null },
	rate: CashRate | null
): string {
	const parts = [rate ? formatAnnualRate(rate.annualRatePct) : 'Sin tasa', 'tasa fija'];

	if (rate?.posting === 'at_maturity') parts.push('abono al vencer');

	if (pocket.closedOn) {
		parts.push(`cerrado el ${dayAndMonth(pocket.closedOn)}`);
	} else if (pocket.maturesOn) {
		parts.push(`vence el ${dayAndMonth(pocket.maturesOn)}`);
	} else {
		parts.push('sin plazo');
	}

	return parts.join(' · ');
}
