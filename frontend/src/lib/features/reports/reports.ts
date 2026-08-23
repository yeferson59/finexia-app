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
	incompleteMonths,
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
	 * Índices de los meses que solo cubren parte del calendario: aquel en el que
	 * empieza el historial y el que está en curso. Su cifra es real pero no
	 * comparable con un mes entero, y la tarjeta la marca.
	 */
	partialMonths: number[];
}

/** Métrica del panel de estadísticas. */
export interface KeyStat {
	label: string;
	value: string;
	/** Qué mide; y por qué falta, cuando el valor es `N/A`. */
	hint?: string;
	/**
	 * Reparo que la cifra arrastra y que hay que leer con ella, no al pasar el
	 * ratón: un Sharpe estimado con pocos meses, un mes incompleto. Se pinta
	 * bajo el valor.
	 */
	note?: string;
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

/** Tramos mínimos para que una cifra de riesgo mida algo y no ruido. */
const MIN_RISK_RETURNS = 10;

/**
 * Días mínimos antes de anualizar: por debajo de un trimestre el factor dispara.
 *
 * Vale para las tres cifras anuales de la tarjeta —rentabilidad anualizada,
 * volatilidad anualizada y Sharpe—, no solo para la primera. Antes la
 * rentabilidad se ocultaba a los 55 días de historial mientras la volatilidad y
 * el Sharpe, que son el mismo número anualizado por otro camino, se publicaban
 * desde los 21: la tarjeta escondía una cifra y enseñaba dos derivadas suyas.
 *
 * Lo que el umbral gobierna es el paso a un año, no la medición. La dispersión
 * de los tramos se publica en cuanto hay tramos que medir, sin anualizar y
 * dicho en la etiqueta: converge mucho antes que una media, y esconderla tres
 * meses era tirar un dato bueno por un factor que sí es prematuro.
 */
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
 * aparece. Los que el historial no cubre enteros —aquel en el que empieza y el
 * que está en curso— quedan marcados como parciales para que nadie los compare
 * con un mes completo; son los mismos que `buildKeyStatistics` deja fuera del
 * mejor y el peor mes.
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

	// Un mes sin retorno propio no se marca aunque esté incompleto: el punto que
	// abre el historial un 31 de enero deja enero vacío, y una celda sin cifra no
	// tiene nada que advertir.
	const incomplete = incompleteMonths(series);

	return [...byYear.entries()]
		.sort(([a], [b]) => b.localeCompare(a))
		.map(([year, values]) => ({
			year,
			values,
			partialMonths: [...incomplete]
				.filter((key) => key.startsWith(`${year}-`))
				.map((key) => Number.parseInt(key.substring(5), 10) - 1)
				.filter((index) => values[index] !== null)
				.sort((a, b) => a - b)
		}));
}

/**
 * Estadísticas del historial, en tres bloques: lo que rindió, el riesgo que
 * costó y de cuánto historial salen las dos cosas.
 *
 * Una métrica que el historial todavía no sostiene sale como `N/A` con la
 * razón en `hint` —antes solo decía `N/A`, y no había forma de saber si
 * faltaban datos o algo estaba roto—.
 *
 * Dos convenciones que la tarjeta tiene que dejar dichas, porque las cifras se
 * contradicen a la vista de quien no las conoce:
 *
 *   - La rentabilidad del periodo es ponderada por tiempo: encadena tramos e
 *     ignora cuándo entró cada aporte. La ganancia sí depende de eso, así que
 *     un +30 % de periodo puede convivir con un +10 % sobre coste cuando el
 *     dinero grande entró después de la subida. Ninguna de las dos está mal;
 *     miden cosas distintas y cada `hint` lo dice.
 *   - El Sharpe se anualiza por la vía de siempre —media aritmética de los
 *     tramos × √tramos por año ÷ volatilidad—, que no es la rentabilidad
 *     compuesta de arriba dividida entre la volatilidad. Multiplicar el Sharpe
 *     por la volatilidad da una anualizada aritmética, más baja que la
 *     compuesta; el `hint` avisa para que nadie cuadre una con la otra.
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

	// Un solo umbral de historial para las cifras anuales, y otro —de tramos—
	// para lo que solo pide dispersión: 90 días de una serie con cuatro puntos no
	// miden oscilación ninguna, y 60 días de una serie diaria sí.
	const hasAnnualSpan = spanDays >= MIN_ANNUALIZED_DAYS;
	const missingDispersion = `Necesita al menos ${MIN_RISK_RETURNS} tramos de historial; llevas ${returns.length}.`;
	const missingRisk = `Necesita al menos ${MIN_RISK_RETURNS} tramos y ${MIN_ANNUALIZED_DAYS} días de historial; llevas ${returns.length} y ${spanDays}.`;

	const perYear = returns.length > 0 ? periodsPerYear(returns) : 0;
	// La desviación de los tramos es la medida; el √tramos es el que espera al
	// trimestre. Por debajo se publica la primera, etiquetada como lo que es.
	const dispersion = returns.length >= MIN_RISK_RETURNS ? stdDev(values) : null;
	const annualVolatility =
		dispersion === null || !hasAnnualSpan ? null : dispersion * Math.sqrt(perYear);
	const volatility = annualVolatility ?? dispersion;
	const sharpe =
		annualVolatility !== null && annualVolatility > 0
			? (mean(values) * perYear) / annualVolatility
			: null;

	const annualized =
		hasAnnualSpan && totalReturn > -1
			? Math.pow(1 + totalReturn, DAYS_PER_YEAR / spanDays) - 1
			: null;

	// El mejor y el peor mes se buscan solo entre los meses enteros: el primero
	// del historial y el que está en curso cubren unos pocos días, y un +0,4 %
	// de tres días entraba como «peor mes» delante de meses completos peores.
	// Sin ningún mes entero todavía se compara lo que hay, marcado con el mismo
	// asterisco que usa el calendario.
	const incomplete = incompleteMonths(series);
	const monthly = [...returnsByMonth(returns)].sort(([a], [b]) => a.localeCompare(b));
	const whole = monthly.filter(([month]) => !incomplete.has(month));
	const comparable = whole.length > 0 ? whole : monthly;
	const onlyPartial = whole.length === 0 && monthly.length > 0;

	const best = comparable.reduce<[string, number] | null>(
		(top, entry) => (top === null || entry[1] > top[1] ? entry : top),
		null
	);
	const worst = comparable.reduce<[string, number] | null>(
		(low, entry) => (low === null || entry[1] < low[1] ? entry : low),
		null
	);

	// Con el año: el historial cruza años y «Oct» a secas no dice cuál.
	const monthLabel = (entry: [string, number] | null) => {
		if (entry === null) return UNAVAILABLE;
		const [year, month] = entry[0].split('-');
		const label = `${MONTHS[Number.parseInt(month, 10) - 1]} ${year}${onlyPartial ? '*' : ''}`;
		return `${formatSignedPercent(entry[1] * 100)} · ${label}`;
	};

	const monthHint = (entry: [string, number] | null, which: string) => {
		if (entry === null)
			return 'Ningún mes del historial tiene todavía un tramo con el que calcularlo.';

		return onlyPartial
			? `El mes que ${which} rindió. * Está incompleto: su cifra cubre unos pocos días y no se compara con un mes entero.`
			: `El mes que ${which} rindió del historial. Los meses incompletos —aquel en el que empieza y el que está en curso— no compiten.`;
	};

	const monthNote = onlyPartial ? '* Mes incompleto.' : undefined;

	const performance: KeyStat[] = [
		{
			label: 'Rentabilidad del periodo',
			value: returns.length > 0 ? formatSignedPercent(totalReturn * 100) : UNAVAILABLE,
			hint: 'Lo que rindió el dinero invertido, encadenando los tramos del historial. Los aportes y retiros no cuentan como rentabilidad, y tampoco cuenta cuándo entraron: por eso esta cifra no se traduce en la ganancia de abajo.',
			tone: toneOf(totalReturn * 100)
		},
		{
			label: 'Rentabilidad anualizada',
			value: annualized === null ? UNAVAILABLE : formatSignedPercent(annualized * 100),
			hint:
				annualized === null
					? hasAnnualSpan
						? 'Una pérdida total no se anualiza.'
						: `Se anualiza a partir de ${MIN_ANNUALIZED_DAYS} días de historial; llevas ${spanDays}. El mismo umbral vale para la volatilidad y el Sharpe.`
					: 'La rentabilidad del periodo compuesta hasta un año. No es una previsión, y no es la que divide el ratio de Sharpe.',
			tone: annualized === null ? 'neutral' : toneOf(annualized * 100)
		},
		{
			label: 'Ganancia / pérdida',
			value: Number.isFinite(gainLoss) ? formatCurrency(gainLoss, currency) : UNAVAILABLE,
			hint: 'Valor de mercado menos capital invertido, a día de hoy. Depende de cuándo entró cada aporte, así que no es la rentabilidad del periodo aplicada al saldo.',
			tone: Number.isFinite(gainLoss) ? toneOf(gainLoss) : 'neutral'
		},
		{
			label: 'Ganancia sobre coste',
			value: Number.isFinite(gainLossPct) ? formatSignedPercent(gainLossPct) : UNAVAILABLE,
			hint: 'La ganancia anterior, en porcentaje de lo invertido. Queda por debajo de la rentabilidad del periodo cuando los aportes grandes entran después de una subida, y por encima cuando entran antes.',
			tone: Number.isFinite(gainLossPct) ? toneOf(gainLossPct) : 'neutral'
		},
		{
			label: 'Mejor mes',
			value: monthLabel(best),
			hint: monthHint(best, 'más'),
			note: best === null ? undefined : monthNote,
			tone: best === null ? 'neutral' : toneOf(best[1] * 100)
		},
		{
			label: 'Peor mes',
			value: monthLabel(worst),
			hint: monthHint(worst, 'menos'),
			note: worst === null ? undefined : monthNote,
			tone: worst === null ? 'neutral' : toneOf(worst[1] * 100)
		}
	];

	const risk: KeyStat[] = [
		{
			// La etiqueta cambia con la cifra: una volatilidad de tramo y una anual
			// se diferencian en un factor de veinte, y llamarlas igual sería peor
			// que no publicar la primera.
			label: annualVolatility === null ? 'Volatilidad por tramo' : 'Volatilidad anualizada',
			value: volatility === null ? UNAVAILABLE : formatPercent(volatility * 100),
			hint:
				volatility === null
					? missingDispersion
					: annualVolatility === null
						? `Cuánto oscila la rentabilidad de un tramo del historial al siguiente, normalmente de un día al otro. Sin llevar a un año: eso pide ${MIN_ANNUALIZED_DAYS} días de historial.`
						: 'Cuánto oscila la rentabilidad, llevada a un año por la raíz de los tramos. Más alta es más movimiento, arriba y abajo.',
			note:
				volatility !== null && annualVolatility === null
					? `Sin anualizar: se lleva a un año a partir de ${MIN_ANNUALIZED_DAYS} días.`
					: undefined,
			tone: 'neutral'
		},
		{
			label: 'Máxima caída',
			value: returns.length > 0 ? formatPercent(maxDrawdown(values) * 100) : UNAVAILABLE,
			hint: 'La peor bajada desde un máximo hasta el siguiente suelo, medida tramo a tramo: puede caer dentro de un mes que cerró en positivo. Mide lo que habrías aguantado, no lo que perdiste.',
			tone: 'down'
		},
		{
			label: 'Ratio de Sharpe',
			value: sharpe === null ? UNAVAILABLE : RATIO.format(sharpe),
			hint:
				sharpe === null
					? missingRisk
					: 'Rentabilidad por unidad de riesgo, con tasa libre de riesgo 0. Sale de la media aritmética de los tramos anualizada, no de la rentabilidad anualizada de arriba: multiplicarlo por la volatilidad no devuelve aquella cifra.',
			// En gris siempre: es un cociente estimado, no una ganancia, y pintarlo en
			// verde lo vendía como un sello de calidad. Con pocos meses el margen de
			// error se come la diferencia entre un 1 y un 3, y la nota lo dice.
			note:
				sharpe === null
					? undefined
					: 'Estimación: con pocos meses de historial su margen de error es amplio.',
			tone: 'neutral'
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
