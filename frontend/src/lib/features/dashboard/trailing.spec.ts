import { describe, it, expect } from 'vitest';
import {
	TRAILING_LABELS,
	defaultTrailingPeriod,
	hasTrailingReturns,
	toTrailingCells,
	trailingReading,
	trailingTone
} from './trailing';
import type { TrailingReturn } from '$lib/api/types';

const entry = (over: Partial<TrailingReturn> & Pick<TrailingReturn, 'period'>): TrailingReturn => ({
	available: true,
	historyStart: '2026-03-02',
	from: '2026-09-24',
	to: '2026-09-25',
	startValue: '1000.00',
	endValue: '1004.10',
	netFlow: '0.00',
	gain: '4.10',
	returnPct: '0.41',
	...over
});

describe('toTrailingCells', () => {
	it('names each window and reads its figures', () => {
		const [day] = toTrailingCells([entry({ period: '1D' })]);

		expect(day).toMatchObject({
			period: '1D',
			label: '1 día',
			available: true,
			pct: 0.41,
			gain: 4.1,
			pctTone: 'up',
			gainTone: 'up'
		});
		expect(day.note).toMatch(/^Desde el 24 de \p{L}+\.? de 2026$/u);
	});

	it('keeps the order the backend sends', () => {
		const cells = toTrailingCells([
			entry({ period: '1D' }),
			entry({ period: '1W' }),
			entry({ period: 'ALL' })
		]);

		expect(cells.map((c) => c.label)).toEqual(['1 día', '7 días', 'Desde el inicio']);
	});

	it('says when the history starts for a window it does not reach', () => {
		const cells = toTrailingCells([
			entry({ period: '1D' }),
			{ period: '1Y', available: false, historyStart: '2026-03-02' }
		]);

		expect(cells[1]).toMatchObject({ available: false, pct: null, gain: null, pctTone: 'neutral' });
		expect(cells[1].note).toMatch(/^El historial empieza el 2 de \p{L}+\.? de 2026$/u);
	});

	// Una ganancia en dinero sin porcentaje: la cuenta estuvo vacía y no hubo
	// capital con el que medirlo. Se enseña la cifra que existe y nada más.
	it('leaves the rate empty when the backend withholds it', () => {
		const [cell] = toTrailingCells([entry({ period: '1W', returnPct: undefined, gain: '0.00' })]);

		expect(cell.pct).toBeNull();
		expect(cell.gain).toBe(0);
		expect(cell.gainTone).toBe('neutral');
	});

	it('colors each figure by its own sign', () => {
		// Mucho dinero entró justo antes de una caída: se perdió dinero aunque la
		// rentabilidad ponderada por tiempo sea positiva.
		const [cell] = toTrailingCells([entry({ period: '1M', returnPct: '1.20', gain: '-35.00' })]);

		expect(cell.pctTone).toBe('up');
		expect(cell.gainTone).toBe('down');
	});

	it('shows nothing until some window has a figure', () => {
		expect(toTrailingCells(undefined)).toEqual([]);
		expect(toTrailingCells([])).toEqual([]);
		expect(
			toTrailingCells([
				{ period: '1D', available: false, historyStart: '2026-09-25' },
				{ period: 'ALL', available: false, historyStart: '2026-09-25' }
			])
		).toEqual([]);
	});

	it('has a label for every period', () => {
		expect(Object.keys(TRAILING_LABELS)).toEqual(['1D', '1W', '1M', '3M', 'YTD', '1Y', 'ALL']);
	});
});

describe('trailingTone', () => {
	it.each([
		[1.5, 'up'],
		[-0.2, 'down'],
		[0, 'neutral'],
		[0.004, 'neutral'],
		[-0.004, 'neutral'],
		[null, 'neutral']
	] as const)('%s → %s', (value, tone) => {
		expect(trailingTone(value)).toBe(tone);
	});
});

describe('hasTrailingReturns', () => {
	it('is true as soon as one window has a figure', () => {
		expect(
			hasTrailingReturns([
				{ period: '1Y', available: false, historyStart: '2026-03-02' },
				entry({ period: 'ALL' })
			])
		).toBe(true);
	});

	it('is false without returns or without any window reached', () => {
		expect(hasTrailingReturns(undefined)).toBe(false);
		expect(hasTrailingReturns([])).toBe(false);
		expect(hasTrailingReturns([{ period: 'ALL', available: false }])).toBe(false);
	});
});

describe('defaultTrailingPeriod', () => {
	it('opens on the last month', () => {
		const cells = toTrailingCells([
			entry({ period: '1D' }),
			entry({ period: '1M' }),
			entry({ period: 'ALL' })
		]);

		expect(defaultTrailingPeriod(cells)).toBe('1M');
	});

	it('falls back to the first window with a figure while the history is short', () => {
		const cells = toTrailingCells([
			{ period: '1D', available: false, historyStart: '2026-09-20' },
			entry({ period: '1W' }),
			{ period: '1M', available: false, historyStart: '2026-09-20' }
		]);

		expect(defaultTrailingPeriod(cells)).toBe('1W');
	});

	it('has nothing to open without windows', () => {
		expect(defaultTrailingPeriod([])).toBeNull();
	});
});

describe('trailingReading', () => {
	const read = (over: Partial<TrailingReturn> & Pick<TrailingReturn, 'period'>) =>
		trailingReading(toTrailingCells([entry(over)])[0]);

	it('reads a gain from the start of the window', () => {
		const reading = read({ period: '1M', from: '2026-08-25', gain: '219.88' });

		expect(reading.before).toMatch(/^Desde el 25 de \p{L}+\.? de 2026 ganaste $/u);
		expect(reading).toMatchObject({
			amount: 219.88,
			tone: 'up',
			after: ', sin contar lo que metiste o sacaste.'
		});
	});

	it('reads a loss without a sign: the verb carries it', () => {
		const reading = read({ period: '1W', gain: '-107.90' });

		expect(reading.before).toMatch(/perdiste $/);
		expect(reading).toMatchObject({ amount: 107.9, tone: 'down' });
	});

	it('says there was no change instead of printing zero', () => {
		const reading = read({ period: '1D', gain: '0.00' });

		expect(reading.before).toMatch(/no ganaste ni perdiste, sin contar/);
		expect(reading.amount).toBeNull();
	});

	it('explains a window the history does not reach', () => {
		const [, year] = toTrailingCells([
			entry({ period: '1D' }),
			{ period: '1Y', available: false, historyStart: '2026-03-02' }
		]);
		const reading = trailingReading(year);

		expect(reading.before).toMatch(
			/^El historial empieza el 2 de \p{L}+\.? de 2026, así que todavía no hay cifra/u
		);
		expect(reading.amount).toBeNull();
	});
});
