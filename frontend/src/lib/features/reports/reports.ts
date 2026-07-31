/**
 * Dominio `reports`: los cálculos que alimentan el centro de reportes y las
 * constantes de su presentación.
 *
 * El backend no expone estas vistas; se derivan de la serie de crecimiento
 * agregada (`GET /portfolios/growth`). Vivían dentro de
 * `routes/dashboard/reports/+page.server.ts`, sin ninguna prueba: aquí son
 * funciones puras con su `reports.spec.ts`.
 */

import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

/** Rentabilidad mes a mes de un año; `null` en los meses sin dato previo. */
export interface PerformanceCalendar {
	year: string;
	values: (number | null)[];
}

/** Métrica del panel de estadísticas. */
export interface KeyStat {
	label: string;
	value: string;
}

/** Punto de la proyección a cinco años. */
export interface GrowthProjectionEntry {
	period: string;
	value: number;
}

export const MONTHS = [
	'Ene',
	'Feb',
	'Mar',
	'Abr',
	'May',
	'Jun',
	'Jul',
	'Ago',
	'Sep',
	'Oct',
	'Nov',
	'Dic'
];

/** Reportes descargables, todos servidos por `reports/download?type=`. */
export const REPORT_DOWNLOADS = [
	{ title: 'Resumen mensual', format: 'XLSX', type: 'summary' },
	{ title: 'Estado de resultados', format: 'XLSX', type: 'transactions' },
	{ title: 'Riesgo y volatilidad', format: 'XLSX', type: 'risk' }
];

/** Tramo de color de una celda del calendario, por rentabilidad mensual. */
export function performanceClass(value: number): string {
	if (value >= 2) return 'strong-positive';
	if (value >= 1) return 'positive';
	if (value >= 0) return 'flat-positive';
	if (value > -1) return 'negative';
	return 'strong-negative';
}

/** Último punto de cada mes de la serie, ordenados por `YYYY-MM`. */
function monthlyPoints(points: GrowthDataPoint[]): [string, GrowthDataPoint][] {
	const byYearMonth = new Map<string, GrowthDataPoint>();
	for (const point of points) {
		byYearMonth.set(point.date.substring(0, 7), point);
	}
	return [...byYearMonth.entries()].sort(([a], [b]) => a.localeCompare(b));
}

/**
 * Calendario de rentabilidad por año, del más reciente al más antiguo.
 *
 * El primer mes de la serie no tiene con qué compararse y queda en `null`, lo
 * mismo que un mes cuyo valor previo fuera cero (la variación no está
 * definida).
 */
export function buildPerformanceCalendars(points: GrowthDataPoint[]): PerformanceCalendar[] {
	if (points.length === 0) return [];

	const monthEntries = monthlyPoints(points);
	const byYear = new Map<string, (number | null)[]>();

	for (let i = 0; i < monthEntries.length; i++) {
		const [key] = monthEntries[i];
		const [year, monthStr] = key.split('-');
		const monthIndex = parseInt(monthStr, 10) - 1;

		if (!byYear.has(year)) byYear.set(year, Array(12).fill(null));

		if (i === 0) {
			byYear.get(year)![monthIndex] = null;
		} else {
			const prevVal = parseFloat(monthEntries[i - 1][1].totalValue);
			const currVal = parseFloat(monthEntries[i][1].totalValue);
			byYear.get(year)![monthIndex] =
				prevVal > 0 ? parseFloat(((currVal / prevVal - 1) * 100).toFixed(2)) : null;
		}
	}

	return [...byYear.entries()]
		.sort(([a], [b]) => b.localeCompare(a))
		.map(([year, values]) => ({ year, values }));
}

/**
 * Máxima caída desde un pico y volatilidad anualizada.
 *
 * La volatilidad necesita al menos tres retornos mensuales para significar
 * algo; con menos se muestra `N/A` en vez de un número engañoso.
 */
export function buildKeyStatistics(points: GrowthDataPoint[]): KeyStat[] {
	if (points.length === 0) return [];

	let peak = -Infinity;
	let maxDrawdown = 0;
	for (const p of points) {
		const v = parseFloat(p.totalValue);
		if (v > peak) peak = v;
		if (peak > 0) {
			const dd = ((v - peak) / peak) * 100;
			if (dd < maxDrawdown) maxDrawdown = dd;
		}
	}

	const monthEntries = monthlyPoints(points);
	const returns: number[] = [];
	for (let i = 1; i < monthEntries.length; i++) {
		const prev = parseFloat(monthEntries[i - 1][1].totalValue);
		const curr = parseFloat(monthEntries[i][1].totalValue);
		if (prev > 0) returns.push((curr / prev - 1) * 100);
	}

	let volatilityStr = 'N/A';
	if (returns.length >= 3) {
		const mean = returns.reduce((s, r) => s + r, 0) / returns.length;
		const variance = returns.reduce((s, r) => s + (r - mean) ** 2, 0) / returns.length;
		volatilityStr = `${(Math.sqrt(variance) * Math.sqrt(12)).toFixed(1)}%`;
	}

	return [
		{ label: 'Max Drawdown', value: `${maxDrawdown.toFixed(1)}%` },
		{ label: 'Volatilidad', value: volatilityStr }
	];
}

/**
 * Proyección a cinco años extrapolando el CAGR del historial.
 *
 * Se abstiene en cuanto el dato no da para una proyección honesta: menos de
 * medio año de historial, valores no positivos o un CAGR fuera de un rango
 * plausible (una racha corta y extrema proyectaría cifras absurdas).
 */
export function buildGrowthProjection(
	points: GrowthDataPoint[],
	summary: GrowthSummary
): GrowthProjectionEntry[] {
	const initialValue = parseFloat(summary.initialValue);
	const currentValue = parseFloat(summary.currentValue);

	if (!points.length || initialValue <= 0 || currentValue <= 0) return [];

	const firstDate = new Date(summary.firstDate + 'T00:00:00');
	const lastDate = new Date(points[points.length - 1].date + 'T00:00:00');
	const years = (lastDate.getTime() - firstDate.getTime()) / (365.25 * 86_400_000);

	if (years < 0.5) return [];

	const cagr = Math.pow(currentValue / initialValue, 1 / years) - 1;
	if (!isFinite(cagr) || cagr < -0.5 || cagr > 2.0) return [];

	const startYear = lastDate.getFullYear();
	return Array.from({ length: 5 }, (_, i) => ({
		period: String(startYear + i),
		value: Math.round(currentValue * Math.pow(1 + cagr, i))
	}));
}

/** Coordenadas de la gráfica de proyección dentro del viewBox `0 0 600 280`. */
export function projectionCoordinates(
	entries: GrowthProjectionEntry[]
): { x: number; y: number; period: string }[] {
	if (entries.length === 0) return [];

	const values = entries.map((p) => p.value);
	const min = Math.min(...values);
	const range = Math.max(...values) - min || 1;

	return entries.map((point, i) => ({
		x: 40 + i * 130,
		y: 230 - ((point.value - min) / range) * 180,
		period: point.period
	}));
}
