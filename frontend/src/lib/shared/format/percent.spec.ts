import { describe, it, expect } from 'vitest';
import { formatPercent, formatSignedPercent } from './percent';

describe('formatPercent', () => {
	it('escribe la coma decimal de es-CO', () => {
		expect(formatPercent(12.345)).toBe('12,3%');
		expect(formatPercent(12.345, 2)).toBe('12,35%');
	});

	it('separa los miles', () => {
		expect(formatPercent(1234.5)).toBe('1.234,5%');
	});

	// Una cifra que redondea a cero salía como «-0,0%», que se lee como una
	// pérdida que no existe.
	it('normaliza a cero lo que redondea a cero', () => {
		expect(formatPercent(-0.02)).toBe('0,0%');
		expect(formatPercent(0.02)).toBe('0,0%');
		expect(formatPercent(-0.4, 0)).toBe('0%');
	});

	it('conserva el signo de lo que no redondea a cero', () => {
		expect(formatPercent(-3.2)).toBe('-3,2%');
	});

	it('no escribe «NaN%» cuando no hay cifra', () => {
		expect(formatPercent(Number.NaN)).toBe('—');
		expect(formatSignedPercent(Number.POSITIVE_INFINITY)).toBe('—');
	});
});

describe('formatSignedPercent', () => {
	it('marca la ganancia con un más', () => {
		expect(formatSignedPercent(8.44)).toBe('+8,4%');
	});

	it('deja la pérdida con su menos', () => {
		expect(formatSignedPercent(-8.44)).toBe('-8,4%');
	});

	// Un «+0,0%» promete una ganancia que la propia cifra desmiente.
	it('no firma la banda que redondea a cero', () => {
		expect(formatSignedPercent(0.03)).toBe('0,0%');
		expect(formatSignedPercent(0)).toBe('0,0%');
	});
});
