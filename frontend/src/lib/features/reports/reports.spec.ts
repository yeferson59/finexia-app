import { describe, it, expect } from 'vitest';
import {
	UNAVAILABLE,
	buildKeyStatistics,
	buildPerformanceCalendars,
	performanceClass
} from './reports';
import { periodReturns } from '$lib/shared/finance/returns';
import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

/**
 * Punto de la serie. `cost` es el capital invertido a esa fecha y `netFlow` el
 * dinero que el dueño movió desde el punto anterior. Sin `netFlow` el cálculo
 * cae al respaldo —la variación del coste—, que es el camino del backend
 * anterior y también se prueba.
 */
function point(date: string, totalValue: string, cost = '0', netFlow?: string): GrowthDataPoint {
	const gainLoss = String(Number(totalValue) - Number(cost));
	return {
		date,
		totalValue,
		totalCostBase: cost,
		gainLoss,
		gainLossPct: cost === '0' ? '0' : String((Number(gainLoss) / Number(cost)) * 100),
		...(netFlow === undefined ? {} : { netFlow })
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

/** Todas las métricas de todos los bloques, aplanadas para buscar por etiqueta. */
function statOf(groups: ReturnType<typeof buildKeyStatistics>, label: string) {
	const stat = groups.flatMap((group) => group.stats).find((s) => s.label === label);
	if (!stat) throw new Error(`no hay métrica «${label}»`);
	return stat;
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

describe('periodReturns', () => {
	it('no cuenta un aporte como rentabilidad', () => {
		// El valor se dobla, pero solo porque entró el mismo dinero de más.
		const returns = periodReturns([
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '2000', '2000')
		]);

		expect(returns).toHaveLength(1);
		expect(returns[0].value).toBeCloseTo(0, 10);
	});

	it('mide el movimiento de mercado que sí hubo sobre el aporte', () => {
		// Aporte de 1000 a mitad de tramo y 30 de revalorización encima.
		const [entry] = periodReturns([
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '2030', '2000')
		]);

		// Dietz modificada: 30 / (1000 + 1000/2).
		expect(entry.value).toBeCloseTo(30 / 1500, 10);
	});

	it('descuenta un retiro igual que un aporte', () => {
		const [entry] = periodReturns([
			point('2026-01-01', '2000', '2000'),
			point('2026-01-02', '1000', '1000')
		]);

		expect(entry.value).toBeCloseTo(0, 10);
	});

	it('lleva la fecha de cierre y los días de cada tramo', () => {
		const returns = periodReturns([
			point('2026-01-01', '1000', '1000'),
			point('2026-01-08', '1100', '1000')
		]);

		expect(returns[0]).toMatchObject({ date: '2026-01-08', days: 7 });
	});

	it('salta un tramo cuyo capital de partida no es positivo', () => {
		// Primera valoración de una cuenta vacía: no hay base sobre la que rendir.
		expect(periodReturns([point('2026-01-01', '0', '0'), point('2026-01-02', '0', '0')])).toEqual(
			[]
		);
	});

	it('devuelve una lista vacía con un solo punto', () => {
		expect(periodReturns([point('2026-01-01', '1000', '1000')])).toEqual([]);
	});

	it('acredita una venta con plusvalía en vez de contarla como pérdida', () => {
		// Acciones compradas por 600 se venden por 1000: el valor baja por lo
		// cobrado y el flujo también, así que el día queda plano.
		const [entry] = periodReturns([
			point('2026-01-01', '2000', '1600'),
			point('2026-01-02', '1000', '1000', '-1000')
		]);

		expect(entry.value).toBeCloseTo(0, 10);
	});

	it('sin `netFlow` cae al coste, y ahí la venta sí resta', () => {
		// El mismo caso contra un backend que no publica el flujo: la variación
		// del coste es -600, no -1000, y la plusvalía realizada aparece como
		// caída. Queda probado para que se sepa qué se pierde sin el campo.
		const [entry] = periodReturns([
			point('2026-01-01', '2000', '1600'),
			point('2026-01-02', '1000', '1000')
		]);

		expect(entry.value).toBeLessThan(0);
	});

	it('acredita un dividendo como renta cobrada', () => {
		// El valor no se mueve —el dividendo sale de lo medido— y el flujo
		// negativo es lo que lo convierte en rentabilidad.
		const [entry] = periodReturns([
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '1000', '1000', '-50')
		]);

		expect(entry.value).toBeCloseTo(50 / 975, 10);
	});

	it('el flujo del backend manda sobre la variación del coste', () => {
		// Coste plano pero flujo declarado: gana el flujo.
		const [entry] = periodReturns([
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '2000', '1000', '1000')
		]);

		expect(entry.value).toBeCloseTo(0, 10);
	});
});

describe('buildPerformanceCalendars', () => {
	it('encadena los tramos de cada mes', () => {
		const [calendar] = buildPerformanceCalendars([
			point('2026-01-31', '1000', '1000'),
			point('2026-02-14', '1100', '1000'),
			point('2026-02-28', '1210', '1000')
		]);

		expect(calendar.year).toBe('2026');
		// +10 % encadenado con +10 % es +21 %, no +20 %.
		expect(calendar.values[1]).toBe(21);
		// Enero no tiene tramo propio: su punto solo abre el historial.
		expect(calendar.values[0]).toBeNull();
	});

	it('no pinta un mes en verde por un depósito', () => {
		const [calendar] = buildPerformanceCalendars([
			point('2026-01-31', '1000', '1000'),
			point('2026-02-28', '5000', '5000')
		]);

		expect(calendar.values[1]).toBe(0);
	});

	it('marca como parcial el mes en el que empieza el historial', () => {
		const [calendar] = buildPerformanceCalendars([
			point('2026-07-20', '1000', '1000'),
			point('2026-07-31', '1100', '1000')
		]);

		expect(calendar.partialMonths).toEqual([6]);
	});

	it('marca como parcial el mes en curso, que tampoco está entero', () => {
		// El historial se corta el día 12: agosto lleva doce días, no treinta y uno.
		const [calendar] = buildPerformanceCalendars([
			point('2026-06-30', '1000', '1000'),
			point('2026-07-31', '1100', '1000'),
			point('2026-08-12', '1150', '1000')
		]);

		expect(calendar.partialMonths).toEqual([7]);
	});

	it('no marca parcial un mes que arranca desde el cierre del anterior', () => {
		const [calendar] = buildPerformanceCalendars([
			point('2026-01-31', '1000', '1000'),
			point('2026-02-28', '1100', '1000')
		]);

		expect(calendar.partialMonths).toEqual([]);
	});

	it('ordena los años del más reciente al más antiguo', () => {
		const calendars = buildPerformanceCalendars([
			point('2025-11-30', '1000', '1000'),
			point('2025-12-31', '1100', '1000'),
			point('2026-01-31', '1210', '1000')
		]);

		expect(calendars.map((c) => c.year)).toEqual(['2026', '2025']);
	});

	it('devuelve una lista vacía sin historial', () => {
		expect(buildPerformanceCalendars([])).toEqual([]);
	});
});

describe('buildKeyStatistics', () => {
	it('reparte las métricas en rendimiento, riesgo e historial', () => {
		const groups = buildKeyStatistics(dailySeries([1000, 1010, 1020]), summary());

		expect(groups.map((g) => g.title)).toEqual(['Rendimiento', 'Riesgo', 'Historial']);
	});

	it('mide la mayor caída sobre la rentabilidad, no sobre el saldo', () => {
		const groups = buildKeyStatistics(
			[
				point('2026-01-01', '1000', '1000'),
				point('2026-01-02', '1200', '1000'),
				point('2026-01-03', '900', '1000')
			],
			summary()
		);

		expect(statOf(groups, 'Máxima caída').value).toBe('-25,0%');
	});

	it('no llama caída a un retiro', () => {
		const groups = buildKeyStatistics(
			[
				point('2026-01-01', '2000', '2000'),
				point('2026-01-02', '1000', '1000'),
				point('2026-01-03', '1000', '1000')
			],
			summary()
		);

		expect(statOf(groups, 'Máxima caída').value).toBe('0,0%');
	});

	it('deja el riesgo en N/A con poco historial y dice qué falta', () => {
		const groups = buildKeyStatistics(dailySeries([1000, 1010, 1020]), summary());
		// Dos tramos no dan ni para medir la oscilación, así que la volatilidad no
		// sale ni siquiera sin anualizar.
		const volatility = statOf(groups, 'Volatilidad por tramo');

		expect(volatility.value).toBe(UNAVAILABLE);
		expect(volatility.hint).toMatch(/10 tramos de historial; llevas 2\./);
		expect(statOf(groups, 'Ratio de Sharpe').value).toBe(UNAVAILABLE);
		expect(statOf(groups, 'Ratio de Sharpe').hint).toMatch(/llevas 2 y 2\./);
	});

	it('calcula volatilidad y Sharpe con una serie diaria suficiente', () => {
		// Ciento veinte días alternando: pasa el trimestre que piden las cifras
		// anuales y hay varianza que medir.
		const values = Array.from({ length: 120 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const groups = buildKeyStatistics(dailySeries(values), summary());

		expect(statOf(groups, 'Volatilidad anualizada').value).not.toBe(UNAVAILABLE);
		expect(statOf(groups, 'Volatilidad anualizada').value).toMatch(/^\d+(\.\d+)*,\d%$/);
		expect(statOf(groups, 'Ratio de Sharpe').value).not.toBe(UNAVAILABLE);
	});

	it('no publica ninguna cifra anual por debajo del trimestre', () => {
		// Sesenta días y sesenta puntos: tramos de sobra, historial no. Las dos
		// cifras anuales se callan; publicar el Sharpe mientras se ocultaba la
		// rentabilidad anualizada era enseñar una derivada de un número que se
		// decía no tener.
		const values = Array.from({ length: 60 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const groups = buildKeyStatistics(dailySeries(values), summary());

		for (const label of ['Rentabilidad anualizada', 'Ratio de Sharpe']) {
			expect(statOf(groups, label).value).toBe(UNAVAILABLE);
			expect(statOf(groups, label).hint).toMatch(/90 días/);
		}

		// Lo que no se anualiza sí sale: el periodo entero y su peor caída.
		expect(statOf(groups, 'Rentabilidad del periodo').value).not.toBe(UNAVAILABLE);
		expect(statOf(groups, 'Máxima caída').value).not.toBe(UNAVAILABLE);
	});

	it('publica la volatilidad sin anualizar mientras no llegue al trimestre', () => {
		// La dispersión de los tramos converge mucho antes que una media: se mide
		// con sesenta días, solo que sin el √tramos, y la etiqueta lo dice.
		const values = Array.from({ length: 60 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const groups = buildKeyStatistics(dailySeries(values), summary());
		const volatility = statOf(groups, 'Volatilidad por tramo');

		expect(volatility.value).not.toBe(UNAVAILABLE);
		expect(volatility.note).toMatch(/Sin anualizar/);
		// Y es la desviación cruda, no la anualizada bajo otro nombre: esta serie
		// oscila menos de un punto de un día al otro.
		expect(Number.parseFloat(volatility.value.replace(',', '.'))).toBeLessThan(5);
	});

	it('anualiza la volatilidad y le quita la nota en cuanto pasa el trimestre', () => {
		const values = Array.from({ length: 120 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const volatility = statOf(
			buildKeyStatistics(dailySeries(values), summary()),
			'Volatilidad anualizada'
		);

		expect(volatility.value).not.toBe(UNAVAILABLE);
		expect(volatility.note).toBeUndefined();
	});

	it('no pinta el Sharpe en verde y le pone el reparo al lado', () => {
		const values = Array.from({ length: 120 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const sharpe = statOf(buildKeyStatistics(dailySeries(values), summary()), 'Ratio de Sharpe');

		expect(sharpe.tone).toBe('neutral');
		expect(sharpe.note).toMatch(/margen de error/);
	});

	it('no anualiza por debajo de un trimestre de historial', () => {
		const annualized = statOf(
			buildKeyStatistics(dailySeries(Array.from({ length: 30 }, (_, i) => 1000 + i)), summary()),
			'Rentabilidad anualizada'
		);

		expect(annualized.value).toBe(UNAVAILABLE);
		expect(annualized.hint).toMatch(/90 días/);
	});

	it('anualiza en cuanto el historial da', () => {
		const values = Array.from({ length: 120 }, (_, i) => 1000 + i);
		const annualized = statOf(
			buildKeyStatistics(dailySeries(values), summary()),
			'Rentabilidad anualizada'
		);

		expect(annualized.value).not.toBe(UNAVAILABLE);
		expect(annualized.tone).toBe('up');
	});

	it('separa la rentabilidad del periodo del dinero aportado', () => {
		const groups = buildKeyStatistics(
			[
				point('2026-01-01', '1000', '1000'),
				point('2026-01-02', '1100', '1000'),
				point('2026-01-03', '5100', '5000')
			],
			summary()
		);

		// +10 % y luego un depósito de 4000 que no mueve la rentabilidad.
		expect(statOf(groups, 'Rentabilidad del periodo').value).toBe('+10,0%');
	});

	it('publica los importes en la moneda del resumen', () => {
		const groups = buildKeyStatistics(
			[point('2026-01-01', '1000', '1000'), point('2026-01-02', '1200', '1000')],
			summary({ currency: 'USD', currentValue: '1200', gainLoss: '200', gainLossPct: '20' })
		);

		expect(statOf(groups, 'Valor actual').value).toBe('$1,200.00');
		expect(statOf(groups, 'Capital invertido').value).toBe('$1,000.00');
		expect(statOf(groups, 'Ganancia / pérdida').value).toBe('$200.00');
		expect(statOf(groups, 'Ganancia sobre coste').value).toBe('+20,0%');
	});

	it('saca la ganancia del último punto cuando el resumen no la trae', () => {
		const series = [point('2026-01-01', '1000', '1000'), point('2026-01-02', '1200', '1000')];

		for (const missing of [undefined, '']) {
			const groups = buildKeyStatistics(
				series,
				summary({ currentValue: '1200', gainLoss: missing, gainLossPct: missing })
			);

			expect(statOf(groups, 'Ganancia / pérdida').value).toBe('$200.00');
			expect(statOf(groups, 'Ganancia sobre coste').value).toBe('+20,0%');
		}
	});

	it('nombra el mejor y el peor mes', () => {
		const groups = buildKeyStatistics(
			[
				point('2026-01-31', '1000', '1000'),
				point('2026-02-28', '1100', '1000'),
				point('2026-03-31', '990', '1000')
			],
			summary()
		);

		expect(statOf(groups, 'Mejor mes').value).toBe('+10,0% · Feb 2026');
		expect(statOf(groups, 'Peor mes').value).toBe('-10,0% · Mar 2026');
	});

	it('deja fuera del mejor y el peor mes los que no están enteros', () => {
		// Junio arranca el 28 y rinde un +20 % en dos días; agosto, entero, un
		// +5 %. El mejor mes es agosto: dos días no compiten con treinta y uno.
		const groups = buildKeyStatistics(
			[
				point('2026-06-28', '1000', '1000'),
				point('2026-06-30', '1200', '1000'),
				point('2026-07-31', '1140', '1000'),
				point('2026-08-31', '1197', '1000')
			],
			summary()
		);

		expect(statOf(groups, 'Mejor mes').value).toBe('+5,0% · Ago 2026');
		expect(statOf(groups, 'Peor mes').value).toBe('-5,0% · Jul 2026');
	});

	it('marca el mes cuando el historial no tiene ninguno entero', () => {
		// Diez días de un solo mes: no hay con qué comparar, así que se publica lo
		// que hay con el mismo asterisco que usa el calendario.
		const groups = buildKeyStatistics(
			[point('2026-06-10', '1000', '1000'), point('2026-06-20', '1100', '1000')],
			summary()
		);

		expect(statOf(groups, 'Mejor mes').value).toBe('+10,0% · Jun 2026*');
		expect(statOf(groups, 'Mejor mes').note).toBe('* Mes incompleto.');
	});

	it('resume el periodo cubierto', () => {
		const groups = buildKeyStatistics(dailySeries([1000, 1010, 1020]), summary());
		const period = statOf(groups, 'Periodo cubierto');

		expect(period.value).toMatch(/2026/);
		expect(period.hint).toMatch(/3 puntos/);
	});

	it('devuelve una lista vacía sin historial', () => {
		expect(buildKeyStatistics([], summary())).toEqual([]);
	});
});
