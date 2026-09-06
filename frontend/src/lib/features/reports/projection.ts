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
 * Geometría de la curva de proyección dentro del viewBox `0 0 600 208`.
 *
 * Lo que se dibuja es el porcentaje acumulado, no el dinero, y el eje incluye
 * siempre el cero. La gráfica anterior estiraba el rango de los cinco importes
 * al alto del lienzo: con una tasa del −0,3 % anual, noventa dólares de
 * diferencia ocupaban todo el canvas y la proyección se veía como un
 * desplome. Anclada en el cero, una tasa que no mueve nada dibuja una línea
 * que no se mueve, que es lo que de verdad está pasando.
 *
 * Los importes no viajan aquí: van en la tabla de debajo, que es donde una
 * cifra se lee.
 */
export const PROJECTION_GUTTER = 54;

const PLOT = { right: 570, top: 22, bottom: 160 };

/** Marcas del eje, la línea del cero y los trazos de la curva. */
export interface ProjectionGeometry {
	points: { x: number; y: number; period: string }[];
	/** `points` de la polilínea. */
	line: string;
	/** El mismo trazo cerrado contra la línea del cero, para el relleno. */
	area: string;
	zeroY: number;
	/** Marcas del eje vertical; la del cero no está, la dibuja su propia línea. */
	ticks: { y: number; value: number }[];
}

const EMPTY: ProjectionGeometry = { points: [], line: '', area: '', zeroY: PLOT.bottom, ticks: [] };

export function projectionGeometry(entries: GrowthProjectionEntry[]): ProjectionGeometry {
	if (entries.length === 0) return EMPTY;

	const values = entries.map((entry) => entry.returnPct);
	// El cero entra siempre en el rango, por arriba o por abajo: es contra él
	// contra lo que se lee si la curva sube o baja.
	const lo = Math.min(0, ...values);
	const top = Math.max(0, ...values);
	// Con la tasa clavada en cero el rango es nulo: un punto de margen deja la
	// línea plana en su sitio en vez de dividir entre cero.
	const hi = top - lo < 1e-9 ? lo + 1 : top;
	const y = (value: number) => PLOT.bottom - ((value - lo) / (hi - lo)) * (PLOT.bottom - PLOT.top);

	const step = entries.length > 1 ? (PLOT.right - PROJECTION_GUTTER) / (entries.length - 1) : 0;

	const points = entries.map((entry, i) => ({
		x: PROJECTION_GUTTER + i * step,
		y: y(entry.returnPct),
		period: entry.period
	}));

	const line = points.map((point) => `${point.x},${point.y}`).join(' ');
	const zeroY = y(0);
	const last = points[points.length - 1];

	return {
		points,
		line,
		area: `${line} ${last.x},${zeroY} ${points[0].x},${zeroY}`,
		zeroY,
		// Cuatro marcas repartidas por el rango, sin la que caería encima de la
		// línea del cero y repetiría su etiqueta.
		ticks: Array.from({ length: 4 }, (_, i) => {
			const value = hi - ((hi - lo) * i) / 3;
			return { y: y(value), value };
		}).filter((tick) => Math.abs(tick.y - zeroY) > 6)
	};
}
