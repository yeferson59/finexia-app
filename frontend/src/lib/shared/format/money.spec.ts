import { describe, it, expect } from 'vitest';
import { formatCurrency, currencySymbol, formatCompactCurrency } from './money';

describe('formatCurrency', () => {
	it('formats USD with two decimals and the dollar symbol', () => {
		expect(formatCurrency(1234.5, 'USD')).toBe('$1,234.50');
	});

	it('formats COP with no decimals, grouping by dots', () => {
		const formatted = formatCurrency(1234567, 'COP');
		expect(formatted).toContain('$');
		expect(formatted).toContain('1.234.567');
		expect(formatted).not.toContain(',');
	});

	it('falls back to es-CO formatting for an unmapped currency code', () => {
		const formatted = formatCurrency(10, 'EUR');
		expect(formatted).toContain('EUR');
		expect(formatted).toContain('10,00');
	});
});

describe('currencySymbol', () => {
	// Mismo criterio que formatCurrency: USD se pinta con el dólar de en-US y
	// una moneda sin locale propio cae a es-CO, que para el euro escribe el
	// código en vez del signo.
	it('uses the same locale table as formatCurrency', () => {
		expect(currencySymbol('USD')).toBe('$');
		expect(currencySymbol('EUR')).toBe('EUR');
	});

	it('normalises the code before looking it up', () => {
		expect(currencySymbol(' usd ')).toBe('$');
	});

	// Un código roto no debe tumbar el gráfico que lo pinta.
	it('falls back to the dollar sign instead of throwing on a bad code', () => {
		expect(currencySymbol('')).toBe('$');
		expect(currencySymbol('not-a-currency')).toBe('$');
	});
});

describe('formatCompactCurrency', () => {
	// Entre 1.000 y 10.000 hace falta el decimal: redondeando a miles, un eje
	// que iba de 1.500 a 1.900 pintaba «$2k» en las cinco marcas.
	it('keeps a decimal between a thousand and ten thousand', () => {
		expect(formatCompactCurrency(1500, 'USD')).toBe('$1.5k');
		expect(formatCompactCurrency(1900, 'USD')).toBe('$1.9k');
	});

	it('drops it above ten thousand, where it no longer distinguishes anything', () => {
		expect(formatCompactCurrency(89406.1, 'USD')).toBe('$89k');
	});

	it('switches to millions when the amount reaches them', () => {
		expect(formatCompactCurrency(1_450_000, 'USD')).toBe('$1.4M');
	});

	it('leaves small amounts whole', () => {
		expect(formatCompactCurrency(240, 'USD')).toBe('$240');
	});

	it('carries the symbol of the currency it is given', () => {
		expect(formatCompactCurrency(12000, 'COP')).toBe('$12k');
		expect(formatCompactCurrency(12000, 'EUR')).toBe('EUR12k');
	});

	// Una pérdida es un importe como cualquier otro y conserva su signo.
	it('keeps the sign of a negative amount', () => {
		expect(formatCompactCurrency(-15000, 'USD')).toBe('$-15k');
	});
});

describe('formatCurrency con maxDigits', () => {
	/*
	 * El interés de una cuenta se cotiza a 0,0021 por dólar: con dos decimales
	 * la ficha del activo escribía «$0.00» en la misma fila que su total de
	 * $19.95, y la fila dejaba de cuadrar.
	 */
	it('keeps a sub-cent unit price readable', () => {
		expect(formatCurrency(0.0021, 'USD', 6)).toBe('$0.0021');
	});

	// El mínimo no se mueve, así que un importe normal no arrastra ceros.
	it('does not pad an ordinary amount with the extra digits', () => {
		expect(formatCurrency(1234.5, 'USD', 6)).toBe('$1,234.50');
	});

	// Un techo por debajo del mínimo de la moneda no puede recortarla.
	it('never drops below the currency minimum', () => {
		expect(formatCurrency(1234.5, 'USD', 0)).toBe('$1,234.50');
	});
});
