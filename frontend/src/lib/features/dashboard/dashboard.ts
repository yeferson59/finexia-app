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

/**
 * Escala vertical de la gráfica, con un 3 % de aire arriba y abajo para que la
 * línea no toque los bordes. Sin datos, una escala neutra que no divide por 0.
 */
export function growthScale(values: number[]): { yMin: number; yMax: number; yRange: number } {
	const yMin = values.length > 0 ? Math.min(...values) * 0.97 : 0;
	const yMax = values.length > 0 ? Math.max(...values) * 1.03 : 1;
	return { yMin, yMax, yRange: yMax - yMin || 1 };
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
