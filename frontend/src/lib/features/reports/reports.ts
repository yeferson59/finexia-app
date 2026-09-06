/**
 * Dominio `reports`: la ficha de resultados de la cuenta.
 *
 * El backend no expone estas vistas; se derivan de la serie de crecimiento
 * agregada (`GET /portfolios/growth`). La aritmética que las alimenta vive en
 * `$lib/shared/finance/returns`, que es donde se explica por qué un aporte no
 * puede contar como rentabilidad; aquí solo se arman los bloques y se les da
 * formato.
 *
 * Tres bloques: la cifra de cabecera (`buildRecordSummary`), la matriz de
 * rentabilidad mes a mes (`buildPerformanceCalendars`) y las notas de
 * movimiento y riesgo (`buildKeyStatistics`). La proyección a cinco años va
 * aparte, en `projection.ts`, por tamaño.
 */

import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';
import { formatPercent, formatSignedPercent } from '$lib/shared/format/percent';
import {
	annualize,
	compound,
	incompleteMonths,
	maxDrawdown,
	mean,
	periodReturns,
	periodsPerYear,
	returnsByMonth,
	sortedPoints,
	spanDays,
	stdDev,
	toNumber
} from '$lib/shared/finance/returns';

/** Rentabilidad mes a mes de un año; `null` en los meses sin dato. */
export interface PerformanceCalendar {
	year: string;
	values: (number | null)[];
	/** Lo que rindió el año entero, encadenando los meses con dato. */
	total: number | null;
	/**
	 * Índices de los meses que solo cubren parte del calendario: aquel en el que
	 * empieza el historial y el que está en curso. Su cifra es real pero no
	 * comparable con un mes entero, y la tabla los marca.
	 */
	partialMonths: number[];
}

/** Una fila de las notas de movimiento y riesgo. */
export interface KeyStat {
	label: string;
	value: string;
	/** Dónde cayó la cifra: el mes al que pertenece un máximo o un mínimo. */
	detail?: string;
	/** Qué mide; y qué historial le falta, cuando el valor es `N/A`. */
	hint: string;
	/**
	 * Reparo que la cifra arrastra y que hay que leer con ella: un Sharpe
	 * estimado con pocos meses, una volatilidad sin anualizar.
	 */
	note?: string;
	/** Signo con el que colorear la cifra. `neutral` no la tiñe. */
	tone?: 'up' | 'down' | 'neutral';
}

/**
 * La cabecera de la ficha: qué rindió la cuenta, en cuánto tiempo y sobre
 * cuánto dinero.
 *
 * Sale en números, no en cadenas con formato: los importes los escribe el
 * componente, que es quien sabe si el modo oculto está puesto.
 */
export interface RecordSummary {
	/** Rentabilidad ponderada por tiempo del historial, en %; `null` sin dos cierres. */
	periodReturn: number | null;
	/** La anterior llevada a un año; `null` por debajo del trimestre. */
	annualized: number | null;
	/** Primer y último día con dato, en ISO. */
	from: string;
	to: string;
	/** Meses que cubre el historial, redondeados. */
	months: number;
	value: number;
	cost: number;
	gain: number;
	/** La ganancia anterior en porcentaje de lo invertido. */
	gainPct: number;
	currency: string;
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

const MONTHS_LONG = [
	'enero',
	'febrero',
	'marzo',
	'abril',
	'mayo',
	'junio',
	'julio',
	'agosto',
	'septiembre',
	'octubre',
	'noviembre',
	'diciembre'
];

/**
 * Reportes descargables, todos servidos por `reports/download?type=`.
 *
 * Cada uno dice qué trae dentro: «Estado de resultados» era el historial de
 * transacciones con otro nombre, y quien lo descargaba esperando un estado
 * contable se encontraba con sus movimientos.
 */
export const REPORT_DOWNLOADS = [
	{
		title: 'Resumen mensual',
		description: 'Valor, capital invertido y ganancia de tu cuenta, mes a mes.',
		format: 'XLSX',
		type: 'summary'
	},
	{
		title: 'Transacciones',
		description: 'Cada compra, venta y dividendo que has registrado, con sus comisiones.',
		format: 'XLSX',
		type: 'transactions'
	},
	{
		title: 'Riesgo y volatilidad',
		description:
			'Las medidas de esta página, el detalle mensual y la serie diaria de la que salen.',
		format: 'XLSX',
		type: 'risk'
	}
];

/** Valor de una métrica que el historial todavía no da para calcular. */
export const UNAVAILABLE = 'N/A';

/** Tramos mínimos para que una cifra de riesgo mida algo y no ruido. */
const MIN_RISK_RETURNS = 10;

/**
 * Días mínimos antes de anualizar: por debajo de un trimestre el factor dispara.
 *
 * Vale para las tres cifras anuales de la ficha —rentabilidad anualizada,
 * volatilidad anualizada y Sharpe—, no solo para la primera. Antes la
 * rentabilidad se ocultaba a los 55 días de historial mientras la volatilidad y
 * el Sharpe, que son el mismo número anualizado por otro camino, se publicaban
 * desde los 21: la página escondía una cifra y enseñaba dos derivadas suyas.
 *
 * Lo que el umbral gobierna es el paso a un año, no la medición. La dispersión
 * de los tramos se publica en cuanto hay tramos que medir, sin anualizar y
 * dicho en la etiqueta: converge mucho antes que una media, y esconderla tres
 * meses era tirar un dato bueno por un factor que sí es prematuro.
 */
const MIN_ANNUALIZED_DAYS = 90;

/** Punto porcentual a partir del cual la celda llega a su tinte más intenso. */
const FULL_TINT_AT = 2.5;

/**
 * Fondo de una celda de la matriz, por signo y tamaño del movimiento.
 *
 * Es un degradado continuo y no cinco tramos con su leyenda: la celda imprime
 * su propio porcentaje, así que el color no tiene que ser legible como dato
 * —para eso está la cifra— sino dejar ver de un vistazo dónde se concentró el
 * movimiento. Los cinco escalones de antes pedían aprenderse una clave, y la
 * clave iba repetida debajo de cada año.
 */
export function returnBackground(value: number | null): string {
	if (value === null || !Number.isFinite(value)) return '';

	const intensity = Math.min(Math.abs(value) / FULL_TINT_AT, 1);
	const alpha = (0.05 + intensity * 0.2).toFixed(3);

	return value < 0 ? `rgba(224, 90, 90, ${alpha})` : `rgba(34, 201, 126, ${alpha})`;
}

// ---------------------------------------------------------------------------
// Formato
// ---------------------------------------------------------------------------

const RATIO = new Intl.NumberFormat('es-CO', {
	minimumFractionDigits: 2,
	maximumFractionDigits: 2
});

/** Un día del historial en prosa, «1 de junio de 2025». */
export function formatLongDate(date: string): string {
	const [year, month, day] = date.split('-');
	return `${Number.parseInt(day, 10)} de ${MONTHS_LONG[Number.parseInt(month, 10) - 1]} de ${year}`;
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

/** Suma compuesta de los meses con dato: lo que rindió el año. */
function yearReturn(values: (number | null)[]): number | null {
	const months = values.filter((v): v is number => v !== null);
	if (months.length === 0) return null;
	return (months.reduce((acc, v) => acc * (1 + v / 100), 1) - 1) * 100;
}

// ---------------------------------------------------------------------------
// Bloques
// ---------------------------------------------------------------------------

/**
 * Rentabilidad por mes y por año, del más reciente al más antiguo.
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
			total: yearReturn(values),
			partialMonths: [...incomplete]
				.filter((key) => key.startsWith(`${year}-`))
				.map((key) => Number.parseInt(key.substring(5), 10) - 1)
				.filter((index) => values[index] !== null)
				.sort((a, b) => a - b)
		}));
}

/**
 * La cifra de cabecera: lo que rindió el dinero, en cuánto tiempo y sobre
 * cuánto capital.
 *
 * La rentabilidad del periodo es ponderada por tiempo —encadena tramos e
 * ignora cuándo entró cada aporte— y la ganancia sí depende de eso, así que un
 * +30 % de periodo puede convivir con un +10 % sobre coste cuando el dinero
 * grande entró después de la subida. Las dos salen de aquí juntas a propósito:
 * es la contradicción aparente que más confunde de la página, y la cabecera la
 * explica cuando de verdad se separan.
 */
export function buildRecordSummary(
	points: GrowthDataPoint[],
	summary: GrowthSummary
): RecordSummary | null {
	const series = sortedPoints(points);
	if (series.length === 0) return null;

	const last = series[series.length - 1];
	const returns = periodReturns(series);
	const historyDays = spanDays(series);
	const totalReturn = returns.length > 0 ? compound(returns.map((r) => r.value)) : null;
	const annualized =
		totalReturn === null || historyDays < MIN_ANNUALIZED_DAYS
			? null
			: annualize(totalReturn, historyDays);

	return {
		periodReturn: totalReturn === null ? null : totalReturn * 100,
		annualized: annualized === null ? null : annualized * 100,
		from: series[0].date,
		to: last.date,
		// 30,44 días de media por mes: con 30 justos, un historial de un año
		// natural salía de trece meses.
		months: Math.max(1, Math.round(historyDays / 30.44)),
		// El `||` cuenta el cero como ausencia a propósito: el loader rellena con
		// ceros el resumen que no llegó, y la serie sigue teniendo el valor bueno.
		value: toNumber(summary.currentValue) || toNumber(last.totalValue),
		cost: toNumber(last.totalCostBase),
		gain: amountOr(summary.gainLoss, last.gainLoss),
		gainPct: amountOr(summary.gainLossPct, last.gainLossPct),
		currency: summary.currency || 'USD'
	};
}

/**
 * Cuánto se movió la cuenta para llegar hasta ahí: los dos meses extremos y
 * las tres medidas de riesgo.
 *
 * Cada fila lleva en `hint` qué mide, y la tabla lo pinta a la vista. Antes
 * vivía en un `title`, que en un móvil no se abre y con el teclado tampoco: la
 * explicación de una medida como el Sharpe no es una ayuda opcional, es la
 * mitad del dato.
 *
 * Una medida que el historial todavía no sostiene sale como `N/A` con la razón
 * en su sitio, no con la celda vacía, que se leería como un cero.
 *
 * El Sharpe se anualiza por la vía de siempre —media aritmética de los tramos
 * × √tramos por año ÷ volatilidad—, que no es la rentabilidad compuesta de la
 * cabecera dividida entre la volatilidad. Multiplicarlo por la volatilidad da
 * una anualizada aritmética, más baja que la compuesta; el `hint` avisa para
 * que nadie cuadre una con la otra.
 */
export function buildKeyStatistics(points: GrowthDataPoint[]): KeyStat[] {
	const series = sortedPoints(points);
	if (series.length === 0) return [];

	const returns = periodReturns(series);
	const values = returns.map((r) => r.value);
	const historyDays = spanDays(series);

	// Un solo umbral de historial para las cifras anuales, y otro —de tramos—
	// para lo que solo pide dispersión: 90 días de una serie con cuatro puntos no
	// miden oscilación ninguna, y 60 días de una serie diaria sí.
	const hasAnnualSpan = historyDays >= MIN_ANNUALIZED_DAYS;
	const missingDispersion = `Necesita al menos ${MIN_RISK_RETURNS} tramos de historial; llevas ${returns.length}.`;
	const missingRisk = `Necesita al menos ${MIN_RISK_RETURNS} tramos y ${MIN_ANNUALIZED_DAYS} días de historial; llevas ${returns.length} y ${historyDays}.`;

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

	// El mejor y el peor mes se buscan solo entre los meses enteros: el primero
	// del historial y el que está en curso cubren unos pocos días, y un +0,4 %
	// de tres días entraba como «peor mes» delante de meses completos peores.
	// Sin ningún mes entero todavía se compara lo que hay, marcado con el mismo
	// asterisco que usa la matriz.
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

	// Con el año: el historial cruza años y «octubre» a secas no dice cuál.
	const monthLabel = (entry: [string, number] | null) => {
		if (entry === null) return undefined;
		const [year, month] = entry[0].split('-');
		return `${MONTHS_LONG[Number.parseInt(month, 10) - 1]} de ${year}${onlyPartial ? '*' : ''}`;
	};

	const monthHint = (entry: [string, number] | null, which: string) => {
		if (entry === null)
			return 'Ningún mes del historial tiene todavía un tramo con el que calcularlo.';

		return onlyPartial
			? `El mes que ${which} rindió. Está incompleto: su cifra cubre unos pocos días y no se compara con un mes entero.`
			: `El mes que ${which} rindió. Los meses incompletos —aquel en el que empieza el historial y el que está en curso— no compiten.`;
	};

	const monthNote = onlyPartial ? '* Mes incompleto.' : undefined;

	return [
		{
			label: 'Mejor mes',
			value: best === null ? UNAVAILABLE : formatSignedPercent(best[1] * 100),
			detail: monthLabel(best),
			hint: monthHint(best, 'más'),
			note: best === null ? undefined : monthNote,
			tone: best === null ? 'neutral' : toneOf(best[1] * 100)
		},
		{
			label: 'Peor mes',
			value: worst === null ? UNAVAILABLE : formatSignedPercent(worst[1] * 100),
			detail: monthLabel(worst),
			hint: monthHint(worst, 'menos'),
			note: worst === null ? undefined : monthNote,
			tone: worst === null ? 'neutral' : toneOf(worst[1] * 100)
		},
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
						? `Cuánto oscila la rentabilidad de un tramo del historial al siguiente, normalmente de un día al otro. Sin llevar a un año: eso pide ${MIN_ANNUALIZED_DAYS} días.`
						: 'Cuánto oscila tu rentabilidad de un día al siguiente, llevada a un año. Más alta es más movimiento, arriba y abajo.',
			note:
				volatility !== null && annualVolatility === null
					? `Sin anualizar: se lleva a un año a partir de ${MIN_ANNUALIZED_DAYS} días.`
					: undefined,
			tone: 'neutral'
		},
		{
			label: 'Máxima caída',
			value: returns.length > 0 ? formatPercent(maxDrawdown(values) * 100) : UNAVAILABLE,
			hint: 'La peor bajada desde un máximo hasta el siguiente suelo, medida tramo a tramo: puede caer dentro de un mes que cerró en positivo. Mide lo que habrías tenido que aguantar, no lo que perdiste.',
			tone: 'down'
		},
		{
			label: 'Ratio de Sharpe',
			value: sharpe === null ? UNAVAILABLE : RATIO.format(sharpe),
			hint:
				sharpe === null
					? missingRisk
					: 'Cuánta rentabilidad te dio cada unidad de riesgo, con tasa libre de riesgo 0. Sale de la media de los tramos anualizada, no de la rentabilidad de la cabecera: multiplicarlo por la volatilidad no devuelve aquella cifra.',
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
}
