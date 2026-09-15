import { describe, it, expect } from 'vitest';
import { groupAutomaticInterest, groupCashLedgerByMonth, type CashLedgerRow } from './interest';
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
