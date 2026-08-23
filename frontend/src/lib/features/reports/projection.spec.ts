import { describe, it, expect } from 'vitest';
import { buildGrowthProjection, historySpanDays, projectionCoordinates } from './projection';
import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

/** Punto de la serie: valor de mercado y capital invertido a esa fecha. */
function point(date: string, totalValue: string, cost = '0'): GrowthDataPoint {
	const gainLoss = String(Number(totalValue) - Number(cost));
	return {
		date,
		totalValue,
		totalCostBase: cost,
		gainLoss,
		gainLossPct: cost === '0' ? '0' : String((Number(gainLoss) / Number(cost)) * 100)
	};
}

/** Serie diaria a partir de `2026-01-01`, con el capital invertido fijo. */
function dailySeries(values: number[], cost = '1000'): GrowthDataPoint[] {
	return values.map((value, i) => {
		const day = new Date(Date.UTC(2026, 0, 1 + i)).toISOString().substring(0, 10);
		return point(day, String(value), cost);
	});
}

const summary = (over: Partial<GrowthSummary> = {}): GrowthSummary => ({
	firstDate: '2026-01-01',
	initialValue: '1000',
	currentValue: '1500',
	totalGrowthPct: '50',
	currency: 'USD',
	...over
});

describe('historySpanDays', () => {
	it('cuenta los días entre el primer punto y el último', () => {
		expect(historySpanDays([point('2026-01-01', '1'), point('2026-03-02', '1')])).toBe(60);
	});

	it('es cero sin dos puntos que comparar', () => {
		expect(historySpanDays([point('2026-01-01', '1')])).toBe(0);
		expect(historySpanDays([])).toBe(0);
	});
});

describe('buildGrowthProjection', () => {
	/** Serie de un año que gana un 20 % de mercado, sin aportes. */
	const oneYear = () => {
		const values = Array.from({ length: 366 }, (_, i) => 1000 * (1 + (0.2 * i) / 365));
		return dailySeries(values, '1000');
	};

	it('proyecta cinco años desde la rentabilidad anualizada', () => {
		const projection = buildGrowthProjection(oneYear(), summary({ currentValue: '1200' }))!;

		expect(projection.entries).toHaveLength(5);
		expect(projection.entries[0]).toEqual({ period: '2027', value: 1200, returnPct: 0 });
		expect(projection.entries.map((p) => p.period)).toEqual([
			'2027',
			'2028',
			'2029',
			'2030',
			'2031'
		]);
		expect(projection.entries[4].value).toBeGreaterThan(projection.entries[0].value);
	});

	// El importe depende de cuánto haya hoy en la cuenta; el porcentaje no, y es
	// lo único que la proyección de verdad extrapola.
	it('acumula el porcentaje componiendo la tasa anual, y arranca en cero', () => {
		const projection = buildGrowthProjection(oneYear(), summary({ currentValue: '1200' }))!;
		const rate = projection.annualRatePct / 100;

		expect(projection.annualRatePct).toBeCloseTo(20, 0);
		expect(projection.entries.map((p) => p.returnPct)).toEqual(
			[0, 1, 2, 3, 4].map((i) => (Math.pow(1 + rate, i) - 1) * 100)
		);
	});

	it('da el mismo porcentaje sea cual sea el saldo de la cuenta', () => {
		const small = buildGrowthProjection(oneYear(), summary({ currentValue: '1200' }))!;
		const large = buildGrowthProjection(oneYear(), summary({ currentValue: '1200000' }))!;

		expect(large.entries.map((p) => p.returnPct)).toEqual(small.entries.map((p) => p.returnPct));
		expect(large.entries[4].value).toBeGreaterThan(small.entries[4].value);
	});

	it('se abstiene con menos de medio año de historial', () => {
		const short = dailySeries(Array.from({ length: 100 }, (_, i) => 1000 + i));

		expect(buildGrowthProjection(short, summary())).toBeNull();
	});

	it('no proyecta el dinero aportado como si fuese rendimiento', () => {
		// Un año en el que el saldo se multiplica por diez, todo a base de aportes:
		// la rentabilidad es cero y la proyección queda plana.
		const values = Array.from({ length: 366 }, (_, i) => 1000 + i * 25);
		const funded = values.map((value, i) => {
			const day = new Date(Date.UTC(2026, 0, 1 + i)).toISOString().substring(0, 10);
			return point(day, String(value), String(value));
		});

		const projection = buildGrowthProjection(funded, summary({ currentValue: '10125' }))!;

		expect(projection.entries.map((p) => p.value)).toEqual(Array(5).fill(10125));
		expect(projection.entries.map((p) => p.returnPct)).toEqual(Array(5).fill(0));
		expect(projection.annualRatePct).toBe(0);
	});

	it('se abstiene con una tasa fuera de rango', () => {
		// x10 en un año: fuera del rango plausible para extrapolar.
		const values = Array.from({ length: 366 }, (_, i) => 1000 * (1 + (9 * i) / 365));

		expect(buildGrowthProjection(dailySeries(values, '1000'), summary())).toBeNull();
	});

	it('cae al último punto de la serie cuando el resumen no trae valor', () => {
		// El loader rellena el resumen ausente con ceros; la serie sigue siendo buena.
		const projection = buildGrowthProjection(oneYear(), summary({ currentValue: '0' }))!;

		expect(projection.entries[0].value).toBe(1200);
	});

	it('se abstiene cuando ni el resumen ni la serie dan un valor positivo', () => {
		const emptied = [...oneYear().slice(0, -1), point('2027-01-01', '0', '1000')];

		expect(buildGrowthProjection(emptied, summary({ currentValue: '0' }))).toBeNull();
	});

	it('se abstiene sin puntos', () => {
		expect(buildGrowthProjection([], summary())).toBeNull();
	});
});

describe('projectionCoordinates', () => {
	it('reparte los puntos en el eje x y estira los valores al alto del viewBox', () => {
		const coords = projectionCoordinates([
			{ period: '2026', value: 100, returnPct: 0 },
			{ period: '2027', value: 150, returnPct: 50 },
			{ period: '2028', value: 200, returnPct: 100 }
		]);

		expect(coords.map((c) => c.x)).toEqual([58, 182, 306]);
		// El mínimo se apoya en la base y el máximo llega al techo del área útil.
		expect(coords[0].y).toBe(230);
		expect(coords[2].y).toBe(50);
		expect(coords[1].y).toBe(140);
	});

	it('no divide por cero cuando todos los valores son iguales', () => {
		const coords = projectionCoordinates([
			{ period: '2026', value: 100, returnPct: 0 },
			{ period: '2027', value: 100, returnPct: 0 }
		]);

		expect(coords.every((c) => c.y === 230)).toBe(true);
	});

	it('devuelve una lista vacía sin proyección', () => {
		expect(projectionCoordinates([])).toEqual([]);
	});
});
