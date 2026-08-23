/**
 * Proyección de crecimiento del centro de reportes.
 *
 * Vive aparte de `reports.ts` —que arma el calendario y las estadísticas— por
 * tamaño: los dos bloques no comparten nada más que la serie de la que salen y
 * el módulo común pasaba del presupuesto de 500 líneas.
 *
 * La aritmética es la de `$lib/shared/finance/returns`: se extrapola
 * rentabilidad, nunca la variación del saldo.
 */

import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';
import {
	annualize,
	sortedPoints,
	spanDays,
	timeWeightedReturn,
	toNumber
} from '$lib/shared/finance/returns';

/** Punto de la proyección a cinco años. */
export interface GrowthProjectionEntry {
	period: string;
	/** Valor proyectado de la cartera ese año, en la moneda del resumen. */
	value: number;
	/**
	 * Lo que se habría acumulado desde hoy hasta ese año, en porcentaje.
	 *
	 * El dinero proyectado depende de cuánto haya hoy en la cuenta, así que dos
	 * inversionistas con la misma tasa ven cifras que no se parecen; el
	 * porcentaje es el mismo para los dos y es lo que de verdad extrapola la
	 * proyección. El primer año es el punto de partida y vale 0.
	 */
	returnPct: number;
}

/**
 * La proyección completa: los cinco puntos y la tasa de la que salen.
 *
 * Se llama `…Series` y no `GrowthProjection` a secas porque ese nombre ya es el
 * del componente que la pinta, y la superficie pública de la feature exporta
 * los dos.
 */
export interface GrowthProjectionSeries {
	/** Tasa anualizada extrapolada del historial, en porcentaje. */
	annualRatePct: number;
	entries: GrowthProjectionEntry[];
}

/** Días que cubre el historial, para explicar qué falta cuando no hay proyección. */
export function historySpanDays(points: GrowthDataPoint[]): number {
	return spanDays(points);
}

/**
 * Días mínimos de historial para proyectar a cinco años.
 *
 * Medio año. El panel lo publica para poder decir cuánto falta en su estado
 * vacío, que ahorra volver cada semana a comprobarlo.
 */
export const PROJECTION_MIN_DAYS = 183;

/**
 * Proyección a cinco años extrapolando la rentabilidad anualizada del historial.
 *
 * Extrapola rendimiento, no crecimiento del saldo: proyectar con la variación
 * del valor daba cifras absurdas en cuanto la cuenta recibía un aporte, porque
 * el depósito entraba en la tasa. Se abstiene en cuanto el dato no da para una
 * proyección honesta: menos de medio año de historial, valores no positivos o
 * una tasa fuera de un rango plausible; entonces devuelve `null` y el panel
 * dice cuánto historial falta.
 *
 * Cada punto sale por partida doble —dinero y porcentaje acumulado— y la
 * proyección lleva la tasa de la que salen los dos. El importe solo se entiende
 * sabiendo cuánto hay hoy en la cuenta; el porcentaje es la extrapolación en
 * crudo, y es la cifra que se compara con la de cualquier otra cartera.
 */
export function buildGrowthProjection(
	points: GrowthDataPoint[],
	summary: GrowthSummary
): GrowthProjectionSeries | null {
	const series = sortedPoints(points);
	if (series.length < 2) return null;

	const last = series[series.length - 1];
	const currentValue = toNumber(summary.currentValue) || toNumber(last.totalValue);
	if (!Number.isFinite(currentValue) || currentValue <= 0) return null;

	const historyDays = spanDays(series);
	if (historyDays < PROJECTION_MIN_DAYS) return null;

	const totalReturn = timeWeightedReturn(series);
	if (totalReturn === null) return null;

	const rate = annualize(totalReturn, historyDays);
	if (rate === null || rate < -0.5 || rate > 2.0) return null;

	const startYear = Number.parseInt(last.date.substring(0, 4), 10);

	return {
		annualRatePct: rate * 100,
		entries: Array.from({ length: 5 }, (_, i) => ({
			period: String(startYear + i),
			value: Math.round(currentValue * Math.pow(1 + rate, i)),
			returnPct: (Math.pow(1 + rate, i) - 1) * 100
		}))
	};
}

/**
 * Canal izquierdo de la gráfica de proyección, en unidades del viewBox.
 *
 * Son 58 y no 40: con una tasa pequeña las marcas del eje llegan a «$89.41k», y
 * con el margen anterior el `$` y las primeras cifras se salían del viewBox y
 * se veían cortadas.
 */
export const PROJECTION_GUTTER = 58;

/** Coordenadas de la gráfica de proyección dentro del viewBox `0 0 600 280`. */
export function projectionCoordinates(
	entries: GrowthProjectionEntry[]
): { x: number; y: number; period: string }[] {
	if (entries.length === 0) return [];

	const values = entries.map((p) => p.value);
	const min = Math.min(...values);
	const range = Math.max(...values) - min || 1;

	return entries.map((point, i) => ({
		x: PROJECTION_GUTTER + i * 124,
		y: 230 - ((point.value - min) / range) * 180,
		period: point.period
	}));
}
