import { describe, it, expect } from 'vitest';
import type { Holding } from '$lib/api/types';
import { computePosition, priceLabelFor, txnModeFor } from './asset';

function holding(partial: Partial<Holding>): Holding {
	return {
		id: 'h1',
		assetId: 'a1',
		ticker: 'AAPL',
		name: 'Apple Inc.',
		assetType: 'stock',
		exchange: 'NASDAQ',
		currency: 'USD',
		quantity: '10',
		price: '100',
		marketPrice: '150',
		costCurrency: 'USD',
		category: 'stock',
		entryDate: '2026-01-01',
		notes: '',
		...partial
	};
}

describe('computePosition', () => {
	it('returns null when the asset has no entries', () => {
		expect(computePosition([], 1000)).toBeNull();
	});

	it('aggregates quantity and cost across entries and derives the average cost', () => {
		const position = computePosition(
			[
				holding({ id: 'h1', quantity: '10', price: '100' }),
				holding({ id: 'h2', quantity: '30', price: '200' })
			],
			8000
		);

		expect(position).not.toBeNull();
		expect(position?.totalQty).toBe(40);
		expect(position?.totalCost).toBe(7000);
		expect(position?.averageCost).toBe(175);
		// marketPrice viene de la primera entrada (150) → 40 * 150
		expect(position?.totalValue).toBe(6000);
		expect(position?.gainLoss).toBe(-1000);
		expect(position?.gainLossPercent).toBeCloseTo(-14.2857, 3);
		expect(position?.allocation).toBe(75);
	});

	it('falls back to the average cost when the market price is missing', () => {
		const position = computePosition([holding({ marketPrice: '0', price: '50' })], 0);

		expect(position?.marketPrice).toBe(50);
		expect(position?.gainLoss).toBe(0);
		// sin valor total de portafolio no se puede calcular asignación
		expect(position?.allocation).toBe(0);
	});
});

describe('txnModeFor', () => {
	it('classifies trades, amounts and splits', () => {
		expect(txnModeFor('buy')).toBe('trade');
		expect(txnModeFor('sell')).toBe('trade');
		expect(txnModeFor('transfer_in')).toBe('trade');
		expect(txnModeFor('dividend')).toBe('amount');
		expect(txnModeFor('fee')).toBe('amount');
		expect(txnModeFor('interest')).toBe('amount');
		expect(txnModeFor('split')).toBe('split');
	});
});

describe('priceLabelFor', () => {
	it('labels the amount field per transaction type', () => {
		expect(priceLabelFor('buy')).toBe('Precio unitario');
		expect(priceLabelFor('split')).toBe('Precio unitario');
		expect(priceLabelFor('dividend')).toBe('Monto del dividendo');
		expect(priceLabelFor('interest')).toBe('Monto del interés');
		expect(priceLabelFor('fee')).toBe('Monto de la comisión');
	});
});
