/**
 * Proyección del valor de la cartera a partir de lo que ya ganó.
 *
 * Vive aparte de `dashboard.ts` —que arma la escala y la geometría de la
 * gráfica— porque no comparte nada con ella salvo la serie de la que sale, y
 * el módulo común se acercaba al presupuesto de 500 líneas.
 *
 * Extrapola **rentabilidad mensual, nunca la variación del saldo**. La
 * distinción no es un matiz: una cuenta que pasó de 242 a 1.953 mientras su
 * dueño ingresaba 1.700 no creció un 704 %, ganó un 11 %, y proyectar la
 * primera cifra prometería multiplicar el patrimonio cada trimestre. El
 * retorno de cada mes sale de `$lib/shared/finance/returns`, que descuenta
 * aportes y retiros con Dietz modificada.
 *
 * Es una extrapolación, no una previsión, y por eso no es una línea sino una
 * banda: el mejor mes que tuvo la cartera repetido doce veces marca el techo, y
 * el peor repetido doce veces marca el suelo. Una media escondería que la
 * diferencia entre los dos es justo lo que no se sabe; los dos extremos la
 * enseñan, y lo que hay entre ellos es el rango de lo que ya ha pasado de
 * verdad, sin inventar ninguna cifra intermedia.
 *
 * La proyección publica de cuántos meses salen los dos extremos, para que quien
 * la lea sepa lo delgada que es la base.
 */

import type { GrowthDataPoint } from '$lib/api/types';
import {
	dayGap,
	incompleteMonths,
	periodReturns,
	returnsByMonth,
	sortedPoints,
	toNumber
} from '$lib/shared/finance/returns';

/** Meses que se proyectan hacia adelante. */
export const FORECAST_MONTHS = 12;

/**
 * Meses completos mínimos para proyectar.
 *
 * Dos, que es lo mínimo con lo que un mejor y un peor mes son dos meses
 * distintos: con uno solo la banda sería una línea. Tres meses de historial dan
 * dos meses completos —el mes en que abrió la serie y el mes en curso van a
 * medias—, así que es lo que hace falta para ver la proyección.
 */
export const FORECAST_MIN_MONTHS = 2;

const MONTHS_PER_YEAR = 12;

/** Un mes proyectado, con el techo y el suelo de la banda. */
export interface ForecastEntry {
	/** Último día del mes proyectado, `YYYY-MM-DD`. */
	date: string;
	/** Cuántos meses hay desde hoy hasta este punto: 1 es el mes que viene. */
	monthsAhead: number;
	/** Valor si se repitiera el peor mes del historial, en la moneda de la serie. */
	low: number;
	/** Valor si se repitiera el mejor. */
	high: number;
	/** Lo acumulado desde hoy hasta ese mes con cada extremo, en porcentaje. */
	lowPct: number;
	highPct: number;
}

/** La proyección completa: los meses y los dos meses de los que salen. */
export interface ValueForecast {
	/** El mejor y el peor mes completo del historial, en porcentaje. */
	bestMonthPct: number;
	worstMonthPct: number;
	/** Los dos extremos compuestos a un año, que es como se comparan las tasas. */
	bestAnnualPct: number;
	worstAnnualPct: number;
	/** Cuántos meses completos hay detrás de los dos extremos. */
	basisMonths: number;
	entries: ForecastEntry[];
}

/**
 * Último día del mes que cae `monthsAhead` meses después del de `iso`.
 *
 * El día 0 de un mes es el último del anterior, que es la forma de pedir «fin
 * de mes» sin tener que saber cuál dura 31, 30 o 28.
 */
function monthEndAhead(iso: string, monthsAhead: number): string {
	const year = Number.parseInt(iso.slice(0, 4), 10);
	const month = Number.parseInt(iso.slice(5, 7), 10);
	if (!Number.isFinite(year) || !Number.isFinite(month)) return iso;

	return new Date(Date.UTC(year, month + monthsAhead, 0)).toISOString().slice(0, 10);
}

/**
 * Los retornos de los meses que el historial cubre enteros.
 *
 * Los incompletos quedan fuera: el mes en que empezó la serie corre desde el
 * día que abrió, y el último está a medias mientras el mes no acabe. Tres días
 * de septiembre no son un mes, y meterlos en la media la hundiría o la
 * dispararía según el humor de esa semana.
 */
export function completeMonthlyReturns(points: GrowthDataPoint[]): number[] {
	const series = sortedPoints(points);
	if (series.length < 2) return [];

	const partial = incompleteMonths(series);

	return [...returnsByMonth(periodReturns(series))]
		.filter(([month]) => !partial.has(month))
		.sort(([a], [b]) => a.localeCompare(b))
		.map(([, value]) => value);
}

/**
 * Proyección a doce meses entre el mejor y el peor mes del historial.
 *
 * No hay media: los dos extremos son dos meses que de verdad ocurrieron, y
 * repetirlos doce veces da el techo y el suelo de lo que la cartera ya ha
 * demostrado saber hacer. Promediarlos daría una cifra que no le pasó a nadie y
 * escondería lo único que importa de una extrapolación, que es cuánto se puede
 * separar de sí misma.
 *
 * Los dos extremos van sin recortar. Un mes excepcional produce un techo
 * empinado, y eso es exactamente lo que dice la etiqueta: qué pasaría si se
 * repitiera ese mes. Recortarlo dejaría la etiqueta mintiendo.
 *
 * Devuelve `null` en cuanto el dato no da: menos de dos meses completos, un
 * valor actual no positivo o una pérdida total en la base. El panel entonces
 * dice qué falta en vez de dibujar una banda inventada.
 */
export function buildValueForecast(
	points: GrowthDataPoint[],
	currentValue: number
): ValueForecast | null {
	const series = sortedPoints(points);
	if (series.length < 2) return null;

	const start =
		Number.isFinite(currentValue) && currentValue > 0
			? currentValue
			: toNumber(series[series.length - 1].totalValue);
	if (!Number.isFinite(start) || start <= 0) return null;

	const monthly = completeMonthlyReturns(series);
	if (monthly.length < FORECAST_MIN_MONTHS) return null;

	const best = Math.max(...monthly);
	const worst = Math.min(...monthly);
	// Un mes que se lo llevó todo deja el suelo en cero y la banda sin fondo: no
	// hay tasa que repetir doce veces a partir de ahí.
	if (!Number.isFinite(best) || !Number.isFinite(worst) || worst <= -1) return null;

	const last = series[series.length - 1].date;

	return {
		bestMonthPct: best * 100,
		worstMonthPct: worst * 100,
		bestAnnualPct: (Math.pow(1 + best, MONTHS_PER_YEAR) - 1) * 100,
		worstAnnualPct: (Math.pow(1 + worst, MONTHS_PER_YEAR) - 1) * 100,
		basisMonths: monthly.length,
		entries: Array.from({ length: FORECAST_MONTHS }, (_, i) => {
			const monthsAhead = i + 1;
			const highGrowth = Math.pow(1 + best, monthsAhead) - 1;
			const lowGrowth = Math.pow(1 + worst, monthsAhead) - 1;

			return {
				date: monthEndAhead(last, monthsAhead),
				monthsAhead,
				low: start * (1 + lowGrowth),
				high: start * (1 + highGrowth),
				lowPct: lowGrowth * 100,
				highPct: highGrowth * 100
			};
		})
	};
}

/**
 * La curva proyectada, al mismo paso con el que avanza el historial.
 *
 * La gráfica reparte los puntos por índice, no por fecha: doce puntos mensuales
 * detrás de ochenta y seis diarios meterían un año entero en la doceava parte
 * del ancho, y la proyección se leería como si cubriera una semana. Al paso
 * medio del historial —sus días partidos por sus tramos— un mes de futuro ocupa
 * lo que ocupa un mes de pasado, que es lo único que hace comparable el eje.
 *
 * El primer punto es el último del historial, para que las dos curvas empalmen
 * en el mismo sitio; quien las dibuja lo usa de costura y no lo repite.
 */
export function forecastCurve(
	points: GrowthDataPoint[],
	forecast: ValueForecast | null,
	startValue: number
): { date: string; low: number; high: number }[] {
	if (!forecast || forecast.entries.length === 0) return [];

	const series = sortedPoints(points);
	if (series.length < 2) return [];

	const from = series[series.length - 1].date;
	const end = forecast.entries[forecast.entries.length - 1];
	const spanDays = dayGap(from, end.date);
	if (spanDays <= 0) return [];

	// El paso del historial: sus días entre sus tramos. Nunca menos de uno, que
	// es lo que evita una curva de miles de puntos si la serie trae dos fechas
	// iguales.
	const step = Math.max(1, dayGap(series[0].date, from) / (series.length - 1));
	const best = forecast.bestMonthPct / 100;
	const worst = forecast.worstMonthPct / 100;

	const curve: { date: string; low: number; high: number }[] = [];
	for (let day = 0; day <= spanDays; day += step) {
		const months = Math.round(day) / DAYS_PER_MONTH;
		curve.push({
			date: addDays(from, Math.round(day)),
			low: startValue * Math.pow(1 + worst, months),
			high: startValue * Math.pow(1 + best, months)
		});
	}

	// El último punto es el del último mes, exacto: las cifras que enseña la
	// tabla de abajo tienen que ser a las que llegan los dos bordes.
	const closing = { date: end.date, low: end.low, high: end.high };
	if (curve[curve.length - 1].date === end.date) curve[curve.length - 1] = closing;
	else curve.push(closing);

	return curve;
}

/** Días de un mes medio, para repartir la tasa mensual entre los días. */
const DAYS_PER_MONTH = 365.25 / MONTHS_PER_YEAR;

/** Suma días a una fecha `YYYY-MM-DD`, en UTC. */
function addDays(iso: string, days: number): string {
	return new Date(Date.parse(`${iso}T00:00:00Z`) + days * 86_400_000).toISOString().slice(0, 10);
}

/**
 * Los meses que se destacan bajo la gráfica: el trimestre, el semestre y el año.
 *
 * Doce filas serían la tabla de amortización de una hipoteca. Tres marcas dicen
 * lo mismo —a dónde lleva la tasa— y se leen de un vistazo.
 */
export const FORECAST_MILESTONES = [3, 6, 12];

export function forecastMilestones(forecast: ValueForecast | null): ForecastEntry[] {
	if (!forecast) return [];

	return FORECAST_MILESTONES.map((month) =>
		forecast.entries.find((entry) => entry.monthsAhead === month)
	).filter((entry): entry is ForecastEntry => entry !== undefined);
}
