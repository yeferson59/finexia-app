import { describe, expect, it } from 'vitest';
import {
	fundPeriodRows,
	headlinePeriod,
	parseMarksTable,
	parseStatementDate,
	parseStatementNumber,
	type FundPeriod
} from './performance';

const period = (
	key: FundPeriod['key'],
	pct: string | null,
	ea: string | null = null
): FundPeriod => ({
	key,
	from: pct === null ? null : '2026-08-31T00:00:00Z',
	to: '2026-09-30T00:00:00Z',
	days: pct === null ? null : 30,
	pct,
	eaPct: ea
});

describe('fundPeriodRows', () => {
	it('orders the periods and reads their figures', () => {
		const rows = fundPeriodRows([
			period('inception', '0.5383', '2.18'),
			period('30d', '-0.4577', '-5.43'),
			period('ytd', null)
		]);

		expect(rows.map((r) => r.label)).toEqual(['30 días', 'Año corrido', 'Desde el inicio']);
		expect(rows[0].pct).toBe(-0.4577);
		expect(rows[0].ea).toBe(-5.43);
		expect(rows[1].pct).toBeNull();
	});
});

describe('headlinePeriod', () => {
	it('prefers the last 30 days', () => {
		expect(headlinePeriod([period('inception', '0.5'), period('30d', '0.1')])?.key).toBe('30d');
	});

	it('falls back to inception for a fund younger than a month', () => {
		expect(headlinePeriod([period('30d', null), period('inception', '0.5')])?.key).toBe(
			'inception'
		);
	});

	it('has nothing to say without history', () => {
		expect(headlinePeriod([period('30d', null), period('inception', null)])).toBeNull();
	});
});

describe('parseStatementNumber', () => {
	it.each([
		['12.431,22', 12431.22],
		['12,431.22', 12431.22],
		['12431.22', 12431.22],
		['12431,22', 12431.22],
		['$ 13.050.000', 13050000],
		['1.234.567,891', 1234567.891],
		['100,538', 100.538],
		['1.008', 1.008]
	])('reads %s', (raw, want) => {
		expect(parseStatementNumber(raw)).toBeCloseTo(want, 6);
	});

	it('refuses what has no digits', () => {
		expect(parseStatementNumber('Valor')).toBeNull();
	});
});

describe('parseStatementDate', () => {
	it.each([
		['2026-09-30', '2026-09-30'],
		['30/09/2026', '2026-09-30'],
		['1-9-2026', '2026-09-01']
	])('reads %s', (raw, want) => {
		expect(parseStatementDate(raw)).toBe(want);
	});

	it('refuses a day that does not exist, or a year without its century', () => {
		expect(parseStatementDate('31/09/2026')).toBeNull();
		expect(parseStatementDate('30/09/26')).toBeNull();
	});
});

describe('parseMarksTable', () => {
	it('reads a table pasted from a sheet, header included', () => {
		const table = parseMarksTable(
			'Fecha\tValor unidad\n30/09/2026\t12.431,22\n01/09/2026\t12.345,678901\n\n15/09/2026\t12.380,00'
		);

		expect(table.errors).toEqual([]);
		expect(table.rows).toEqual([
			{ date: '2026-09-01', value: 12345.678901 },
			{ date: '2026-09-15', value: 12380 },
			{ date: '2026-09-30', value: 12431.22 }
		]);
	});

	it('takes semicolons and spaces, and the last figure of a line', () => {
		const table = parseMarksTable('2026-09-30;13.050.000\n2026-08-31 $ 15.110.000');

		expect(table.rows.map((r) => r.value)).toEqual([15110000, 13050000]);
	});

	it('reports the lines it cannot read and keeps the last figure of a day', () => {
		const table = parseMarksTable('2026-09-01 100\nsin fecha 5\n2026-09-01 101');

		expect(table.errors).toEqual([{ line: 2, text: 'sin fecha 5' }]);
		expect(table.rows).toEqual([{ date: '2026-09-01', value: 101 }]);
	});
});
