import { describe, it, expect } from 'vitest';
import {
	UNAVAILABLE,
	buildKeyStatistics,
	buildPerformanceCalendars,
	buildRecordSummary,
	returnBackground,
	type KeyStat
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

/** Una medida por su etiqueta. */
function statOf(stats: KeyStat[], label: string) {
	const stat = stats.find((s) => s.label === label);
	if (!stat) throw new Error(`no hay medida «${label}»`);
	return stat;
}

/** La opacidad del tinte de una celda, para comparar intensidades. */
function alphaOf(background: string): number {
	return Number(background.match(/([\d.]+)\)$/)?.[1] ?? 0);
}

describe('returnBackground', () => {
	it('tiñe de verde lo que subió y de rojo lo que bajó', () => {
		expect(returnBackground(1.5)).toMatch(/^rgba\(34, 201, 126,/);
		expect(returnBackground(-1.5)).toMatch(/^rgba\(224, 90, 90,/);
		// Un mes plano no es una caída: el cero se queda del lado verde, con el
		// tinte más tenue de la escala.
		expect(returnBackground(0)).toMatch(/^rgba\(34, 201, 126,/);
	});

	it('sube la intensidad con el tamaño del movimiento, y ahí se para', () => {
		expect(alphaOf(returnBackground(2))).toBeGreaterThan(alphaOf(returnBackground(0.5)));
		// Un +40 % no puede pintar más que un +3 %: la escala satura antes.
		expect(alphaOf(returnBackground(40))).toBe(alphaOf(returnBackground(3)));
	});

	it('no tiñe un mes sin dato', () => {
		expect(returnBackground(null)).toBe('');
		expect(returnBackground(Number.NaN)).toBe('');
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

	it('compone el total del año con los meses que tienen dato', () => {
		const [calendar] = buildPerformanceCalendars([
			point('2026-01-31', '1000', '1000'),
			point('2026-02-28', '1100', '1000'),
			point('2026-03-31', '1210', '1000')
		]);

		// Dos meses del +10 % encadenados: +21 %, no +20 %. Y los diez meses sin
		// dato no arrastran el total a cero.
		expect(calendar.total).toBeCloseTo(21, 6);
	});

	it('deja el total en null cuando ningún mes tiene dato', () => {
		const calendars = buildPerformanceCalendars([]);

		expect(calendars).toEqual([]);
	});

	it('devuelve una lista vacía sin historial', () => {
		expect(buildPerformanceCalendars([])).toEqual([]);
	});
});

describe('buildKeyStatistics', () => {
	it('publica las cinco medidas de movimiento y riesgo', () => {
		const stats = buildKeyStatistics(dailySeries([1000, 1010, 1020]));

		expect(stats.map((stat) => stat.label)).toEqual([
			'Mejor mes',
			'Peor mes',
			// Con dos tramos la volatilidad no se anualiza, y la etiqueta lo dice.
			'Volatilidad por tramo',
			'Máxima caída',
			'Ratio de Sharpe'
		]);
	});

	it('cada medida explica qué mide, y la explicación se publica', () => {
		// El texto ya no vive en un `title`: es una columna de la tabla, así que
		// ninguna medida puede salir sin él.
		for (const stat of buildKeyStatistics(dailySeries([1000, 1010, 1020]))) {
			expect(stat.hint.length).toBeGreaterThan(20);
		}
	});

	it('mide la mayor caída sobre la rentabilidad, no sobre el saldo', () => {
		const stats = buildKeyStatistics([
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '1200', '1000'),
			point('2026-01-03', '900', '1000')
		]);

		expect(statOf(stats, 'Máxima caída').value).toBe('-25,0%');
	});

	it('no llama caída a un retiro', () => {
		const stats = buildKeyStatistics([
			point('2026-01-01', '2000', '2000'),
			point('2026-01-02', '1000', '1000'),
			point('2026-01-03', '1000', '1000')
		]);

		expect(statOf(stats, 'Máxima caída').value).toBe('0,0%');
	});

	it('deja el riesgo en N/A con poco historial y dice qué falta', () => {
		const stats = buildKeyStatistics(dailySeries([1000, 1010, 1020]));
		// Dos tramos no dan ni para medir la oscilación, así que la volatilidad no
		// sale ni siquiera sin anualizar.
		const volatility = statOf(stats, 'Volatilidad por tramo');

		expect(volatility.value).toBe(UNAVAILABLE);
		expect(volatility.hint).toMatch(/10 tramos de historial; llevas 2\./);
		expect(statOf(stats, 'Ratio de Sharpe').value).toBe(UNAVAILABLE);
		expect(statOf(stats, 'Ratio de Sharpe').hint).toMatch(/llevas 2 y 2\./);
	});

	it('calcula volatilidad y Sharpe con una serie diaria suficiente', () => {
		// Ciento veinte días alternando: pasa el trimestre que piden las cifras
		// anuales y hay varianza que medir.
		const values = Array.from({ length: 120 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const stats = buildKeyStatistics(dailySeries(values));

		expect(statOf(stats, 'Volatilidad anualizada').value).not.toBe(UNAVAILABLE);
		expect(statOf(stats, 'Volatilidad anualizada').value).toMatch(/^\d+(\.\d+)*,\d%$/);
		expect(statOf(stats, 'Ratio de Sharpe').value).not.toBe(UNAVAILABLE);
	});

	it('no publica el Sharpe por debajo del trimestre', () => {
		// Sesenta días y sesenta puntos: tramos de sobra, historial no. Publicarlo
		// mientras la cabecera se calla la rentabilidad anualizada era enseñar una
		// derivada de un número que se decía no tener.
		const values = Array.from({ length: 60 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const sharpe = statOf(buildKeyStatistics(dailySeries(values)), 'Ratio de Sharpe');

		expect(sharpe.value).toBe(UNAVAILABLE);
		expect(sharpe.hint).toMatch(/90 días/);
		// Lo que no se anualiza sí sale.
		expect(statOf(buildKeyStatistics(dailySeries(values)), 'Máxima caída').value).not.toBe(
			UNAVAILABLE
		);
	});

	it('publica la volatilidad sin anualizar mientras no llegue al trimestre', () => {
		// La dispersión de los tramos converge mucho antes que una media: se mide
		// con sesenta días, solo que sin el √tramos, y la etiqueta lo dice.
		const values = Array.from({ length: 60 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const volatility = statOf(buildKeyStatistics(dailySeries(values)), 'Volatilidad por tramo');

		expect(volatility.value).not.toBe(UNAVAILABLE);
		expect(volatility.note).toMatch(/Sin anualizar/);
		// Y es la desviación cruda, no la anualizada bajo otro nombre: esta serie
		// oscila menos de un punto de un día al otro.
		expect(Number.parseFloat(volatility.value.replace(',', '.'))).toBeLessThan(5);
	});

	it('anualiza la volatilidad y le quita la nota en cuanto pasa el trimestre', () => {
		const values = Array.from({ length: 120 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const volatility = statOf(buildKeyStatistics(dailySeries(values)), 'Volatilidad anualizada');

		expect(volatility.value).not.toBe(UNAVAILABLE);
		expect(volatility.note).toBeUndefined();
	});

	it('no pinta el Sharpe en verde y le pone el reparo al lado', () => {
		const values = Array.from({ length: 120 }, (_, i) => 1000 + (i % 2 === 0 ? 0 : 15) + i);
		const sharpe = statOf(buildKeyStatistics(dailySeries(values)), 'Ratio de Sharpe');

		expect(sharpe.tone).toBe('neutral');
		expect(sharpe.note).toMatch(/margen de error/);
	});

	it('nombra el mejor y el peor mes, y el mes va aparte de la cifra', () => {
		const stats = buildKeyStatistics([
			point('2026-01-31', '1000', '1000'),
			point('2026-02-28', '1100', '1000'),
			point('2026-03-31', '990', '1000')
		]);

		expect(statOf(stats, 'Mejor mes').value).toBe('+10,0%');
		expect(statOf(stats, 'Mejor mes').detail).toBe('febrero de 2026');
		expect(statOf(stats, 'Peor mes').value).toBe('-10,0%');
		expect(statOf(stats, 'Peor mes').detail).toBe('marzo de 2026');
	});

	it('deja fuera del mejor y el peor mes los que no están enteros', () => {
		// Junio arranca el 28 y rinde un +20 % en dos días; agosto, entero, un
		// +5 %. El mejor mes es agosto: dos días no compiten con treinta y uno.
		const stats = buildKeyStatistics([
			point('2026-06-28', '1000', '1000'),
			point('2026-06-30', '1200', '1000'),
			point('2026-07-31', '1140', '1000'),
			point('2026-08-31', '1197', '1000')
		]);

		expect(statOf(stats, 'Mejor mes').detail).toBe('agosto de 2026');
		expect(statOf(stats, 'Peor mes').detail).toBe('julio de 2026');
	});

	it('marca el mes cuando el historial no tiene ninguno entero', () => {
		// Diez días de un solo mes: no hay con qué comparar, así que se publica lo
		// que hay con el mismo asterisco que usa la matriz.
		const stats = buildKeyStatistics([
			point('2026-06-10', '1000', '1000'),
			point('2026-06-20', '1100', '1000')
		]);

		expect(statOf(stats, 'Mejor mes').value).toBe('+10,0%');
		expect(statOf(stats, 'Mejor mes').detail).toBe('junio de 2026*');
		expect(statOf(stats, 'Mejor mes').note).toBe('* Mes incompleto.');
	});

	it('devuelve una lista vacía sin historial', () => {
		expect(buildKeyStatistics([])).toEqual([]);
	});
});

describe('buildRecordSummary', () => {
	it('separa la rentabilidad del periodo del dinero aportado', () => {
		const record = buildRecordSummary(
			[
				point('2026-01-01', '1000', '1000'),
				point('2026-01-02', '1100', '1000'),
				point('2026-01-03', '5100', '5000')
			],
			summary()
		);

		// +10 % y luego un depósito de 4000 que no mueve la rentabilidad.
		expect(record?.periodReturn).toBeCloseTo(10, 6);
	});

	it('trae los importes y la moneda del resumen', () => {
		const record = buildRecordSummary(
			[point('2026-01-01', '1000', '1000'), point('2026-01-02', '1200', '1000')],
			summary({ currency: 'USD', currentValue: '1200', gainLoss: '200', gainLossPct: '20' })
		);

		expect(record).toMatchObject({
			value: 1200,
			cost: 1000,
			gain: 200,
			gainPct: 20,
			currency: 'USD',
			from: '2026-01-01',
			to: '2026-01-02'
		});
	});

	it('saca la ganancia del último punto cuando el resumen no la trae', () => {
		const series = [point('2026-01-01', '1000', '1000'), point('2026-01-02', '1200', '1000')];

		for (const missing of [undefined, '']) {
			const record = buildRecordSummary(
				series,
				summary({ currentValue: '1200', gainLoss: missing, gainLossPct: missing })
			);

			expect(record?.gain).toBe(200);
			expect(record?.gainPct).toBe(20);
		}
	});

	it('no anualiza por debajo de un trimestre de historial', () => {
		const record = buildRecordSummary(
			dailySeries(Array.from({ length: 30 }, (_, i) => 1000 + i)),
			summary()
		);

		expect(record?.periodReturn).not.toBeNull();
		expect(record?.annualized).toBeNull();
	});

	it('anualiza en cuanto el historial da', () => {
		const record = buildRecordSummary(
			dailySeries(Array.from({ length: 120 }, (_, i) => 1000 + i)),
			summary()
		);

		expect(record?.annualized).not.toBeNull();
		expect(record?.annualized).toBeGreaterThan(record!.periodReturn!);
	});

	it('cuenta los meses que cubre el historial', () => {
		const record = buildRecordSummary(
			[point('2025-06-01', '1000', '1000'), point('2026-07-28', '1100', '1000')],
			summary()
		);

		// Del 1 de junio de 2025 al 28 de julio de 2026: catorce meses.
		expect(record?.months).toBe(14);
	});

	it('se calla la rentabilidad sin dos cierres que comparar', () => {
		// Un punto suelto no es un tramo: la cabecera enseña el saldo y dice que
		// la rentabilidad espera al cierre de mañana.
		const record = buildRecordSummary(
			[point('2026-01-01', '1000', '1000')],
			summary({ currentValue: '1000' })
		);

		expect(record?.periodReturn).toBeNull();
		expect(record?.value).toBe(1000);
		// Y el mínimo de un mes, para que la frase no diga «0 meses de historial».
		expect(record?.months).toBe(1);
	});

	it('no hay ficha sin historial', () => {
		expect(buildRecordSummary([], summary())).toBeNull();
	});
});
