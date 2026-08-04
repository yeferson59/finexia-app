import { describe, it, expect } from 'vitest';
import {
	PLOT,
	buildSlices,
	filterByPeriod,
	generatePieSlice,
	growthScale,
	nearestIndex,
	polarToCartesian,
	toAssetEntries,
	toPlotX,
	toPlotY
} from './dashboard';
import type { AllocationItem, GrowthDataPoint } from '$lib/api/types';

const item = (category: string, marketValue: string, percent: number): AllocationItem => ({
	category,
	marketValue,
	percent
});

describe('toAssetEntries', () => {
	it('traduce categoría, importe y color', () => {
		expect(toAssetEntries([item('stocks', '1900.50', 60)])).toEqual([
			{ name: 'Acciones', value: 1900.5, percent: 60, color: '#d4912a' }
		]);
	});

	it('no pierde una categoría que el backend añada por su cuenta', () => {
		const [entry] = toAssetEntries([item('nft', '10', 5)]);
		expect(entry.name).toBe('nft');
		expect(entry.color).toBe('#5ab4e0');
	});

	it('trata un importe vacío como cero', () => {
		expect(toAssetEntries([item('cash', '', 0)])[0].value).toBe(0);
	});
});

describe('polarToCartesian', () => {
	it('sitúa el cero grados arriba del círculo', () => {
		const p = polarToCartesian(0, 75);
		expect(p.x).toBeCloseTo(100);
		expect(p.y).toBeCloseTo(25);
	});

	it('avanza en el sentido de las agujas del reloj', () => {
		const p = polarToCartesian(90, 75);
		expect(p.x).toBeCloseTo(175);
		expect(p.y).toBeCloseTo(100);
	});
});

describe('generatePieSlice', () => {
	it('convierte el porcentaje en grados', () => {
		expect(generatePieSlice(25, 0).endAngle).toBe(90);
		expect(generatePieSlice(50, 90).endAngle).toBe(270);
	});

	it('marca el arco largo solo por encima de media vuelta', () => {
		expect(generatePieSlice(50, 0).d).toContain('A 75 75 0 0 1');
		expect(generatePieSlice(75, 0).d).toContain('A 75 75 0 1 1');
	});
});

describe('buildSlices', () => {
	it('encadena las porciones sin dejar hueco entre ellas', () => {
		const slices = buildSlices(toAssetEntries([item('stocks', '60', 60), item('cash', '40', 40)]));

		expect(slices[0].startAngle).toBe(0);
		expect(slices[0].endAngle).toBe(216);
		expect(slices[1].startAngle).toBe(216);
		expect(slices[1].endAngle).toBe(360);
	});
});

describe('filterByPeriod', () => {
	const now = new Date('2026-07-15T12:00:00');
	const points: GrowthDataPoint[] = ['2026-07-01', '2026-05-01', '2026-01-01', '2025-01-01'].map(
		(date) => ({ date, totalValue: '1', totalCostBase: '1', gainLoss: '0', gainLossPct: '0' })
	);

	it('devuelve la serie entera con «Todo»', () => {
		expect(filterByPeriod(points, 'Todo', now)).toHaveLength(4);
	});

	it('recorta al último mes', () => {
		expect(filterByPeriod(points, '1M', now).map((p) => p.date)).toEqual(['2026-07-01']);
	});

	it('recorta a los últimos tres meses y al último año', () => {
		expect(filterByPeriod(points, '3M', now)).toHaveLength(2);
		expect(filterByPeriod(points, '1Y', now)).toHaveLength(3);
	});
});

describe('growthScale', () => {
	it('redondea el dominio hacia fuera sin recortar la serie', () => {
		const { yMin, yMax } = growthScale([137, 892]);
		expect(yMin).toBeLessThanOrEqual(137);
		expect(yMax).toBeGreaterThanOrEqual(892);
	});

	it('da marcas en números redondos, todas múltiplos del mismo paso', () => {
		const { ticks } = growthScale([100, 200]);
		expect(ticks).toEqual([200, 180, 160, 140, 120, 100]);
	});

	it('ordena las marcas de mayor a menor y cubre el dominio entero', () => {
		const { ticks, yMin, yMax } = growthScale([1234, 98765]);
		expect(ticks[0]).toBe(yMax);
		expect(ticks[ticks.length - 1]).toBe(yMin);
		expect([...ticks].sort((a, b) => b - a)).toEqual(ticks);
	});

	it('no divide por cero con una serie plana', () => {
		const scale = growthScale([100, 100]);
		expect(scale.yRange).not.toBe(0);
		expect(scale.yMin).toBeLessThan(100);
		expect(scale.yMax).toBeGreaterThan(100);
	});

	it('no divide por cero con una serie plana en cero', () => {
		expect(growthScale([0, 0]).yRange).not.toBe(0);
	});

	it('da una escala neutra sin datos', () => {
		const scale = growthScale([]);
		expect(scale.yMin).toBe(0);
		expect(scale.yMax).toBe(1);
		expect(scale.yRange).toBe(1);
	});

	it('ignora los valores no finitos en vez de romper la escala', () => {
		const scale = growthScale([100, Number.NaN, 200, Number.POSITIVE_INFINITY]);
		expect(Number.isFinite(scale.yMin)).toBe(true);
		expect(Number.isFinite(scale.yMax)).toBe(true);
	});
});

describe('nearestIndex', () => {
	it('devuelve el punto bajo el cursor', () => {
		expect(nearestIndex(PLOT.padL, 5)).toBe(0);
		expect(nearestIndex(PLOT.padL + PLOT.plotW, 5)).toBe(4);
		expect(nearestIndex(PLOT.padL + PLOT.plotW / 2, 5)).toBe(2);
	});

	it('recorta a los extremos cuando el cursor sale del lienzo', () => {
		expect(nearestIndex(-500, 5)).toBe(0);
		expect(nearestIndex(5000, 5)).toBe(4);
	});

	it('con un solo punto siempre devuelve ese punto', () => {
		expect(nearestIndex(1234, 1)).toBe(0);
	});
});

describe('toPlotX / toPlotY', () => {
	it('reparte los puntos entre los márgenes del lienzo', () => {
		expect(toPlotX(0, 3)).toBe(PLOT.padL);
		expect(toPlotX(2, 3)).toBe(PLOT.padL + PLOT.plotW);
	});

	it('centra un punto único', () => {
		expect(toPlotX(0, 1)).toBe(PLOT.padL + PLOT.plotW / 2);
	});

	it('invierte el eje vertical: el máximo arriba y el mínimo abajo', () => {
		const scale = growthScale([100, 200]);
		expect(toPlotY(scale.yMax, scale)).toBeCloseTo(PLOT.padT);
		expect(toPlotY(scale.yMin, scale)).toBeCloseTo(PLOT.padT + PLOT.plotH);
	});

	it('mantiene la línea dentro del lienzo', () => {
		const scale = growthScale([1500, 1900]);
		for (const v of [1500, 1700, 1900]) {
			const y = toPlotY(v, scale);
			expect(y).toBeGreaterThanOrEqual(PLOT.padT);
			expect(y).toBeLessThanOrEqual(PLOT.padT + PLOT.plotH);
		}
	});
});
