import { describe, it, expect } from 'vitest';
import { groupHoldings, computeTypeBreakdown, formatPct, type RawHolding } from './portfolio';

const raw: RawHolding[] = [
	{
		ticker: 'AAPL',
		name: 'Apple',
		assetType: 'stock',
		quantity: '10',
		price: '100',
		marketPrice: '150'
	},
	// Same ticker in another platform: must aggregate.
	{
		ticker: 'AAPL',
		name: 'Apple',
		assetType: 'stock',
		quantity: '5',
		price: '120',
		marketPrice: '150'
	},
	{
		ticker: 'BTC',
		name: 'Bitcoin',
		assetType: 'crypto',
		quantity: '1',
		price: '20000',
		marketPrice: '25000'
	}
];

describe('groupHoldings', () => {
	it('aggregates the same ticker across platforms and computes derived metrics', () => {
		const holdings = groupHoldings(raw);
		expect(holdings).toHaveLength(2);

		const aapl = holdings.find((h) => h.symbol === 'AAPL')!;
		expect(aapl.quantity).toBe(15);
		// cost basis = 10*100 + 5*120 = 1600; value = 15*150 = 2250
		expect(aapl.costBasis).toBe(1600);
		expect(aapl.value).toBe(2250);
		expect(aapl.gainLoss).toBe(650);
	});

	it('computes allocation percentages that sum to ~100', () => {
		const holdings = groupHoldings(raw);
		const total = holdings.reduce((s, h) => s + h.allocation, 0);
		expect(Math.round(total)).toBe(100);
	});

	it('returns an empty array for no entries', () => {
		expect(groupHoldings([])).toEqual([]);
	});
});

describe('computeTypeBreakdown', () => {
	it('groups holdings by asset type with readable labels and percentages', () => {
		const breakdown = computeTypeBreakdown(groupHoldings(raw));
		const types = breakdown.map((b) => b.type);
		expect(types).toContain('stock');
		expect(types).toContain('crypto');
		const sumPct = breakdown.reduce((s, b) => s + b.pct, 0);
		expect(Math.round(sumPct)).toBe(100);
	});
});

describe('formatPct', () => {
	it('prefixes non-negative values with a plus sign', () => {
		expect(formatPct(12.345)).toBe('+12.35%');
		expect(formatPct(-3.2)).toBe('-3.20%');
	});
});
