import { describe, it, expect } from 'vitest';
import { formatCurrency, currencySymbol } from './money';

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
