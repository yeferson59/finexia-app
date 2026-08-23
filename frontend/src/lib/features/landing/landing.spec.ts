import { describe, it, expect } from 'vitest';
import {
	countdownBetween,
	chartPoints,
	chartY,
	METRICS_MARKET_VALUE,
	METRICS_INVESTED
} from './landing';

const at = (iso: string) => new Date(iso).getTime();

describe('countdownBetween', () => {
	it('descompone el tiempo restante en días, horas, minutos y segundos', () => {
		expect(countdownBetween(at('2026-10-01T09:00:00Z'), at('2026-09-29T07:58:57Z'))).toEqual({
			days: '02',
			hours: '01',
			mins: '01',
			secs: '03'
		});
	});

	it('rellena a dos dígitos', () => {
		expect(countdownBetween(at('2026-10-01T09:00:09Z'), at('2026-10-01T09:00:00Z'))).toEqual({
			days: '00',
			hours: '00',
			mins: '00',
			secs: '09'
		});
	});

	it('se queda en ceros pasada la fecha, sin contar hacia atrás', () => {
		expect(countdownBetween(at('2026-10-01T09:00:00Z'), at('2026-12-01T00:00:00Z'))).toEqual({
			days: '00',
			hours: '00',
			mins: '00',
			secs: '00'
		});
	});
});

describe('chartY', () => {
	const scale = { min: 100, max: 200 };
	const box = { left: 0, right: 100, top: 0, bottom: 200 };

	it('lleva el mínimo al borde inferior y el máximo al superior', () => {
		expect(chartY(100, scale, box)).toBe(200);
		expect(chartY(200, scale, box)).toBe(0);
	});

	it('interpola los valores intermedios', () => {
		expect(chartY(150, scale, box)).toBe(100);
	});

	it('no divide por cero cuando la escala es plana', () => {
		expect(chartY(150, { min: 150, max: 150 }, box)).toBe(200);
	});
});

describe('chartPoints', () => {
	const scale = { min: 0, max: 10 };
	const box = { left: 10, right: 110, top: 0, bottom: 100 };

	it('reparte los valores a lo ancho del área de trazado', () => {
		expect(chartPoints([0, 5, 10], scale, box)).toBe('10,100 60,50 110,0');
	});

	it('deja un único valor en el borde izquierdo, sin dividir por cero', () => {
		expect(chartPoints([5], scale, box)).toBe('10,50');
	});
});

describe('series de la sección de métricas', () => {
	it('el valor de mercado nunca cae por debajo del capital invertido', () => {
		METRICS_MARKET_VALUE.forEach((value, i) => {
			expect(value).toBeGreaterThan(METRICS_INVESTED[i]);
		});
	});

	it('cierra en los $248.500 y $221.100 que anuncia la landing', () => {
		const value = METRICS_MARKET_VALUE.at(-1)!;
		const invested = METRICS_INVESTED.at(-1)!;
		expect(value).toBe(248.5);
		expect(invested).toBe(221.1);
		// Los $27.400 de ganancia son el 12,4% del capital invertido.
		expect(((value - invested) / invested) * 100).toBeCloseTo(12.4, 1);
	});
});
