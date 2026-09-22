import { describe, expect, it } from 'vitest';
import {
	FORECAST_MONTHS,
	buildValueForecast,
	completeMonthlyReturns,
	forecastCurve,
	forecastMilestones
} from './projection';
import type { GrowthDataPoint } from '$lib/api/types';

/**
 * Un punto de la serie. `netFlow` es lo que entró o salió desde el anterior;
 * sin él la aritmética se cae al respaldo del coste, que aquí no se usa.
 */
const point = (
	date: string,
	totalValue: number,
	totalCostBase: number,
	netFlow = 0
): GrowthDataPoint => ({
	date,
	totalValue: String(totalValue),
	totalCostBase: String(totalCostBase),
	gainLoss: String(totalValue - totalCostBase),
	gainLossPct: '0',
	netFlow: String(netFlow)
});

/**
 * Serie mensual limpia: un punto el último día de cada mes, sin aportes, con
 * el valor multiplicándose por `1 + retorno` cada mes.
 *
 * `returns[0]` es el de enero y **no cuenta**: el mes en que arranca la serie
 * se considera siempre parcial, porque su retorno corre desde el día que abrió
 * el historial. Así que una serie de n retornos deja n − 1 meses completos.
 */
function monthlySeries(returns: number[]): GrowthDataPoint[] {
	const ends = ['01-31', '02-28', '03-31', '04-30', '05-31', '06-30', '07-31'];
	let value = 1000;
	const points = [point('2026-01-01', value, 1000)];

	returns.forEach((rate, i) => {
		value *= 1 + rate;
		points.push(point(`2026-${ends[i]}`, value, 1000));
	});

	return points;
}

describe('completeMonthlyReturns', () => {
	it('deja fuera el mes en que empieza la serie y el que está a medias', () => {
		const points = [
			point('2026-01-01', 1000, 1000),
			point('2026-01-31', 1100, 1000),
			point('2026-02-28', 1210, 1000),
			point('2026-03-31', 1331, 1000),
			point('2026-07-15', 1400, 1000)
		];

		const monthly = completeMonthlyReturns(points);

		// Enero queda fuera por ser el mes en que abre la serie y julio por estar
		// a medias: quedan febrero y marzo, los dos al 10 %.
		expect(monthly).toHaveLength(2);
		expect(monthly[0]).toBeCloseTo(0.1, 6);
		expect(monthly[1]).toBeCloseTo(0.1, 6);
	});

	it('no devuelve nada con un solo punto', () => {
		expect(completeMonthlyReturns([point('2026-01-01', 1000, 1000)])).toEqual([]);
	});
});

describe('buildValueForecast', () => {
	it('toma el mejor y el peor mes, no su media', () => {
		// Febrero +1 %, marzo +5 %, abril +3 %: los extremos son el 1 % y el 5 %,
		// y el 3 % de en medio no aparece por ninguna parte.
		const forecast = buildValueForecast(monthlySeries([0.02, 0.01, 0.05, 0.03]), 2000);

		expect(forecast).not.toBeNull();
		expect(forecast?.basisMonths).toBe(3);
		expect(forecast?.worstMonthPct).toBeCloseTo(1, 6);
		expect(forecast?.bestMonthPct).toBeCloseTo(5, 6);
		expect(forecast?.entries).toHaveLength(FORECAST_MONTHS);
	});

	it('compone cada extremo doce veces', () => {
		const forecast = buildValueForecast(monthlySeries([0, 0.01, 0.05]), 1000);
		const year = forecast?.entries[11];

		expect(year?.low).toBeCloseTo(1000 * 1.01 ** 12, 6);
		expect(year?.high).toBeCloseTo(1000 * 1.05 ** 12, 6);
		expect(year?.lowPct).toBeCloseTo((1.01 ** 12 - 1) * 100, 6);
		expect(forecast?.worstAnnualPct).toBeCloseTo((1.01 ** 12 - 1) * 100, 6);
		expect(forecast?.bestAnnualPct).toBeCloseTo((1.05 ** 12 - 1) * 100, 6);
	});

	/* Con un mes en negativo el suelo de la banda baja: es lo que pasaría si se
	   repitiera aquel mes, y esconderlo sería enseñar solo la mitad buena. */
	it('deja el suelo por debajo de hoy cuando hubo un mes malo', () => {
		const forecast = buildValueForecast(monthlySeries([0, 0.04, -0.03]), 1000);

		expect(forecast?.worstMonthPct).toBeCloseTo(-3, 6);
		expect(forecast?.entries[11].low).toBeLessThan(1000);
		expect(forecast?.entries[11].high).toBeGreaterThan(1000);
	});

	it('proyecta a fin de mes, doce meses hacia adelante', () => {
		const forecast = buildValueForecast(monthlySeries([0.01, 0.01, 0.01]), 1000);
		// La base son febrero y marzo; enero abre la serie y no cuenta.
		expect(forecast?.basisMonths).toBe(2);

		// La serie cierra el 31 de marzo: el primer mes proyectado es abril.
		expect(forecast?.entries[0].date).toBe('2026-04-30');
		expect(forecast?.entries[0].monthsAhead).toBe(1);
		expect(forecast?.entries[11].date).toBe('2027-03-31');
	});

	/*
	 * Lo que se extrapola es rentabilidad. Una serie cuyo valor se dispara solo
	 * porque su dueño ingresó dinero no rindió nada, y su banda tiene que ser
	 * plana por los dos lados.
	 */
	it('no confunde un aporte con una ganancia', () => {
		const points = [
			point('2026-01-01', 1000, 1000),
			point('2026-01-31', 1000, 1000),
			point('2026-02-28', 5000, 5000, 4000),
			point('2026-03-31', 5000, 5000)
		];

		const forecast = buildValueForecast(points, 5000);

		expect(forecast?.bestMonthPct).toBeCloseTo(0, 6);
		expect(forecast?.entries[11].high).toBeCloseTo(5000, 6);
		expect(forecast?.entries[11].low).toBeCloseTo(5000, 6);
	});

	/* Tres meses de historial son dos meses cerrados, que es justo el mínimo:
	   con menos, el mejor y el peor mes serían el mismo mes. */
	it('aparece con tres meses de historial y no antes', () => {
		expect(buildValueForecast(monthlySeries([0.02, 0.02]), 1000)).toBeNull();
		expect(buildValueForecast(monthlySeries([0.02, 0.02, 0.02]), 1000)).not.toBeNull();
	});

	it('se abstiene tras un mes que se lo llevó todo', () => {
		expect(buildValueForecast(monthlySeries([0, 0.02, -1]), 1000)).toBeNull();
	});

	it('sin valor de resumen arranca del último punto de la serie', () => {
		const series = monthlySeries([0.02, 0.02, 0.02]);
		const last = Number(series[series.length - 1].totalValue);

		expect(buildValueForecast(series, 0)?.entries[0].high).toBeCloseTo(last * 1.02, 6);
		expect(buildValueForecast([], 1000)).toBeNull();
	});
});

describe('forecastCurve', () => {
	/*
	 * El eje reparte los puntos por índice: si el futuro avanzara a otro paso que
	 * el pasado, un año de proyección ocuparía lo que unas semanas de historial.
	 */
	it('avanza al mismo paso que el historial', () => {
		const series = monthlySeries([0.01, 0.01, 0.01]);
		const curve = forecastCurve(series, buildValueForecast(series, 1000), 1000);

		// El historial va a un punto por mes, así que la curva de doce meses trae
		// del orden de doce puntos, no cientos ni tres.
		expect(curve.length).toBeGreaterThan(10);
		expect(curve.length).toBeLessThan(16);
	});

	it('abre la banda en el punto de hoy y cierra en las cifras del último mes', () => {
		const series = monthlySeries([0, 0.01, 0.05]);
		const forecast = buildValueForecast(series, 1000);
		const curve = forecastCurve(series, forecast, 1000);

		// Los dos bordes salen del mismo sitio: hoy la cartera vale una sola cosa.
		expect(curve[0].date).toBe('2026-03-31');
		expect(curve[0].low).toBeCloseTo(1000, 6);
		expect(curve[0].high).toBeCloseTo(1000, 6);

		// El final coincide con las cifras que la tabla de abajo publica para el
		// mes doce: la banda y los números no pueden decir cosas distintas.
		const year = forecast?.entries[FORECAST_MONTHS - 1];
		const end = curve[curve.length - 1];
		expect(end.date).toBe(year?.date);
		expect(end.low).toBeCloseTo(year?.low ?? 0, 6);
		expect(end.high).toBeCloseTo(year?.high ?? 0, 6);
	});

	it('no devuelve nada sin proyección', () => {
		expect(forecastCurve(monthlySeries([0.01, 0.01, 0.01]), null, 1000)).toEqual([]);
	});
});

describe('forecastMilestones', () => {
	it('destaca el trimestre, el semestre y el año', () => {
		const forecast = buildValueForecast(monthlySeries([0.01, 0.01, 0.01]), 1000);

		expect(forecastMilestones(forecast).map((e) => e.monthsAhead)).toEqual([3, 6, 12]);
	});

	it('no destaca nada sin proyección', () => {
		expect(forecastMilestones(null)).toEqual([]);
	});
});
