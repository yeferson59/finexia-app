import { describe, it, expect } from 'vitest';
import { formatCurrency } from './money';

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
