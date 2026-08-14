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

	// El caso que motivó el cambio: activo en EUR dentro de un portafolio en
	// USD. Los totales y el ROI salen de los importes convertidos; el promedio
	// y el precio de mercado siguen siendo por unidad y en su moneda.
	it('keeps unit prices native and totals in the base currency', () => {
		const position = computePosition(
			[
				holding({
					ticker: 'MC.FR',
					currency: 'EUR',
					costCurrency: 'EUR',
					quantity: '2',
					price: '100',
					marketPrice: '110',
					costBasisBase: '220',
					marketValueBase: '242',
					fxConverted: true
				})
			],
			1000,
			'USD'
		);

		expect(position?.averageCost).toBe(100);
		expect(position?.marketPrice).toBe(110);
		expect(position?.totalCost).toBe(220);
		expect(position?.totalValue).toBe(242);
		expect(position?.gainLoss).toBe(22);
		expect(position?.gainLossPercent).toBeCloseTo(10, 6);
		expect(position?.allocation).toBeCloseTo(24.2, 6);
		expect(position?.baseCurrency).toBe('USD');
		expect(position?.currency).toBe('EUR');
	});

	it('reports the position as unconverted when a rate was missing', () => {
		const position = computePosition(
			[holding({ costBasisBase: '200', marketValueBase: '220', fxConverted: false })],
			1000,
			'USD'
		);

		expect(position?.fxConverted).toBe(false);
	});

	// Sin los campos (backend anterior) se vuelve al cálculo nativo y no se
	// marca nada: no hay evidencia de que falte ninguna tasa.
	it('falls back to the native totals when the base amounts are absent', () => {
		const position = computePosition([holding({ quantity: '10', price: '100' })], 1000, 'USD');

		expect(position?.totalCost).toBe(1000);
		expect(position?.totalValue).toBe(1500);
		expect(position?.fxConverted).toBe(true);
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
