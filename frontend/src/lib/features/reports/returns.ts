/**
 * La aritmética de rentabilidad del centro de reportes.
 *
 * La serie de crecimiento (`GET /portfolios/growth`) da, por día, el valor de
 * mercado de la cuenta y el capital invertido a esa fecha. La distinción entre
 * los dos importes es la que sostiene este módulo: una cartera sube de valor
 * porque el mercado la revaloriza, pero también porque se le mete dinero, y
 * comparar el valor de dos fechas mezcla las dos cosas hasta llamar
 * «rentabilidad» a un depósito —de ahí venían los meses al +150 % de una cuenta
 * recién abierta—. Aquí cada retorno descuenta el aporte de su tramo con la
 * fórmula de Dietz modificada, así que lo que sale es rendimiento.
 *
 * Limitación conocida: una venta sale de la serie a valor de mercado pero solo
 * baja el capital invertido por su coste, así que el tramo en el que se vende
 * absorbe la plusvalía realizada. Es el precio de trabajar con valor y coste;
 * la alternativa —reconstruir los flujos desde las transacciones— pide un
 * endpoint que todavía no existe.
 *
 * Interno de la feature: `reports.ts` lo consume y publica los paneles.
 */

import type { GrowthDataPoint } from '$lib/api/types';

/** Retorno de un tramo de la serie, ya limpio de aportes y retiros. */
export interface PeriodReturn {
	/** Fecha del punto que cierra el tramo, `YYYY-MM-DD`. */
	date: string;
	/** Retorno fraccional del tramo: `0.012` es +1,2 %. */
	value: number;
	/** Días entre el punto anterior y este. */
	days: number;
}

const DAY_MS = 86_400_000;

export const DAYS_PER_YEAR = 365.25;

export function toNumber(raw: string | undefined): number {
	return Number.parseFloat(raw ?? '');
}

/** Días entre dos fechas `YYYY-MM-DD`, leídas en UTC para no perder uno por zona horaria. */
export function dayGap(from: string, to: string): number {
	const start = Date.parse(`${from}T00:00:00Z`);
	const end = Date.parse(`${to}T00:00:00Z`);
	if (!Number.isFinite(start) || !Number.isFinite(end)) return 0;
	return Math.round((end - start) / DAY_MS);
}

/** La serie ordenada por fecha; el backend ya la manda así, esto es defensa. */
export function sortedPoints(points: GrowthDataPoint[]): GrowthDataPoint[] {
	return [...points].sort((a, b) => a.date.localeCompare(b.date));
}

/**
 * Retorno de cada tramo de la serie por Dietz modificada.
 *
 * `flow` es lo que entró (o salió) durante el tramo, medido por la variación
 * del capital invertido, y se resta del numerador para que un depósito no
 * cuente como ganancia. El denominador lleva medio flujo porque el dinero nuevo
 * solo estuvo trabajando parte del tramo; con eso el primer día de una cuenta
 * —que abre con valor cero— deja de dividir por cero.
 */
export function periodReturns(points: GrowthDataPoint[]): PeriodReturn[] {
	const series = sortedPoints(points);
	const returns: PeriodReturn[] = [];

	for (let i = 1; i < series.length; i++) {
		const prev = series[i - 1];
		const curr = series[i];

		const prevValue = toNumber(prev.totalValue);
		const currValue = toNumber(curr.totalValue);
		const prevCost = toNumber(prev.totalCostBase);
		const currCost = toNumber(curr.totalCostBase);
		if (![prevValue, currValue, prevCost, currCost].every(Number.isFinite)) continue;

		const flow = currCost - prevCost;
		const base = prevValue + flow / 2;
		if (base <= 0) continue;

		const days = dayGap(prev.date, curr.date);
		if (days <= 0) continue;

		returns.push({ date: curr.date, value: (currValue - prevValue - flow) / base, days });
	}

	return returns;
}

/** Encadena retornos fraccionales: el rendimiento del conjunto, no su suma. */
export function compound(values: number[]): number {
	return values.reduce((acc, value) => acc * (1 + value), 1) - 1;
}

/** Media aritmética; la serie de retornos nunca llega aquí vacía. */
export function mean(values: number[]): number {
	return values.reduce((sum, value) => sum + value, 0) / values.length;
}

/** Desviación típica muestral: la serie es una muestra, no la población entera. */
export function stdDev(values: number[]): number {
	if (values.length < 2) return 0;
	const average = mean(values);
	const variance =
		values.reduce((sum, value) => sum + (value - average) ** 2, 0) / (values.length - 1);
	return Math.sqrt(variance);
}

/**
 * Tramos por año, para anualizar.
 *
 * Sale de la mediana del espaciado y no de la media: la serie es diaria, pero
 * un job caído deja un hueco de varios días que arrastraría la media entera.
 */
export function periodsPerYear(returns: PeriodReturn[]): number {
	const gaps = returns.map((r) => r.days).sort((a, b) => a - b);
	const median = gaps[Math.floor(gaps.length / 2)] || 1;
	return DAYS_PER_YEAR / median;
}

/** Peor caída desde un máximo del índice de rentabilidad, en fracción (≤ 0). */
export function maxDrawdown(values: number[]): number {
	let index = 1;
	let peak = 1;
	let worst = 0;

	for (const value of values) {
		index *= 1 + value;
		if (index > peak) peak = index;
		const drawdown = index / peak - 1;
		if (drawdown < worst) worst = drawdown;
	}

	return worst;
}

/** Retorno encadenado de cada mes con dato, en clave `YYYY-MM`. */
export function returnsByMonth(returns: PeriodReturn[]): Map<string, number> {
	const buckets = new Map<string, number[]>();

	for (const entry of returns) {
		const key = entry.date.substring(0, 7);
		const bucket = buckets.get(key);
		if (bucket) bucket.push(entry.value);
		else buckets.set(key, [entry.value]);
	}

	return new Map([...buckets].map(([key, values]) => [key, compound(values)]));
}
