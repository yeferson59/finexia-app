import { describe, it, expect } from 'vitest';
import {
	annualize,
	cumulativeReturns,
	gainOverCost,
	spanDays,
	timeWeightedReturn,
	type ReturnSeriesPoint
} from './returns';

/** Punto de la serie; `netFlow` es el dinero que entró desde el punto anterior. */
function point(date: string, totalValue: string, cost = '0', netFlow?: string): ReturnSeriesPoint {
	return {
		date,
		totalValue,
		totalCostBase: cost,
		...(netFlow === undefined ? {} : { netFlow })
	};
}

describe('spanDays', () => {
	it('cuenta los días entre el primer punto y el último', () => {
		expect(spanDays([point('2026-01-01', '1'), point('2026-03-02', '1')])).toBe(60);
	});

	it('ordena antes de medir', () => {
		expect(spanDays([point('2026-03-02', '1'), point('2026-01-01', '1')])).toBe(60);
	});

	it('es cero sin dos puntos que comparar', () => {
		expect(spanDays([point('2026-01-01', '1')])).toBe(0);
		expect(spanDays([])).toBe(0);
	});
});

describe('timeWeightedReturn', () => {
	it('encadena los tramos del historial', () => {
		const series = [
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '1100', '1000'),
			point('2026-01-03', '1210', '1000')
		];

		expect(timeWeightedReturn(series)).toBeCloseTo(0.21, 10);
	});

	// El caso que motivó toda esta aritmética: la cuenta vale el doble porque le
	// metieron dinero, no porque haya ganado nada.
	it('no cuenta un aporte como rentabilidad', () => {
		const series = [
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '2000', '2000', '1000')
		];

		expect(timeWeightedReturn(series)).toBe(0);
	});

	it('es null mientras no haya un tramo que medir', () => {
		expect(timeWeightedReturn([point('2026-01-01', '1000', '1000')])).toBeNull();
		expect(timeWeightedReturn([])).toBeNull();
	});
});

describe('annualize', () => {
	it('deja igual lo que ya cubre un año', () => {
		expect(annualize(0.2, 365.25)).toBeCloseTo(0.2, 10);
	});

	it('compone medio año hasta el año entero', () => {
		expect(annualize(0.2, 365.25 / 2)).toBeCloseTo(0.44, 10);
	});

	// Una base negativa elevada a un exponente fraccionario no da un número real.
	it('se abstiene tras una pérdida total o sin días que cubrir', () => {
		expect(annualize(-1, 200)).toBeNull();
		expect(annualize(-1.5, 200)).toBeNull();
		expect(annualize(0.2, 0)).toBeNull();
		expect(annualize(Number.NaN, 200)).toBeNull();
	});
});

describe('gainOverCost', () => {
	it('divide el valor de mercado entre lo invertido', () => {
		expect(gainOverCost(point('2026-01-01', '1250', '1000'))).toBeCloseTo(0.25, 10);
	});

	// Dividir por cero no da un 0 %, da nada: es una cuenta vacía.
	it('es null sin capital invertido', () => {
		expect(gainOverCost(point('2026-01-01', '0', '0'))).toBeNull();
		expect(gainOverCost(point('2026-01-01', '100', ''))).toBeNull();
	});
});

describe('cumulativeReturns', () => {
	it('ancla la curva en cero el primer día', () => {
		const curve = cumulativeReturns([
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '1100', '1000')
		]);

		expect(curve[0].twr).toBe(0);
		expect(curve[1].twr).toBeCloseTo(0.1, 10);
	});

	it('devuelve un punto por fecha, en orden', () => {
		const curve = cumulativeReturns([
			point('2026-01-03', '1210', '1000'),
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '1100', '1000')
		]);

		expect(curve.map((p) => p.date)).toEqual(['2026-01-01', '2026-01-02', '2026-01-03']);
		expect(curve[2].twr).toBeCloseTo(0.21, 10);
	});

	// Las dos lecturas de la misma serie: la de la izquierda descuenta el aporte,
	// la de la derecha no puede porque solo mira el saldo de ese día.
	it('separa la rentabilidad de la ganancia sobre coste', () => {
		const curve = cumulativeReturns([
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '2100', '2000', '1000')
		]);

		expect(curve[1].twr).toBeCloseTo(100 / 1500, 10);
		expect(curve[1].overCost).toBeCloseTo(0.05, 10);
	});

	// Un tramo que la Dietz descarta no puede dejar un agujero en la línea.
	it('arrastra el índice por un tramo sin base con la que medir', () => {
		const curve = cumulativeReturns([
			point('2026-01-01', '0', '0'),
			point('2026-01-02', '1000', '1000', '1000'),
			point('2026-01-03', '1100', '1000')
		]);

		expect(curve[0].twr).toBe(0);
		expect(curve[1].twr).toBe(0);
		expect(curve[2].twr).toBeCloseTo(0.1, 10);
	});

	it('devuelve una lista vacía sin serie', () => {
		expect(cumulativeReturns([])).toEqual([]);
	});
});
