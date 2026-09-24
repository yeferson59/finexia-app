import { describe, expect, it } from 'vitest';
import {
	fundCreateSchema,
	fundErrorMessage,
	fundMarkErrorMessage,
	fundMarkSchema,
	toFundDateTime
} from './schemas';

const ID = '3f7c1d2e-4b5a-4c6d-8e9f-0a1b2c3d4e5f';

function fund(over: Record<string, unknown> = {}) {
	return {
		portfolioId: ID,
		sourceId: ID,
		currency: 'COP',
		name: ' FIC Renta Fija ',
		date: '2026-09-01',
		units: '1000',
		unitValue: '12345.678901',
		currentUnitValue: '',
		currentDate: '2026-09-30',
		...over
	};
}

describe('fundCreateSchema', () => {
	it('accepts a fund without a current value', () => {
		const parsed = fundCreateSchema.safeParse(fund());

		expect(parsed.success).toBe(true);
		expect(parsed.data?.name).toBe('FIC Renta Fija');
		expect(parsed.data?.units).toBe(1000);
		expect(parsed.data?.currentUnitValue).toBeUndefined();
	});

	it('accepts a current value on or after the purchase', () => {
		const parsed = fundCreateSchema.safeParse(fund({ currentUnitValue: '12431.22' }));

		expect(parsed.success).toBe(true);
		expect(parsed.data?.currentUnitValue).toBe(12431.22);
	});

	it.each([
		['no units', { units: '0' }],
		['no unit value', { unitValue: '' }],
		['a blank name', { name: '  ' }],
		['an unknown currency', { currency: 'XXX' }],
		['a current value before the purchase', { currentUnitValue: '1', currentDate: '2026-08-31' }],
		['a huge unit value', { unitValue: '1e12' }]
	])('rejects %s', (_, over) => {
		expect(fundCreateSchema.safeParse(fund(over)).success).toBe(false);
	});
});

describe('fundMarkSchema', () => {
	it('needs a positive unit value and a day', () => {
		expect(
			fundMarkSchema.safeParse({ id: ID, date: '2026-09-30', unitValue: '12431.22' }).success
		).toBe(true);
		expect(fundMarkSchema.safeParse({ id: ID, date: '2026-09-30', unitValue: '0' }).success).toBe(
			false
		);
		expect(fundMarkSchema.safeParse({ id: ID, date: '', unitValue: '1' }).success).toBe(false);
	});
});

describe('messages', () => {
	it('explains why a fund cannot be dropped', () => {
		expect(fundErrorMessage(409, 'the fund still has positions')).toContain('portafolio');
	});

	it('explains a future mark', () => {
		expect(fundMarkErrorMessage(400, 'invalid fund mark: date cannot be in the future')).toContain(
			'futura'
		);
	});

	it('falls back to a reload on a 404', () => {
		expect(fundMarkErrorMessage(404)).toContain('Recarga');
	});

	it('sends a day as midnight UTC', () => {
		expect(toFundDateTime('2026-09-30')).toBe('2026-09-30T00:00:00Z');
	});
});
