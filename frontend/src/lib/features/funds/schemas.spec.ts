import { describe, expect, it } from 'vitest';
import {
	fundBalanceCreateSchema,
	fundContributionSchema,
	fundCreateSchema,
	fundDeleteSchema,
	fundErrorMessage,
	fundLinkErrorMessage,
	fundLinkSchema,
	fundMarkErrorMessage,
	fundMarkSchema,
	fundMarksBulkSchema,
	fundMovementEditSchema,
	fundMovementErrorMessage,
	fundUnitsContributionSchema,
	fundUnitsMovementEditSchema,
	fundUnitsWithdrawalSchema,
	fundUnlinkSchema,
	fundWithdrawalSchema,
	publicFundSearchErrorMessage,
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

describe('fundMarksBulkSchema', () => {
	it('reads the table the browser already parsed', () => {
		const parsed = fundMarksBulkSchema.safeParse({
			id: ID,
			tracking: 'units',
			marks: JSON.stringify([{ date: '2026-09-30', value: 12431.22 }])
		});

		expect(parsed.success).toBe(true);
		expect(parsed.data?.marks).toEqual([{ date: '2026-09-30', value: 12431.22 }]);
	});

	it('refuses an empty table, a broken one, or one too long', () => {
		const bulk = (marks: string) =>
			fundMarksBulkSchema.safeParse({ id: ID, tracking: 'balance', marks });
		const many = JSON.stringify(
			Array.from({ length: 401 }, () => ({ date: '2026-09-30', value: 1 }))
		);

		expect(bulk('[]').success).toBe(false);
		expect(bulk('{no').success).toBe(false);
		expect(bulk(many).success).toBe(false);
	});
});

describe('a fund linked to the Superfinanciera', () => {
	const PUBLIC_ID = '5-31-2852-1-800';

	it('may leave the unit value of the purchase to the published one', () => {
		const parsed = fundCreateSchema.safeParse(fund({ unitValue: '', publicFundId: PUBLIC_ID }));

		expect(parsed.success).toBe(true);
		expect(parsed.data?.unitValue).toBeUndefined();
		expect(parsed.data?.publicFundId).toBe(PUBLIC_ID);
	});

	it('still takes a unit value the owner wrote', () => {
		const parsed = fundCreateSchema.safeParse(fund({ publicFundId: PUBLIC_ID }));

		expect(parsed.data?.unitValue).toBe(12345.678901);
	});

	it.each([
		['a fund in dollars', { currency: 'USD', publicFundId: PUBLIC_ID }],
		['a malformed id', { publicFundId: '5-31-2852' }]
	])('rejects %s', (_, over) => {
		expect(fundCreateSchema.safeParse(fund(over)).success).toBe(false);
	});

	it('links and unlinks by id', () => {
		expect(fundLinkSchema.safeParse({ id: ID, publicFundId: PUBLIC_ID }).success).toBe(true);
		expect(fundLinkSchema.safeParse({ id: ID, publicFundId: '' }).success).toBe(false);
		expect(fundUnlinkSchema.safeParse({ id: ID }).success).toBe(true);
	});

	it('explains why a link failed', () => {
		expect(fundLinkErrorMessage(503)).toMatch(/no respondió/);
		expect(fundLinkErrorMessage(409)).toMatch(/por unidades, en pesos/);
		expect(fundLinkErrorMessage(404, 'public fund not found')).toMatch(/catálogo/);
		expect(fundErrorMessage(503)).toMatch(/no respondió/);
		expect(
			fundErrorMessage(
				400,
				'invalid fund: unitValue is required: the SFC published none for 2026-09-20'
			)
		).toMatch(/no publicó un valor/);
		expect(publicFundSearchErrorMessage(400)).toMatch(/dos letras/);
	});
});

describe('movements of a fund followed by units', () => {
	it('takes units and a unit value for a contribution', () => {
		const parsed = fundUnitsContributionSchema.safeParse({
			id: ID,
			portfolioId: ID,
			sourceId: ID,
			date: '2026-09-10',
			units: '150.5',
			unitValue: '12345.678901',
			notes: ' extracto '
		});

		expect(parsed.success).toBe(true);
		expect(parsed.data?.units).toBe(150.5);
		expect(parsed.data?.notes).toBe('extracto');
		expect(
			fundUnitsContributionSchema.safeParse({ ...parsed.data, id: ID, units: '0' }).success
		).toBe(false);
	});

	it('lets «Retirar todo» leave the units out, and nothing else', () => {
		const withdrawal = (over: Record<string, unknown> = {}) => ({
			id: ID,
			entryId: ID,
			date: '2026-09-20',
			units: '',
			unitValue: '12000',
			fees: '',
			...over
		});

		const all = fundUnitsWithdrawalSchema.safeParse(withdrawal({ all: 'on' }));
		expect(all.success).toBe(true);
		expect(all.data?.units).toBeUndefined();
		expect(all.data?.fees).toBe(0);

		expect(fundUnitsWithdrawalSchema.safeParse(withdrawal()).success).toBe(false);
		expect(fundUnitsWithdrawalSchema.safeParse(withdrawal({ units: '10' })).success).toBe(true);
		expect(
			fundUnitsWithdrawalSchema.safeParse(withdrawal({ units: '10', fees: '120000' })).success
		).toBe(false);
	});

	it('corrects a movement by units or by money', () => {
		const txn = { txnId: ID, date: '2026-09-12', notes: '' };

		expect(
			fundUnitsMovementEditSchema.safeParse({ ...txn, units: '10', unitValue: '100', fees: '' })
				.success
		).toBe(true);
		expect(
			fundUnitsMovementEditSchema.safeParse({ ...txn, units: '10', unitValue: '100', fees: '1000' })
				.success
		).toBe(false);
		expect(fundMovementEditSchema.safeParse({ ...txn, amount: '500', fees: '5' }).success).toBe(
			true
		);
		expect(fundMovementEditSchema.safeParse({ ...txn, amount: '', fees: '' }).success).toBe(false);
	});

	it('explains the refusals of a movement by units', () => {
		expect(fundMovementErrorMessage(400, 'invalid fund: units must be greater than 0')).toContain(
			'unidades'
		);
		expect(
			fundMovementErrorMessage(400, 'invalid fund: unit value must be greater than 0')
		).toContain('valor de unidad');
	});
});

describe('deleting a fund', () => {
	it('reads whether its positions go with it', () => {
		expect(fundDeleteSchema.safeParse({ id: ID, withPositions: 'on' }).data?.withPositions).toBe(
			true
		);
		expect(fundDeleteSchema.safeParse({ id: ID, withPositions: null }).data?.withPositions).toBe(
			false
		);
	});

	it('explains cash that was already spent', () => {
		expect(fundErrorMessage(409, 'insufficient cash balance')).toContain('efectivo');
	});
});
