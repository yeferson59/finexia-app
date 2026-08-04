/**
 * Helpers puros de los widgets del dashboard: la geometría del donut de
 * asignación y la de la gráfica de crecimiento.
 *
 * Vivían dentro de sus componentes, donde no había forma de probarlas: son
 * matemáticas de SVG que se rompen en silencio (un ángulo mal, un eje que se
 * sale del lienzo) porque el widget sigue pintando algo.
 */

import type { AllocationItem, GrowthDataPoint } from '$lib/api/types';

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

/** Punto de la serie de crecimiento, ya convertido a números. */
export interface GrowthPoint {
	date: string;
	/** Valor de mercado. */
	mv: number;
	/** Capital invertido (coste). */
	cb: number;
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
