/**
 * Helpers puros de la rentabilidad del efectivo: la tasa que rinde cada cuenta
 * y lo que ganaría con ella. Sin dependencias de Svelte ni de red; el contrato
 * `CashRate` viene de `$lib/api/types`.
 */

import type { CashRate } from '$lib/api/types';
import { formatCalendarDate } from '$lib/shared/format/date';

export type { CashRate };

/** Los días con que la E.A. se pasa a diaria, como hacen los bancos. */
export const DAYS_PER_YEAR = 365;

/** «9,25% E.A.»: la tasa como la publica la entidad. */
export function formatAnnualRate(pct: string | number): string {
	const value = typeof pct === 'number' ? pct : parseFloat(pct) || 0;
	return `${value.toLocaleString('es-CO', { maximumFractionDigits: 4 })}% E.A.`;
}

/** «8%»: un tramo, sin repetir el «E.A.» que ya dijo la tasa principal. */
function formatRatePct(pct: number): string {
	return `${pct.toLocaleString('es-CO', { maximumFractionDigits: 4 })}%`;
}

/**
 * La tasa diaria equivalente a una efectiva anual, como fracción:
 * `(1 + EA)^(1/365) − 1`. Un 9 % E.A. da 0,000236131.
 */
export function dailyRateFromAnnual(annualPct: number): number {
	return Math.pow(1 + annualPct / 100, 1 / DAYS_PER_YEAR) - 1;
}

/** Un tramo de la tasa: desde qué saldo de la cuenta rige, y a qué tasa E.A. */
export interface RateTier {
	fromBalance: number;
	annualRatePct: number;
}

/** Los tramos de una versión como números, del más bajo al más alto. */
export function rateTiers(rate: Pick<CashRate, 'tiers'>): RateTier[] {
	return rate.tiers
		.map((t) => ({
			fromBalance: parseFloat(t.fromBalance) || 0,
			annualRatePct: parseFloat(t.annualRatePct) || 0
		}))
		.sort((a, b) => a.fromBalance - b.fromBalance);
}

/**
 * Lo que rinde en un día una cuenta que guarda `balance`, antes de retención:
 * la tasa diaria de cada tramo sobre la parte del saldo que cae en él. La tasa
 * principal rige desde cero, y sin tramos es la única. Es la misma cuenta que
 * hace el backend.
 */
export function dailyInterest(balance: number, annualPct: number, tiers: RateTier[] = []): number {
	let earned = 0;
	let from = 0;
	let pct = annualPct;

	for (let i = 0; from < balance; i++) {
		const next = tiers[i];
		const to = next && next.fromBalance < balance ? next.fromBalance : balance;

		// Un tramo al 0 % es un tope: lo que cae en él no rinde.
		if (pct > 0) earned += (to - from) * dailyRateFromAnnual(pct);

		if (!next) break;
		from = next.fromBalance;
		pct = next.annualRatePct;
	}

	return earned;
}

/**
 * La tasa E.A. a la que rinde todo el saldo de una cuenta con tramos: su día
 * capitalizado un año, `(1 + día / saldo)^365 − 1`, en porcentaje. Sin tramos o
 * sin saldo es la tasa principal.
 *
 * Doce por ciento hasta cinco millones y ocho sobre el resto dan, con ocho
 * millones, un 10,48 % E.A.
 */
export function effectiveAnnualRate(
	balance: number,
	annualPct: number,
	tiers: RateTier[] = []
): number {
	if (tiers.length === 0 || !(balance > 0)) return annualPct;

	return (
		(Math.pow(1 + dailyInterest(balance, annualPct, tiers) / balance, DAYS_PER_YEAR) - 1) * 100
	);
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
 * Con tramos, cada día rinde por tramos sobre el saldo de ese día, como en el
 * backend: lo ganado puede caer en el tramo siguiente, y por encima de un tope
 * ya no capitaliza. Los tramos son de la cuenta entera, así que la proyección
 * de una cuenta repartida en varios portafolios se calcula sobre la suma.
 */
export function projectInterest(
	balance: number,
	annualPct: number,
	withholdingPct = 0,
	tiers: RateTier[] = []
): InterestProjection {
	if (!(balance > 0) || !(annualPct > 0)) return { day: 0, month: 0, year: 0 };

	const kept = 1 - Math.min(Math.max(withholdingPct, 0), 100) / 100;

	// Sin tramos, la fórmula cerrada da lo mismo sin recorrer el año.
	if (tiers.length === 0) {
		const daily = dailyRateFromAnnual(annualPct) * kept;
		const earned = (days: number) => balance * (Math.pow(1 + daily, days) - 1);

		return { day: earned(1), month: earned(30), year: earned(DAYS_PER_YEAR) };
	}

	const earned: Record<number, number> = {};
	let held = balance;
	for (let day = 1; day <= DAYS_PER_YEAR; day++) {
		held += dailyInterest(held, annualPct, tiers) * kept;
		earned[day] = held - balance;
	}

	return { day: earned[1], month: earned[30], year: earned[DAYS_PER_YEAR] };
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

/** El día del calendario de una fecha ISO, que es como el backend la guarda. */
export const calendarDay = (iso: string) => iso.slice(0, 10);

/**
 * La tasa de una cuenta —plataforma, moneda y bolsillo— en `today`
 * (`2026-09-14`). `pocketId` en `null` es la cuenta principal, que es lo que
 * había antes de que hubiera bolsillos.
 *
 * Las fechas se comparan como días del calendario, que es como las guarda el
 * backend: una tasa desde el 15 no puede empezar el 14 en una zona al oeste de
 * Greenwich.
 */
export function cashAccountRate(
	rates: CashRate[],
	sourceId: string,
	currency: string,
	today: string,
	pocketId: string | null = null
): CashAccountRate {
	const versions = rates.filter(
		(r) => r.sourceId === sourceId && r.currency === currency && (r.pocketId ?? null) === pocketId
	);

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
export function nextCalendarDay(date: string): string {
	const next = new Date(`${calendarDay(date)}T00:00:00Z`);
	next.setUTCDate(next.getUTCDate() + 1);
	return next.toISOString().slice(0, 10);
}

/** «30 de nov»: una fecha en la forma corta que usan las líneas de la pantalla. */
export const dayAndMonth = (iso: string) =>
	formatCalendarDate(calendarDay(iso), { day: 'numeric', month: 'short' });

/**
 * Cómo paga una versión por encima de su tasa: «hasta $ 5.000.000 · 8 %
 * después». La tasa principal rige desde cero, así que cada tramo cierra el
 * que viene antes —«hasta» ese saldo— y el último dice qué se gana de ahí en
 * adelante. Un tramo al 0 % no se nombra: no rinde, y decir «hasta» ya lo
 * cuenta.
 *
 * Los importes pasan por `money`, que con el modo oculto los enmascara, porque
 * la línea de la cuenta se ve con él puesto.
 */
function tiersNote(rate: CashRate, money: (amount: number) => string): string {
	const tiers = rateTiers(rate);
	if (tiers.length === 0) return '';

	const parts = tiers.map(
		(tier, i) =>
			(i === 0 ? '' : `${formatRatePct(tiers[i - 1].annualRatePct)} `) +
			`hasta ${money(tier.fromBalance)}`
	);

	const last = tiers[tiers.length - 1];
	if (last.annualRatePct > 0) parts.push(`${formatRatePct(last.annualRatePct)} después`);

	return ` ${parts.join(' · ')}`;
}

/** Cada cuánto abona la cuenta, cuando no es cada día. */
function postingNote(posting: CashRate['posting']): string {
	if (posting === 'monthly') return ' · abono mensual';
	if (posting === 'at_maturity') return ' · abono al vencer';

	return '';
}

/**
 * La tasa de una cuenta en una línea: la que rige, sus tramos y lo que cambia
 * después, o desde cuándo no rinde. `null` si nunca tuvo tasa.
 *
 * `money` da formato al saldo desde el que rige un tramo; por defecto no se
 * nombra ninguno, para quien solo quiera la tasa.
 */
export function describeCashAccountRate(
	{ current, upcoming, latest }: CashAccountRate,
	money: (amount: number) => string = (amount) => String(amount)
): string | null {
	if (current) {
		const now =
			formatAnnualRate(current.annualRatePct) +
			tiersNote(current, money) +
			postingNote(current.posting);
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
 * una cuenta sin tasa de cambio no pesa, como en el total. Una cuenta con
 * tramos cuenta con la tasa a la que rinde todo su saldo.
 *
 * `null` si no hay ninguna cuenta con dinero.
 */
export function cashYield(
	accounts: {
		sourceId: string;
		currency: string;
		balance: number;
		value: number;
		pocketId?: string | null;
	}[],
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

		const current = cashAccountRate(
			rates,
			account.sourceId,
			account.currency,
			today,
			account.pocketId ?? null
		).current;
		if (!current) {
			idle += 1;
			continue;
		}

		earning += account.value;
		weighted +=
			account.value *
			effectiveAnnualRate(
				account.balance,
				parseFloat(current.annualRatePct) || 0,
				rateTiers(current)
			);
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
	// Antes que los de la tasa principal: sus mensajes contienen los de ella.
	if (details.includes('tiers take at most')) {
		return 'Una tasa tiene como mucho 10 tramos.';
	}
	if (details.includes('tiers must go up')) {
		return 'Cada tramo tiene que empezar en un saldo mayor que el anterior.';
	}
	if (details.includes('tier fromBalance takes at most')) {
		return 'Escribe el saldo de cada tramo con hasta ocho decimales.';
	}
	if (details.includes('tier fromBalance must be greater than 0')) {
		return 'El saldo desde el que rige un tramo tiene que ser mayor que cero.';
	}
	if (details.includes('tier annualRatePct takes at most')) {
		return 'Escribe la tasa de cada tramo con hasta cuatro decimales.';
	}
	if (details.includes('tier annualRatePct must be')) {
		return 'La tasa de un tramo va de 0 % a 100 %.';
	}
	if (details.includes('annualRatePct takes at most')) {
		return 'Escribe la tasa con hasta cuatro decimales.';
	}
	if (details.includes('withholdingPct takes at most')) {
		return 'Escribe la retención con hasta dos decimales.';
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
