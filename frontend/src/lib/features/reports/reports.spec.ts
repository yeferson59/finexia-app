import { describe, it, expect } from 'vitest';
import {
	buildGrowthProjection,
	buildKeyStatistics,
	buildPerformanceCalendars,
	performanceClass,
	projectionCoordinates
} from './reports';
import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

/** Punto de la serie de crecimiento; solo `date` y `totalValue` importan aquí. */
function point(date: string, totalValue: string): GrowthDataPoint {
	return { date, totalValue, totalCostBase: '0', gainLoss: '0', gainLossPct: '0' };
}

describe('performanceClass', () => {
	it('reparte los tramos de color por rentabilidad', () => {
		expect(performanceClass(3)).toBe('strong-positive');
		expect(performanceClass(1.5)).toBe('positive');
		expect(performanceClass(0)).toBe('flat-positive');
		expect(performanceClass(-0.5)).toBe('negative');
		expect(performanceClass(-4)).toBe('strong-negative');
	});
});

describe('buildPerformanceCalendars', () => {
	it('calcula la variación mes a mes y deja el primero sin dato', () => {
		const [calendar] = buildPerformanceCalendars([
			point('2026-01-31', '1000'),
			point('2026-02-28', '1100'),
			point('2026-03-31', '1045')
		]);

		expect(calendar.year).toBe('2026');
		expect(calendar.values[0]).toBeNull();
		expect(calendar.values[1]).toBe(10);
		expect(calendar.values[2]).toBe(-5);
		// Los meses sin dato se quedan vacíos.
		expect(calendar.values.slice(3)).toEqual(Array(9).fill(null));
	});

	it('se queda con el último punto de cada mes', () => {
		const [calendar] = buildPerformanceCalendars([
			point('2026-01-15', '900'),
			point('2026-01-31', '1000'),
			point('2026-02-28', '1200')
		]);

		expect(calendar.values[1]).toBe(20);
	});

	it('ordena los años del más reciente al más antiguo', () => {
		const calendars = buildPerformanceCalendars([
			point('2025-11-30', '1000'),
			point('2025-12-31', '1100'),
			point('2026-01-31', '1210')
		]);

		expect(calendars.map((c) => c.year)).toEqual(['2026', '2025']);
	});

	it('no inventa una variación cuando el mes previo valía cero', () => {
		const [calendar] = buildPerformanceCalendars([
			point('2026-01-31', '0'),
			point('2026-02-28', '500')
		]);

		expect(calendar.values[1]).toBeNull();
	});

	it('devuelve una lista vacía sin historial', () => {
		expect(buildPerformanceCalendars([])).toEqual([]);
	});
});

describe('buildKeyStatistics', () => {
	it('mide la mayor caída desde un pico', () => {
		const [drawdown] = buildKeyStatistics([
			point('2026-01-31', '1000'),
			point('2026-02-28', '1200'),
			point('2026-03-31', '900')
		]);

		expect(drawdown).toEqual({ label: 'Max Drawdown', value: '-25.0%' });
	});

	it('no da volatilidad con menos de tres retornos mensuales', () => {
		const stats = buildKeyStatistics([point('2026-01-31', '1000'), point('2026-02-28', '1100')]);

		expect(stats[1]).toEqual({ label: 'Volatilidad', value: 'N/A' });
	});

	it('anualiza la volatilidad cuando hay retornos suficientes', () => {
		const stats = buildKeyStatistics([
			point('2026-01-31', '1000'),
			point('2026-02-28', '1100'),
			point('2026-03-31', '1050'),
			point('2026-04-30', '1150')
		]);

		expect(stats[1].value).toMatch(/^\d+\.\d%$/);
		expect(stats[1].value).not.toBe('N/A');
	});

	it('devuelve una lista vacía sin historial', () => {
		expect(buildKeyStatistics([])).toEqual([]);
	});
});

describe('buildGrowthProjection', () => {
	const summary = (over: Partial<GrowthSummary> = {}): GrowthSummary => ({
		firstDate: '2024-01-01',
		initialValue: '1000',
		currentValue: '1500',
		totalGrowthPct: '50',
		...over
	});

	it('proyecta cinco años desde el CAGR del historial', () => {
		const projection = buildGrowthProjection([point('2026-01-01', '1500')], summary());

		expect(projection).toHaveLength(5);
		expect(projection[0]).toEqual({ period: '2026', value: 1500 });
		expect(projection.map((p) => p.period)).toEqual(['2026', '2027', '2028', '2029', '2030']);
		// Con un CAGR positivo la serie crece de forma monótona.
		expect(projection[4].value).toBeGreaterThan(projection[0].value);
	});

	it('se abstiene con menos de medio año de historial', () => {
		const projection = buildGrowthProjection(
			[point('2026-03-01', '1500')],
			summary({ firstDate: '2026-01-01' })
		);

		expect(projection).toEqual([]);
	});

	it('se abstiene con un CAGR fuera de rango o valores no positivos', () => {
		// x10 en dos años: fuera del rango plausible.
		expect(
			buildGrowthProjection([point('2026-01-01', '10000')], summary({ currentValue: '10000' }))
		).toEqual([]);
		expect(
			buildGrowthProjection([point('2026-01-01', '0')], summary({ initialValue: '0' }))
		).toEqual([]);
	});

	it('se abstiene sin puntos', () => {
		expect(buildGrowthProjection([], summary())).toEqual([]);
	});
});

describe('projectionCoordinates', () => {
	it('reparte los puntos en el eje x y estira los valores al alto del viewBox', () => {
		const coords = projectionCoordinates([
			{ period: '2026', value: 100 },
			{ period: '2027', value: 150 },
			{ period: '2028', value: 200 }
		]);

		expect(coords.map((c) => c.x)).toEqual([40, 170, 300]);
		// El mínimo se apoya en la base y el máximo llega al techo del área útil.
		expect(coords[0].y).toBe(230);
		expect(coords[2].y).toBe(50);
		expect(coords[1].y).toBe(140);
	});

	it('no divide por cero cuando todos los valores son iguales', () => {
		const coords = projectionCoordinates([
			{ period: '2026', value: 100 },
			{ period: '2027', value: 100 }
		]);

		expect(coords.every((c) => c.y === 230)).toBe(true);
	});

	it('devuelve una lista vacía sin proyección', () => {
		expect(projectionCoordinates([])).toEqual([]);
	});
});
