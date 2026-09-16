/**
 * Helpers puros de cómo se enseña lo que rinde el efectivo: el mapa de
 * rendimiento del resumen, su escala, la tasa de una cuenta en dos líneas, el
 * nombre de su moneda y el plazo de un depósito. Sin dependencias de Svelte ni
 * de red.
 *
 * El mapa pinta cada cajón con dinero como un bloque: el ancho es lo que guarda
 * y el alto, la tasa a la que rinde hoy. El área es entonces lo que gana, y el
 * dinero parado se ve como lo que es —ancho sin altura—, que es la pregunta que
 * esta pantalla responde: si vale la pena mover algo.
 */

import type { CashRate } from '$lib/api/types';
import type { CashAccount } from './cash';
import {
	cashAccountRate,
	dayAndMonth,
	effectiveAnnualRate,
	formatAnnualRate,
	nextCalendarDay,
	projectInterest,
	rateTiers,
	type CashAccountRate
} from './rates';
import { daysBetween } from './deposits';

/** Un cajón con dinero en el mapa de rendimiento. */
export interface CashYieldBlock {
	key: string;
	sourceId: string;
	sourceName: string;
	currency: string;
	pocketId: string | null;
	/** Vacío en la cuenta principal. */
	pocketName: string;
	/** Un depósito a tasa fija: su tasa no cambia hasta que vence. */
	fixed: boolean;
	/** Lo que guarda, en `currency`. */
	balance: number;
	/** Lo mismo en la moneda de la pantalla: es lo que da el ancho. */
	value: number;
	/** La tasa E.A. a la que rinde todo el saldo hoy; 0 si está parado. */
	pct: number;
	/** Su parte del efectivo que entra en el mapa, de 0 a 1. */
	share: number;
	/** Lo que rendiría en un año sobre el saldo de hoy, neto de retención, en `currency`. */
	yearly: number;
}

export interface CashYieldMap {
	/** Los que rinden, de la tasa más alta a la más baja, y luego los parados. */
	blocks: CashYieldBlock[];
	/** La parte del efectivo que rinde algo, de 0 a 1. */
	earningShare: number;
	/** La parte que está parada. */
	idleShare: number;
	/** La tasa más alta del mapa. */
	top: number;
}

/**
 * El mapa de rendimiento de unas cuentas —las principales y sus bolsillos, cada
 * una por su lado, como las devuelve `groupCashAccounts`— leído en `today`.
 *
 * Solo entran las que tienen dinero y un valor en la moneda de la pantalla: sin
 * tasa de cambio no hay ancho que darles, y un bloque a valor nominal mezclaría
 * monedas. Con tramos, la altura es la tasa a la que rinde todo el saldo, no la
 * del primer tramo.
 *
 * Se ordenan de la tasa más alta a la más baja, así que el mapa baja como una
 * escalera y lo que queda a la derecha, pegado al suelo, es lo que no rinde.
 */
export function cashYieldMap(
	accounts: CashAccount[],
	rates: CashRate[],
	today: string
): CashYieldMap {
	const funded = accounts.filter((a) => a.balance > 0 && a.fxConverted && a.value > 0);
	const total = funded.reduce((sum, a) => sum + a.value, 0);

	const blocks = funded.map((account): CashYieldBlock => {
		const current = cashAccountRate(
			rates,
			account.sourceId,
			account.currency,
			today,
			account.pocketId
		).current;
		const annual = current ? parseFloat(current.annualRatePct) || 0 : 0;
		const tiers = current ? rateTiers(current) : [];

		return {
			key: account.key,
			sourceId: account.sourceId,
			sourceName: account.sourceName,
			currency: account.currency,
			pocketId: account.pocketId,
			pocketName: account.pocketName,
			fixed: account.pocketKind === 'fixed',
			balance: account.balance,
			value: account.value,
			pct: current ? effectiveAnnualRate(account.balance, annual, tiers) : 0,
			share: total > 0 ? account.value / total : 0,
			yearly: current
				? projectInterest(account.balance, annual, parseFloat(current.withholdingPct) || 0, tiers)
						.year
				: 0
		};
	});

	blocks.sort((a, b) => b.pct - a.pct || b.value - a.value || a.key.localeCompare(b.key));

	const earningShare = blocks.filter((b) => b.pct > 0).reduce((sum, b) => sum + b.share, 0);

	return {
		blocks,
		earningShare,
		idleShare: blocks.length > 0 ? 1 - earningShare : 0,
		top: blocks[0]?.pct ?? 0
	};
}

/** Los pasos que puede dar la escala de tasas: cifras que se leen de un vistazo. */
const RATE_STEPS = [0.5, 1, 2, 2.5, 3, 4, 5, 10, 20, 25, 50];

/**
 * La escala vertical del mapa: un techo redondo por encima de la tasa más alta
 * y, como mucho, cuatro marcas sobre el cero. Un 10,4 % da 0, 3, 6, 9 y 12.
 *
 * Sin ninguna tasa no hay nada que medir: la escala es solo el suelo.
 */
export function cashRateScale(top: number): { max: number; ticks: number[] } {
	if (!(top > 0)) return { max: 1, ticks: [0] };

	const step = RATE_STEPS.find((s) => top / s <= 4) ?? Math.ceil(top / 4 / 50) * 50;
	const max = Math.ceil(top / step - 1e-9) * step;
	const ticks: number[] = [];
	for (let i = 0; i * step <= max + 1e-9; i++) ticks.push(Number((i * step).toFixed(2)));

	return { max, ticks };
}

/** La tasa de una cuenta como se lee en su fila: la cifra y lo que la matiza. */
export interface CashRateLines {
	/** «9,25% E.A.», o el estado cuando hoy no rinde. */
	rate: string;
	/** Tramos, forma de abono y lo que cambia después; vacío si no hay nada. */
	detail: string;
	/** Si rinde hoy. */
	earning: boolean;
}

const ratePct = (pct: number) => `${pct.toLocaleString('es-CO', { maximumFractionDigits: 4 })}%`;

/**
 * Los tramos de una versión en palabras: «hasta $ 20.000.000, 6% después». La
 * tasa principal ya la dice la línea de arriba, así que el primer tramo solo
 * dice hasta dónde llega; un tramo al 0 % es un tope y no se nombra su tasa.
 */
function tiersDetail(rate: CashRate, money: (amount: number) => string): string {
	const tiers = rateTiers(rate);
	if (tiers.length === 0) return '';

	const parts = tiers.map(
		(tier, i) =>
			(i === 0 ? '' : `${ratePct(tiers[i - 1].annualRatePct)} `) +
			`hasta ${money(tier.fromBalance)}`
	);
	const last = tiers[tiers.length - 1];
	if (last.annualRatePct > 0) parts.push(`${ratePct(last.annualRatePct)} después`);

	return parts.join(', ');
}

const POSTING_DETAIL: Record<CashRate['posting'], string> = {
	daily: '',
	monthly: 'abono mensual',
	at_maturity: 'abono al vencer'
};

/**
 * La tasa de una cuenta en dos líneas, para su fila: la que rige arriba y, en
 * la de abajo, sus tramos, cada cuánto abona y lo que cambia después. `null` si
 * la cuenta nunca tuvo tasa.
 *
 * Dice lo mismo que `describeCashAccountRate`, partido donde la fila lo parte:
 * la cifra se compara en vertical con las de las otras cuentas, y el matiz no.
 */
export function cashRateLines(
	{ current, upcoming, latest }: CashAccountRate,
	money: (amount: number) => string = (amount) => String(amount)
): CashRateLines | null {
	if (current) {
		const detail = [tiersDetail(current, money), POSTING_DETAIL[current.posting]];
		if (upcoming) {
			detail.push(
				`${formatAnnualRate(upcoming.annualRatePct)} desde el ${dayAndMonth(upcoming.effectiveFrom)}`
			);
		} else if (current.endedOn) {
			detail.push(`hasta el ${dayAndMonth(current.endedOn)}`);
		}

		return {
			rate: formatAnnualRate(current.annualRatePct),
			detail: detail.filter(Boolean).join(', '),
			earning: true
		};
	}

	if (upcoming) {
		return {
			rate: 'Sin rendir hoy',
			detail: `${formatAnnualRate(upcoming.annualRatePct)} desde el ${dayAndMonth(upcoming.effectiveFrom)}`,
			earning: false
		};
	}

	if (latest?.endedOn) {
		return {
			rate: 'En pausa',
			detail: `sin rendir desde el ${dayAndMonth(nextCalendarDay(latest.endedOn))}`,
			earning: false
		};
	}

	return null;
}

/**
 * El nombre de una moneda con mayúscula inicial: «Peso colombiano», «Euro». Es
 * el de una cuenta sola en su plataforma, donde «Cuenta principal» no diría
 * nada. El código tal cual si el navegador no lo conoce.
 */
export function cashCurrencyName(code: string): string {
	try {
		const name = new Intl.DisplayNames('es', { type: 'currency' }).of(code) ?? code;
		return name.charAt(0).toUpperCase() + name.slice(1);
	} catch {
		return code;
	}
}

/** Dónde está un depósito dentro de su plazo. */
export interface CashDepositTerm {
	/** Días del plazo entero. */
	total: number;
	/** Los que ya pasaron, sin salirse del plazo. */
	elapsed: number;
	/** Los que faltan; 0 el día que vence y después. */
	left: number;
	/** `elapsed / total`, de 0 a 1. */
	progress: number;
}

/**
 * El plazo de un depósito leído en `today`. `null` si no vence —no hay línea
 * que recorrer— o si las fechas no dan un plazo.
 */
export function cashDepositTerm(
	openedOn: string,
	maturesOn: string | null,
	today: string
): CashDepositTerm | null {
	if (!maturesOn) return null;

	const total = daysBetween(openedOn, maturesOn);
	if (!(total > 0)) return null;

	const elapsed = Math.min(Math.max(daysBetween(openedOn, today), 0), total);

	return { total, elapsed, left: total - elapsed, progress: elapsed / total };
}
