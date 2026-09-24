import { describe, expect, it } from 'vitest';
import {
	fundBalanceCreateSchema,
	fundContributionSchema,
	fundCreateSchema,
	fundErrorMessage,
	fundMarkErrorMessage,
	fundMarkSchema,
	fundMovementErrorMessage,
	fundWithdrawalSchema,
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

describe('fundBalanceCreateSchema', () => {
	const balanceFund = (over: Record<string, unknown> = {}) => ({
		portfolioId: ID,
		sourceId: ID,
		currency: 'COP',
		name: 'Bolsillo de inversión',
		date: '2026-01-10',
		amount: '12000000',
		currentBalance: '12640000',
		currentDate: '2026-09-22',
		...over
	});

	it('accepts what went in and what it holds now', () => {
		const parsed = fundBalanceCreateSchema.safeParse(balanceFund());

		expect(parsed.success).toBe(true);
		expect(parsed.data?.amount).toBe(12000000);
		expect(parsed.data?.currentBalance).toBe(12640000);
	});

	it('leaves the current balance out when it is blank', () => {
		expect(
			fundBalanceCreateSchema.safeParse(balanceFund({ currentBalance: '' })).data?.currentBalance
		).toBeUndefined();
	});

	it('refuses a balance from before the first contribution', () => {
		expect(
			fundBalanceCreateSchema.safeParse(balanceFund({ currentDate: '2026-01-09' })).success
		).toBe(false);
	});
});

describe('fundMarkSchema with a balance', () => {
	it('takes a balance instead of a unit value', () => {
		const parsed = fundMarkSchema.safeParse({ id: ID, date: '2026-09-30', balance: '13050000' });

		expect(parsed.success).toBe(true);
		expect(parsed.data?.balance).toBe(13050000);
		expect(parsed.data?.unitValue).toBeUndefined();
	});

	it('takes one of the two, never both or neither', () => {
		expect(
			fundMarkSchema.safeParse({ id: ID, date: '2026-09-30', unitValue: '1', balance: '1' }).success
		).toBe(false);
		expect(fundMarkSchema.safeParse({ id: ID, date: '2026-09-30' }).success).toBe(false);
	});
});

describe('movements', () => {
	it('accepts a contribution with an optional balance before', () => {
		const parsed = fundContributionSchema.safeParse({
			id: ID,
			portfolioId: ID,
			sourceId: ID,
			date: '2026-08-15',
			amount: '5000000',
			balanceBefore: ''
		});

		expect(parsed.success).toBe(true);
		expect(parsed.data?.balanceBefore).toBeUndefined();
	});

	it('reads the «all» checkbox and keeps the fees under the amount', () => {
		const withdrawal = (over: Record<string, unknown> = {}) => ({
			id: ID,
			entryId: ID,
			date: '2026-09-20',
			amount: '2000000',
			fees: 0,
			...over
		});

		expect(fundWithdrawalSchema.safeParse(withdrawal({ all: 'on' })).data?.all).toBe(true);
		expect(fundWithdrawalSchema.safeParse(withdrawal()).data?.all).toBe(false);
		expect(fundWithdrawalSchema.safeParse(withdrawal({ fees: '2000000' })).success).toBe(false);
	});

	it('explains a withdrawal larger than the position', () => {
		const message = fundMovementErrorMessage(
			409,
			'the position does not hold that much of the fund: on 2026-08-05 it held 1003.00 at 100.3'
		);

		expect(message).toContain('1003.00');
		expect(message).toContain('Retirar todo');
	});

	it('explains a balance before the first contribution', () => {
		expect(
			fundMovementErrorMessage(
				409,
				'the fund held no units on that day: leave out the balance before the first contribution'
			)
		).toContain('deja vacío');
	});
});
