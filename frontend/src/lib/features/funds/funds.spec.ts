import { describe, expect, it } from 'vitest';
import {
	FUND_STALE_DAYS,
	calendarDaysBetween,
	daysSinceMark,
	fundGain,
	fundPlatforms,
	isStale,
	markBefore,
	markChange,
	marksWithChange,
	type Fund,
	type FundMark
} from './funds';

function mark(date: string, unitValue: string): FundMark {
	return {
		date: `${date}T00:00:00Z`,
		unitValue,
		balance: null,
		notes: '',
		createdAt: `${date}T12:00:00Z`,
		updatedAt: `${date}T12:00:00Z`
	};
}

describe('calendarDaysBetween', () => {
	it('counts calendar days, whatever the time part says', () => {
		expect(calendarDaysBetween('2026-09-01', '2026-09-30T00:00:00Z')).toBe(29);
		expect(calendarDaysBetween('2026-09-30', '2026-09-01')).toBe(-29);
	});
});

describe('staleness', () => {
	it('has no age without a mark', () => {
		expect(daysSinceMark({ valuedOn: null }, '2026-09-30')).toBeNull();
		expect(isStale({ valuedOn: null }, '2026-09-30')).toBe(false);
	});

	it(`warns only past ${FUND_STALE_DAYS} days`, () => {
		expect(isStale({ valuedOn: '2026-08-26T00:00:00Z' }, '2026-09-30')).toBe(false);
		expect(isStale({ valuedOn: '2026-08-25T00:00:00Z' }, '2026-09-30')).toBe(true);
	});
});

describe('fundGain', () => {
	it('is the value over the cost', () => {
		const gain = fundGain({ value: '12431220', cost: '12345678.901', pricedAtCost: false });

		expect(gain?.amount).toBeCloseTo(85541.099, 3);
		expect(gain?.pct).toBeCloseTo(0.6929, 4);
	});

	it('says nothing for a fund valued at cost', () => {
		expect(fundGain({ value: '1000', cost: '1000', pricedAtCost: true })).toBeNull();
	});

	it('has no percentage without a cost', () => {
		expect(fundGain({ value: '10', cost: '0', pricedAtCost: false })?.pct).toBeNull();
	});
});

describe('marks', () => {
	const history = [
		mark('2026-09-01', '12345.678901'),
		mark('2026-09-30', '12431.22'),
		mark('2026-09-15', '12300')
	];

	it('measures the change between two marks', () => {
		const change = markChange(history[0], history[1]);

		expect(change?.days).toBe(29);
		expect(change?.pct).toBeCloseTo(0.6929, 4);
	});

	it('cannot measure a change backwards or against nothing', () => {
		expect(markChange(history[1], history[0])).toBeNull();
		expect(markChange(mark('2026-09-01', '0'), history[1])).toBeNull();
	});

	it('lists the newest first, each against the one before', () => {
		const listed = marksWithChange(history);

		expect(listed.map((m) => m.mark.date.slice(0, 10))).toEqual([
			'2026-09-30',
			'2026-09-15',
			'2026-09-01'
		]);
		expect(listed[1].change?.pct).toBeCloseTo(-0.37, 4);
		expect(listed[2].change).toBeNull();
	});

	it('compares a new mark with the one before its day, not the one it replaces', () => {
		expect(markBefore(history, '2026-09-30')?.date).toBe('2026-09-15T00:00:00Z');
		expect(markBefore(history, '2026-09-20')?.date).toBe('2026-09-15T00:00:00Z');
		expect(markBefore(history, '2026-09-01')).toBeNull();
	});
});

describe('fundPlatforms', () => {
	it('names each platform once', () => {
		const positions = [
			{ sourceName: 'Fiduciaria' },
			{ sourceName: 'Fiduciaria' },
			{ sourceName: '' }
		] as Fund['positions'];

		expect(fundPlatforms({ positions })).toBe('Fiduciaria · Sin plataforma');
	});
});
