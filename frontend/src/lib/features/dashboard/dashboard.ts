/**
 * Helpers puros de los widgets del dashboard: la geometría del donut de
 * asignación y la de la gráfica de crecimiento.
 *
 * Vivían dentro de sus componentes, donde no había forma de probarlas: son
 * matemáticas de SVG que se rompen en silencio (un ángulo mal, un eje que se
 * sale del lienzo) porque el widget sigue pintando algo.
 */

import type { AllocationItem, GrowthDataPoint } from '$lib/api/types';
import { cumulativeReturns } from '$lib/shared/finance/returns';

// ---------------------------------------------------------------------------
// Asignación de activos (donut)
// ---------------------------------------------------------------------------

/** Entrada del donut, ya resuelta a etiqueta, importe y color. */
export interface AssetEntry {
	name: string;
	value: number;
	percent: number;
	color: string;
}

export const CATEGORY_LABELS: Record<string, string> = {
	stocks: 'Acciones',
	etfs: 'ETFs',
	cryptos: 'Crypto',
	bonds: 'Bonos',
	cash: 'Efectivo',
	real_estates: 'Inmuebles',
	commodities: 'Commodities',
	others: 'Otros'
};

export const CATEGORY_COLORS: Record<string, string> = {
	stocks: '#d4912a',
	etfs: '#22c97e',
	cryptos: '#6b8cef',
	bonds: '#b988e0',
	cash: '#8a8780',
	real_estates: '#e0885a',
	commodities: '#e0c15a',
	others: '#5ab4e0'
};

/** Color de reserva de una categoría que el backend añada y aquí no esté. */
const FALLBACK_COLOR = '#5ab4e0';

/**
 * Traduce la asignación del backend a entradas del donut. Una categoría
 * desconocida conserva su nombre crudo en vez de desaparecer del gráfico.
 */
export function toAssetEntries(allocation: AllocationItem[]): AssetEntry[] {
	return allocation.map((item) => ({
		name: CATEGORY_LABELS[item.category] ?? item.category,
		value: parseFloat(item.marketValue || '0'),
		percent: item.percent,
		color: CATEGORY_COLORS[item.category] ?? FALLBACK_COLOR
	}));
}

/** Punto del círculo en grados, con 0º arriba (de ahí el -90). */
export function polarToCartesian(angle: number, radius: number, cx = 100, cy = 100) {
	const radians = (angle - 90) * (Math.PI / 180);
	return {
		x: cx + radius * Math.cos(radians),
		y: cy + radius * Math.sin(radians)
	};
}

/** Porción del donut como `path`, con el flag de arco largo por encima de 180º. */
export function generatePieSlice(
	percent: number,
	startAngle: number
): { d: string; startAngle: number; endAngle: number } {
	const cx = 100;
	const cy = 100;
	const radius = 75;
	const endAngle = startAngle + (percent / 100) * 360;
	const largeArc = endAngle - startAngle > 180 ? 1 : 0;

	const startPoint = polarToCartesian(startAngle, radius, cx, cy);
	const endPoint = polarToCartesian(endAngle, radius, cx, cy);

	const d = [
		`M ${cx} ${cy}`,
		`L ${startPoint.x} ${startPoint.y}`,
		`A ${radius} ${radius} 0 ${largeArc} 1 ${endPoint.x} ${endPoint.y}`,
		'Z'
	].join(' ');

	return { d, startAngle, endAngle };
}

/** Encadena las porciones: cada una arranca donde acabó la anterior. */
export function buildSlices(items: AssetEntry[]) {
	let angle = 0;
	return items.map((asset) => {
		const slice = generatePieSlice(asset.percent, angle);
		angle = slice.endAngle;
		return { ...asset, ...slice };
	});
}

// ---------------------------------------------------------------------------
// Crecimiento del portafolio (gráfica de líneas)
// ---------------------------------------------------------------------------

export type Period = '1M' | '3M' | '6M' | '1Y' | 'Todo';

export const PERIODS: Period[] = ['1M', '3M', '6M', '1Y', 'Todo'];

/** Medidas del lienzo SVG, compartidas por ejes, líneas y relleno. */
export const PLOT = {
	padL: 52,
	padR: 20,
	padT: 20,
	padB: 32,
	svgW: 600,
	svgH: 240,
	get plotW() {
		return this.svgW - this.padL - this.padR;
	},
	get plotH() {
		return this.svgH - this.padT - this.padB;
	}
};

/**
 * Recorta la serie al periodo elegido. `now` es un parámetro para poder
 * probarlo sin depender del reloj.
 */
export function filterByPeriod(
	points: GrowthDataPoint[],
	period: Period,
	now: Date = new Date()
): GrowthDataPoint[] {
	if (period === 'Todo') return points;
	const monthsMap: Record<Exclude<Period, 'Todo'>, number> = {
		'1M': 1,
		'3M': 3,
		'6M': 6,
		'1Y': 12
	};
	const cutoff = new Date(now.getFullYear(), now.getMonth() - monthsMap[period], now.getDate());
	return points.filter((d) => new Date(d.date) >= cutoff);
}

/**
 * Punto de la serie que dibuja la gráfica, ya convertido a números.
 *
 * Los dos campos son los dos trazos, y lo que miden depende de la vista: en la
 * de dinero son valor de mercado y capital invertido; en la de porcentaje,
 * rentabilidad acumulada y ganancia sobre coste. Los nombres se quedaron con
 * los de la vista original —renombrarlos a `a` y `b` habría hecho ilegible el
 * lienzo— y `GROWTH_LABELS` es quien dice, en cada vista, cómo se llama cada
 * uno de cara al usuario.
 */
export interface GrowthPoint {
	date: string;
	/** Trazo principal (línea ámbar). */
	mv: number;
	/** Trazo de referencia (línea gris discontinua). */
	cb: number;
}

/**
 * Qué mide la gráfica.
 *
 * El dinero cuenta la historia de la cuenta —cuánto hay— y el porcentaje la del
 * inversionista: un depósito sube la primera curva sin que nada se haya
 * revalorizado, y deja la segunda donde estaba. Son la misma serie leída de dos
 * formas, y por eso conviven en un conmutador y no en dos gráficas.
 */
export type GrowthView = 'value' | 'percent';

export const GROWTH_VIEWS: { id: GrowthView; label: string; hint: string }[] = [
	{ id: 'value', label: 'Valor', hint: 'Ver la gráfica en dinero' },
	{ id: 'percent', label: '%', hint: 'Ver la gráfica en rentabilidad' }
];

/** Cómo se llaman los dos trazos —y la tabla accesible— en cada vista. */
export const GROWTH_LABELS: Record<
	GrowthView,
	{ primary: string; secondary: string; caption: string }
> = {
	value: {
		primary: 'Valor de mercado',
		secondary: 'Capital invertido',
		caption: 'Valor de mercado y capital invertido del portafolio, por fecha'
	},
	percent: {
		primary: 'Rentabilidad acumulada',
		secondary: 'Ganancia sobre coste',
		caption: 'Rentabilidad acumulada y ganancia sobre coste del portafolio, por fecha'
	}
};

/** La serie en dinero: valor de mercado contra capital invertido. */
export function toValuePoints(points: GrowthDataPoint[]): GrowthPoint[] {
	return points.map((point) => ({
		date: point.date,
		mv: parseFloat(point.totalValue || '0'),
		cb: parseFloat(point.totalCostBase || '0')
	}));
}

/**
 * La serie en porcentaje: rentabilidad acumulada contra ganancia sobre coste.
 *
 * La primera arranca en 0 % el día en que empieza el tramo dibujado y descuenta
 * aportes y retiros, así que es lo que rindió el dinero. La segunda es la foto
 * de cada día —mercado sobre coste—, que sí se mueve con cuándo entró cada
 * aporte; verlas juntas es lo que enseña la diferencia entre haber ganado y
 * haber ingresado.
 *
 * Una fecha sin capital invertido no tiene ganancia sobre coste que enseñar
 * —dividir por cero no da un 0 %, da nada—; se dibuja en la línea de equilibrio
 * porque una gráfica no puede tener agujeros, y es el día en que la cuenta
 * estaba vacía.
 */
export function toPercentPoints(points: GrowthDataPoint[]): GrowthPoint[] {
	return cumulativeReturns(points).map((point) => ({
		date: point.date,
		mv: point.twr * 100,
		cb: (point.overCost ?? 0) * 100
	}));
}

/** La serie lista para dibujar en la vista pedida. */
export function toGrowthPoints(points: GrowthDataPoint[], view: GrowthView): GrowthPoint[] {
	return view === 'percent' ? toPercentPoints(points) : toValuePoints(points);
}

/** Escala vertical de la gráfica, con las marcas del eje ya calculadas. */
export interface GrowthScale {
	yMin: number;
	yMax: number;
	yRange: number;
	/** Marcas del eje, de mayor a menor; todas múltiplos del mismo paso. */
	ticks: number[];
}

/**
 * Redondea un número al 1, 2, 5 o 10 más cercano de su magnitud. Es lo que
 * separa un eje con marcas de 25.000 de uno con marcas de 27.431,66.
 */
function niceNum(range: number, round: boolean): number {
	if (!Number.isFinite(range) || range <= 0) return 1;
	const exponent = Math.floor(Math.log10(range));
	const fraction = range / 10 ** exponent;
	let nice: number;
	if (round) {
		if (fraction < 1.5) nice = 1;
		else if (fraction < 3) nice = 2;
		else if (fraction < 7) nice = 5;
		else nice = 10;
	} else {
		if (fraction <= 1) nice = 1;
		else if (fraction <= 2) nice = 2;
		else if (fraction <= 5) nice = 5;
		else nice = 10;
	}
	return nice * 10 ** exponent;
}

/** Marcas objetivo del eje; el paso redondeado puede dar una o dos más. */
const GROWTH_TICKS = 5;

/** Quita la basura decimal que deja sumar el paso en coma flotante. */
const clean = (v: number) => Number(v.toFixed(10));

function scaleFrom(min: number, max: number): GrowthScale {
	const step = niceNum((max - min) / (GROWTH_TICKS - 1), true);
	const yMin = Math.floor(min / step) * step;
	const yMax = Math.ceil(max / step) * step;
	const yRange = yMax - yMin || step;

	const ticks: number[] = [];
	for (let value = yMax; value >= yMin - step / 2; value -= step) ticks.push(clean(value));

	return { yMin: clean(yMin), yMax: clean(yMax), yRange: clean(yRange), ticks };
}

/**
 * Escala vertical de la gráfica. El dominio se redondea hacia fuera hasta un
 * paso "bonito", de modo que las marcas del eje caen en números legibles y la
 * serie sigue cabiendo entera. Sin datos, una escala neutra que no divide por 0.
 */
export function growthScale(values: number[]): GrowthScale {
	const finite = values.filter((v) => Number.isFinite(v));
	if (finite.length === 0) return scaleFrom(0, 1);

	const min = Math.min(...finite);
	const max = Math.max(...finite);
	// Serie plana: abre una ventana alrededor del valor para que la línea quede
	// centrada en vez de pegada a un borde, y el rango nunca sea 0.
	if (min === max) {
		const pad = Math.abs(min) * 0.05 || 1;
		return scaleFrom(min - pad, max + pad);
	}
	return scaleFrom(min, max);
}

/**
 * Índice del punto más cercano a una posición horizontal del lienzo. Es lo que
 * convierte el ratón (o las flechas del teclado) en un punto de la serie.
 */
export function nearestIndex(x: number, n: number): number {
	if (n <= 1) return 0;
	const ratio = (x - PLOT.padL) / PLOT.plotW;
	return Math.min(n - 1, Math.max(0, Math.round(ratio * (n - 1))));
}

/** Posición horizontal del punto `i` de `n`; con un solo punto, centrado. */
export function toPlotX(i: number, n: number): number {
	if (n <= 1) return PLOT.padL + PLOT.plotW / 2;
	return PLOT.padL + (i / (n - 1)) * PLOT.plotW;
}

/** Posición vertical de un valor dentro de la escala (el eje y crece hacia abajo). */
export function toPlotY(v: number, scale: { yMin: number; yRange: number }): number {
	return PLOT.padT + PLOT.plotH - ((v - scale.yMin) / scale.yRange) * PLOT.plotH;
}
