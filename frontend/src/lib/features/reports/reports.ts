/**
 * Dominio `reports`: los paneles del centro de reportes y las constantes de su
 * presentación.
 *
 * El backend no expone estas vistas; se derivan de la serie de crecimiento
 * agregada (`GET /portfolios/growth`). La aritmética que las alimenta vive en
 * `returns.ts`, que es donde se explica por qué un aporte no puede contar como
 * rentabilidad; aquí solo se arman los bloques y se les da formato.
 */

import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';
import { formatCurrency } from '$lib/shared/format/money';
import {
	DAYS_PER_YEAR,
	compound,
	dayGap,
	maxDrawdown,
	mean,
	periodReturns,
	periodsPerYear,
	returnsByMonth,
	sortedPoints,
	stdDev,
	toNumber
} from './returns';

/** Rentabilidad mes a mes de un año; `null` en los meses sin dato. */
export interface PerformanceCalendar {
	year: string;
	values: (number | null)[];
	/**
	 * Índice del mes que solo cubre parte del calendario porque el historial
	 * empieza dentro de él. Su cifra es real pero no comparable con un mes
	 * entero, y la tarjeta la marca.
	 */
	partialMonth: number | null;
}

/** Métrica del panel de estadísticas. */
export interface KeyStat {
	label: string;
	value: string;
	/** Qué mide; y por qué falta, cuando el valor es `N/A`. */
	hint?: string;
	/** Signo con el que colorear la cifra. `neutral` no la tiñe. */
	tone?: 'up' | 'down' | 'neutral';
}

/** Bloque de métricas afines dentro del panel de estadísticas. */
export interface KeyStatGroup {
	title: string;
	stats: KeyStat[];
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

/** Valor de una métrica que el historial todavía no da para calcular. */
export const UNAVAILABLE = 'N/A';

/** Retornos mínimos —y días mínimos— para que una cifra de riesgo signifique algo. */
const MIN_RISK_RETURNS = 10;
const MIN_RISK_DAYS = 21;

/** Días mínimos antes de anualizar: por debajo de un trimestre el factor dispara. */
const MIN_ANNUALIZED_DAYS = 90;

/** Días mínimos para proyectar a cinco años. */
const MIN_PROJECTION_DAYS = 183;

/** Tramo de color de una celda del calendario, por rentabilidad mensual. */
export function performanceClass(value: number): string {
	if (value >= 2) return 'strong-positive';
	if (value >= 1) return 'positive';
	if (value >= 0) return 'flat-positive';
	if (value > -1) return 'negative';
	return 'strong-negative';
}

// ---------------------------------------------------------------------------
// Formato
// ---------------------------------------------------------------------------

const PERCENT = new Intl.NumberFormat('es-CO', {
	minimumFractionDigits: 1,
	maximumFractionDigits: 1
});

const RATIO = new Intl.NumberFormat('es-CO', {
	minimumFractionDigits: 2,
	maximumFractionDigits: 2
});

/**
 * Porcentaje con coma decimal. El resto de cifras de la aplicación usan `Intl`
 * con la configuración de es-CO; estas se escapaban con un punto.
 */
function formatPercent(value: number): string {
	// Sin esto, una cifra que redondea a cero sale como «-0,0 %».
	return `${PERCENT.format(Math.abs(value) < 0.05 ? 0 : value)}%`;
}

/** Como `formatPercent`, con el `+` explícito que pide una cifra de rendimiento. */
function formatSignedPercent(value: number): string {
	return `${value >= 0.05 ? '+' : ''}${formatPercent(value)}`;
}

/**
 * El importe del resumen, o el del último punto de la serie si aquel no lo
 * trae. Los dos campos son opcionales en el contrato de la API, y un backend
 * anterior los omite.
 */
function amountOr(summaryValue: string | undefined, pointValue: string): number {
	const value = toNumber(summaryValue);
	return Number.isFinite(value) ? value : toNumber(pointValue);
}

/** Tono con el que colorear una cifra con signo. */
function toneOf(value: number): KeyStat['tone'] {
	if (value >= 0.05) return 'up';
	if (value <= -0.05) return 'down';
	return 'neutral';
}

/**
 * Un día del historial, «12 Jul». Se compone a mano en vez de con `Intl`: en
 * es-CO el formato corto es «12 de jul», que en una celda de métrica ocupa el
 * doble y no casa con las abreviaturas del calendario de al lado.
 */
function formatDay(date: string, withYear: boolean): string {
	const [year, month, day] = date.split('-');
	const label = `${Number.parseInt(day, 10)} ${MONTHS[Number.parseInt(month, 10) - 1]}`;
	return withYear ? `${label} ${year}` : label;
}

/** Rango del historial, «12 Jul – 22 Ago 2026»; el año se repite solo si cambia. */
function formatSpan(from: string, to: string): string {
	const sameYear = from.substring(0, 4) === to.substring(0, 4);
	return `${formatDay(from, !sameYear)} – ${formatDay(to, true)}`;
}

// ---------------------------------------------------------------------------
// Paneles
// ---------------------------------------------------------------------------

/**
 * Calendario de rentabilidad por año, del más reciente al más antiguo.
 *
 * Cada mes encadena los retornos de sus tramos, así que un mes sin snapshots no
 * aparece y el primero del historial cubre solo desde el día en que empieza:
 * queda marcado como parcial para que nadie lo compare con un mes entero.
 */
export function buildPerformanceCalendars(points: GrowthDataPoint[]): PerformanceCalendar[] {
	const series = sortedPoints(points);
	const returns = periodReturns(series);
	if (returns.length === 0) return [];

	const byYear = new Map<string, (number | null)[]>();
	for (const [key, value] of returnsByMonth(returns)) {
		const [year, month] = key.split('-');
		if (!byYear.has(year)) byYear.set(year, Array(12).fill(null));
		byYear.get(year)![Number.parseInt(month, 10) - 1] = Number((value * 100).toFixed(2));
	}

	// El primer mes con retorno es parcial solo si el historial empieza dentro de
	// él. Si el primer punto cae en el mes anterior —un cierre de enero seguido
	// de febrero—, febrero está entero y marcarlo sobraría.
	const firstReturnMonth = returns[0].date.substring(0, 7);
	const partialKey = series[0].date.substring(0, 7) === firstReturnMonth ? firstReturnMonth : null;
	const [partialYear, partialMonth] = (partialKey ?? '').split('-');

	return [...byYear.entries()]
		.sort(([a], [b]) => b.localeCompare(a))
		.map(([year, values]) => ({
			year,
			values,
			partialMonth:
				partialKey !== null && year === partialYear ? Number.parseInt(partialMonth, 10) - 1 : null
		}));
}

/**
 * Estadísticas del historial, en tres bloques: lo que rindió, el riesgo que
 * costó y de cuánto historial salen las dos cosas.
 *
 * Una métrica que el historial todavía no sostiene sale como `N/A` con la
 * razón en `hint` —antes solo decía `N/A`, y no había forma de saber si
 * faltaban datos o algo estaba roto—.
 */
export function buildKeyStatistics(
	points: GrowthDataPoint[],
	summary: GrowthSummary
): KeyStatGroup[] {
	const series = sortedPoints(points);
	if (series.length === 0) return [];

	const last = series[series.length - 1];
	const currency = summary.currency || 'USD';
	const returns = periodReturns(series);
	const values = returns.map((r) => r.value);
	const spanDays = dayGap(series[0].date, last.date);

	const totalReturn = compound(values);
	const invested = toNumber(last.totalCostBase);
	// El `||` cuenta el cero como ausencia a propósito: el loader rellena con
	// ceros el resumen que no llegó, y la serie sigue teniendo el valor bueno.
	const currentValue = toNumber(summary.currentValue) || toNumber(last.totalValue);
	const gainLoss = amountOr(summary.gainLoss, last.gainLoss);
	const gainLossPct = amountOr(summary.gainLossPct, last.gainLossPct);

	const hasRisk = returns.length >= MIN_RISK_RETURNS && spanDays >= MIN_RISK_DAYS;
	const missingRisk = `Necesita al menos ${MIN_RISK_RETURNS} puntos y ${MIN_RISK_DAYS} días de historial; llevas ${returns.length} y ${spanDays}.`;

	const perYear = returns.length > 0 ? periodsPerYear(returns) : 0;
	const volatility = hasRisk ? stdDev(values) * Math.sqrt(perYear) : null;
	const sharpe =
		hasRisk && volatility && volatility > 0 ? (mean(values) * perYear) / volatility : null;

	const annualized =
		spanDays >= MIN_ANNUALIZED_DAYS && totalReturn > -1
			? Math.pow(1 + totalReturn, DAYS_PER_YEAR / spanDays) - 1
			: null;

	const monthly = [...returnsByMonth(returns)].sort(([a], [b]) => a.localeCompare(b));
	const best = monthly.reduce<[string, number] | null>(
		(top, entry) => (top === null || entry[1] > top[1] ? entry : top),
		null
	);
	const worst = monthly.reduce<[string, number] | null>(
		(low, entry) => (low === null || entry[1] < low[1] ? entry : low),
		null
	);

	// Con el año: el historial cruza años y «Oct» a secas no dice cuál.
	const monthLabel = (entry: [string, number] | null) => {
		if (entry === null) return UNAVAILABLE;
		const [year, month] = entry[0].split('-');
		return `${formatSignedPercent(entry[1] * 100)} · ${MONTHS[Number.parseInt(month, 10) - 1]} ${year}`;
	};

	const performance: KeyStat[] = [
		{
			label: 'Rentabilidad del periodo',
			value: returns.length > 0 ? formatSignedPercent(totalReturn * 100) : UNAVAILABLE,
			hint: 'Lo que rindió el dinero invertido, encadenando los tramos del historial. Los aportes y retiros no cuentan como rentabilidad.',
			tone: toneOf(totalReturn * 100)
		},
		{
			label: 'Rentabilidad anualizada',
			value: annualized === null ? UNAVAILABLE : formatSignedPercent(annualized * 100),
			hint:
				annualized === null
					? `Se anualiza a partir de ${MIN_ANNUALIZED_DAYS} días de historial; llevas ${spanDays}.`
					: 'La rentabilidad del periodo llevada a un año. No es una previsión.',
			tone: annualized === null ? 'neutral' : toneOf(annualized * 100)
		},
		{
			label: 'Ganancia / pérdida',
			value: Number.isFinite(gainLoss) ? formatCurrency(gainLoss, currency) : UNAVAILABLE,
			hint: 'Valor de mercado menos capital invertido, a día de hoy.',
			tone: Number.isFinite(gainLoss) ? toneOf(gainLoss) : 'neutral'
		},
		{
			label: 'Ganancia sobre coste',
			value: Number.isFinite(gainLossPct) ? formatSignedPercent(gainLossPct) : UNAVAILABLE,
			hint: 'La ganancia anterior, en porcentaje de lo invertido.',
			tone: Number.isFinite(gainLossPct) ? toneOf(gainLossPct) : 'neutral'
		},
		{
			label: 'Mejor mes',
			value: monthLabel(best),
			hint: 'El mes que más rindió del historial.',
			tone: best === null ? 'neutral' : toneOf(best[1] * 100)
		},
		{
			label: 'Peor mes',
			value: monthLabel(worst),
			hint: 'El mes que menos rindió del historial.',
			tone: worst === null ? 'neutral' : toneOf(worst[1] * 100)
		}
	];

	const risk: KeyStat[] = [
		{
			label: 'Volatilidad anualizada',
			value: volatility === null ? UNAVAILABLE : formatPercent(volatility * 100),
			hint:
				volatility === null
					? missingRisk
					: 'Cuánto oscila la rentabilidad, llevado a un año. Más alta es más movimiento, arriba y abajo.',
			tone: 'neutral'
		},
		{
			label: 'Máxima caída',
			value: returns.length > 0 ? formatPercent(maxDrawdown(values) * 100) : UNAVAILABLE,
			hint: 'La peor bajada desde un máximo hasta el siguiente suelo. Mide lo que habrías aguantado, no lo que perdiste.',
			tone: 'down'
		},
		{
			label: 'Ratio de Sharpe',
			value: sharpe === null ? UNAVAILABLE : RATIO.format(sharpe),
			hint:
				sharpe === null
					? missingRisk
					: 'Rentabilidad por unidad de riesgo, con tasa libre de riesgo 0. Por encima de 1 se considera bueno.',
			tone: sharpe === null ? 'neutral' : toneOf(sharpe)
		}
	];

	const history: KeyStat[] = [
		{
			label: 'Periodo cubierto',
			value: formatSpan(series[0].date, last.date),
			hint: `${series.length} ${series.length === 1 ? 'punto' : 'puntos'} de historial, uno por día con datos.`,
			tone: 'neutral'
		},
		{
			label: 'Capital invertido',
			value: Number.isFinite(invested) ? formatCurrency(invested, currency) : UNAVAILABLE,
			hint: 'Lo que costaron las posiciones que tienes abiertas.',
			tone: 'neutral'
		},
		{
			label: 'Valor actual',
			value: Number.isFinite(currentValue) ? formatCurrency(currentValue, currency) : UNAVAILABLE,
			hint: `Valor de mercado de la cuenta, en ${currency}.`,
			tone: 'neutral'
		}
	];

	return [
		{ title: 'Rendimiento', stats: performance },
		{ title: 'Riesgo', stats: risk },
		{ title: 'Historial', stats: history }
	];
}

/** Días que cubre el historial, para explicar qué falta cuando no hay proyección. */
export function historySpanDays(points: GrowthDataPoint[]): number {
	const series = sortedPoints(points);
	if (series.length < 2) return 0;
	return dayGap(series[0].date, series[series.length - 1].date);
}

/** Días de historial que exige la proyección, para el texto del estado vacío. */
export const PROJECTION_MIN_DAYS = MIN_PROJECTION_DAYS;

/**
 * Proyección a cinco años extrapolando la rentabilidad anualizada del historial.
 *
 * Extrapola rendimiento, no crecimiento del saldo: proyectar con la variación
 * del valor daba cifras absurdas en cuanto la cuenta recibía un aporte, porque
 * el depósito entraba en la tasa. Se abstiene en cuanto el dato no da para una
 * proyección honesta: menos de medio año de historial, valores no positivos o
 * una tasa fuera de un rango plausible.
 */
export function buildGrowthProjection(
	points: GrowthDataPoint[],
	summary: GrowthSummary
): GrowthProjectionEntry[] {
	const series = sortedPoints(points);
	if (series.length < 2) return [];

	const last = series[series.length - 1];
	const currentValue = toNumber(summary.currentValue) || toNumber(last.totalValue);
	if (!Number.isFinite(currentValue) || currentValue <= 0) return [];

	const spanDays = dayGap(series[0].date, last.date);
	if (spanDays < MIN_PROJECTION_DAYS) return [];

	const totalReturn = compound(periodReturns(series).map((r) => r.value));
	if (totalReturn <= -1) return [];

	const rate = Math.pow(1 + totalReturn, DAYS_PER_YEAR / spanDays) - 1;
	if (!Number.isFinite(rate) || rate < -0.5 || rate > 2.0) return [];

	const startYear = Number.parseInt(last.date.substring(0, 4), 10);
	return Array.from({ length: 5 }, (_, i) => ({
		period: String(startYear + i),
		value: Math.round(currentValue * Math.pow(1 + rate, i))
	}));
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
