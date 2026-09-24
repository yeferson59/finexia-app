import { describe, it, expect } from 'vitest';
import {
	cashRecalcChange,
	cashRecalcSent,
	firstRateDay,
	groupAutomaticInterest,
	groupCashLedgerByMonth,
	type CashLedgerRow,
	type CashRecalcDone
} from './interest';
import type { CashRate, CashRecalculation } from '$lib/api/types';
import type { CashMovement } from './cash';

const movement = (over: Partial<CashMovement> & { date: string }): CashMovement => ({
	id: crypto.randomUUID(),
	entryId: 'e1',
	type: 'transfer_in',
	kind: 'deposit',
	amount: '100',
	currency: 'COP',
	fees: '0',
	feesCurrency: 'COP',
	notes: '',
	editable: true,
	automatic: false,
	originTicker: '',
	portfolioId: 'p1',
	portfolioName: 'Ahorro',
	sourceId: 's1',
	sourceName: 'Nu',
	ticker: 'CASH-COP',
	createdAt: '2026-09-01T00:00:00Z',
	...over
});

const credit = (date: string, amount: string, over: Partial<CashMovement> = {}) =>
	movement({ date, amount, type: 'cash_interest', kind: 'interest', automatic: true, ...over });

/** Qué es cada fila: el tipo del movimiento, o el grupo con cuántos abonos lleva. */
const shape = (rows: CashLedgerRow[]) =>
	rows.map((row) =>
		row.type === 'interest' ? `grupo de ${row.movements.length}` : row.movement.kind
	);

describe('groupAutomaticInterest', () => {
	it('junta los abonos de un saldo en el mes, donde cae el más reciente', () => {
		const rows = groupAutomaticInterest([
			credit('2026-09-14T00:00:00Z', '2361.87'),
			movement({ date: '2026-09-13T00:00:00Z', kind: 'withdrawal', amount: '50000' }),
			credit('2026-09-13T00:00:00Z', '2361.31'),
			credit('2026-09-12T00:00:00Z', '2360.75')
		]);

		expect(shape(rows)).toEqual(['grupo de 3', 'withdrawal']);

		const group = rows[0];
		if (group.type !== 'interest') throw new Error('se esperaba un grupo');
		expect(group.amount).toBeCloseTo(7083.93, 2);
		expect(group.date).toBe('2026-09-14T00:00:00Z');
		expect(group.since).toBe('2026-09-12T00:00:00Z');
		expect(group.sourceName).toBe('Nu');
	});

	it('separa los saldos y los meses', () => {
		const rows = groupAutomaticInterest([
			credit('2026-09-02T00:00:00Z', '1'),
			credit('2026-09-02T00:00:00Z', '4', { entryId: 'e2', currency: 'USD' }),
			credit('2026-09-01T00:00:00Z', '1'),
			credit('2026-09-01T00:00:00Z', '4', { entryId: 'e2', currency: 'USD' }),
			credit('2026-08-31T00:00:00Z', '1'),
			credit('2026-08-30T00:00:00Z', '1')
		]);

		expect(shape(rows)).toEqual(['grupo de 2', 'grupo de 2', 'grupo de 2']);
		expect(rows.map((row) => (row.type === 'interest' ? row.currency : ''))).toEqual([
			'COP',
			'USD',
			'COP'
		]);
	});

	it('deja como movimientos un abono solo en su mes y los intereses anotados a mano', () => {
		const rows = groupAutomaticInterest([
			credit('2026-09-14T00:00:00Z', '5'),
			movement({
				date: '2026-09-10T00:00:00Z',
				kind: 'interest',
				type: 'cash_interest',
				amount: '30'
			}),
			credit('2026-08-31T00:00:00Z', '4')
		]);

		expect(rows.every((row) => row.type === 'movement')).toBe(true);
		expect(shape(rows)).toEqual(['interest', 'interest', 'interest']);
	});
});

describe('groupCashLedgerByMonth', () => {
	it('parte las filas por mes, en el orden en que llegan', () => {
		const months = groupCashLedgerByMonth(
			groupAutomaticInterest([
				movement({ date: '2026-09-03T00:00:00Z' }),
				credit('2026-09-02T00:00:00Z', '1'),
				credit('2026-09-01T00:00:00Z', '1'),
				movement({ date: '2026-08-31T00:00:00Z' })
			])
		);

		expect(months.map((m) => m.key)).toEqual(['2026-09', '2026-08']);
		expect(months[0].label).toBe('Septiembre de 2026');
		expect(shape(months[0].rows)).toEqual(['deposit', 'grupo de 2']);
	});
});

describe('firstRateDay', () => {
	const rate = (effectiveFrom: string, pocketId: string | null = null) =>
		({ sourceId: 's1', currency: 'COP', pocketId, effectiveFrom }) as CashRate;

	it('toma la versión más antigua del cajón', () => {
		const rates = [
			rate('2026-09-20T00:00:00Z'),
			rate('2026-09-16T00:00:00Z'),
			rate('2026-09-01T00:00:00Z', 'p1')
		];
		expect(firstRateDay(rates, 's1', 'COP', null)).toBe('2026-09-16');
		expect(firstRateDay(rates, 's1', 'COP', 'p1')).toBe('2026-09-01');
	});

	it('es null sin tasa', () => {
		expect(firstRateDay([], 's1', 'COP', null)).toBeNull();
	});
});

describe('cashRecalcChange', () => {
	const days = (n: number, net: string) => ({
		days: n,
		from: n ? '2026-09-16T00:00:00Z' : null,
		through: n ? '2026-09-22T00:00:00Z' : null,
		net
	});

	const done = (over: Partial<CashRecalcDone> = {}, result: Partial<CashRecalculation> = {}) => ({
		key: 's1:COP',
		where: 'rappi, education',
		currency: 'COP',
		from: '2026-09-01',
		rateFrom: '2026-09-16',
		...over,
		result: {
			cleared: { from: '2026-09-16T00:00:00Z', balances: 1, days: 7 },
			through: '2026-09-22T00:00:00Z',
			credited: 0,
			recomputed: 7,
			before: days(7, '333.73619817'),
			after: days(7, '333.73619817'),
			new: days(0, '0'),
			...result
		}
	});

	it('compara los mismos días, sin contar los nuevos como cambio', () => {
		const change = cashRecalcChange(done({}, { new: days(1, '47.72') }));
		expect(change.delta).toBe(0);
		expect(change.fresh).toBeCloseTo(47.72, 8);
	});

	it('da la diferencia de los días rehechos', () => {
		const change = cashRecalcChange(done({}, { after: days(7, '380.12') }));
		expect(change.before).toBeCloseTo(333.73619817, 8);
		expect(change.delta).toBeCloseTo(46.38380183, 8);
	});

	it('no cuenta el ruido de redondeo como cambio', () => {
		expect(cashRecalcChange(done({}, { after: days(7, '333.739') })).delta).toBe(0);
	});

	it('avisa si se pidió desde antes de que la cuenta rindiera', () => {
		expect(cashRecalcChange(done()).beforeRate).toBe(true);
		expect(cashRecalcChange(done({ from: '2026-09-16' })).beforeRate).toBe(false);
		expect(cashRecalcChange(done({ rateFrom: null })).beforeRate).toBe(false);
	});
});

describe('cashRecalcSent', () => {
	const account = {
		key: 's1:COP:p1',
		sourceId: 's1',
		sourceName: 'rappi',
		currency: 'COP',
		pocketId: 'p1' as string | null,
		pocketName: 'education'
	};
	const rates = [
		{ sourceId: 's1', currency: 'COP', pocketId: 'p1', effectiveFrom: '2026-09-16T00:00:00Z' }
	] as CashRate[];

	it('nombra el bolsillo y toma el primer día de su tasa', () => {
		expect(cashRecalcSent(account, rates, '2026-09-01')).toEqual({
			key: 's1:COP:p1',
			where: 'rappi, education',
			currency: 'COP',
			from: '2026-09-01',
			rateFrom: '2026-09-16'
		});
	});

	it('en la cuenta principal nombra la moneda', () => {
		const main = { ...account, key: 's1:COP', pocketId: null, pocketName: '' };
		expect(cashRecalcSent(main, rates, '2026-09-01')).toMatchObject({
			where: 'rappi, COP',
			rateFrom: null
		});
	});
});
