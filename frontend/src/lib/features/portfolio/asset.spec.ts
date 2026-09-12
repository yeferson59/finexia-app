import { describe, it, expect } from 'vitest';
import type { Holding } from '$lib/api/types';
import {
	computePosition,
	formatUnits,
	priceLabelFor,
	tradeDateWarnings,
	transactionsNeedingRate,
	txnModeFor,
	unitNoun
} from './asset';

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

describe('unitNoun', () => {
	// «42 AAPL» usaba el ticker de unidad de medida. Lo que se tiene son
	// acciones, y de un ETF, participaciones.
	it('names what each asset class is counted in', () => {
		expect(unitNoun('stock', 1)).toBe('acción');
		expect(unitNoun('stock', 42)).toBe('acciones');
		expect(unitNoun('etf', 120)).toBe('participaciones');
		expect(unitNoun('bond', 1)).toBe('título');
	});

	// Inventarle un nombre propio a cripto o al efectivo arriesga más de lo que
	// aclara, y una clase que el backend añada no puede quedarse sin palabra.
	it('falls back to plain units for anything else', () => {
		expect(unitNoun('crypto', 0.15)).toBe('unidades');
		expect(unitNoun('cash', 9500)).toBe('unidades');
		expect(unitNoun('esoteric', 1)).toBe('unidad');
	});
});

describe('formatUnits', () => {
	it('writes the amount with es-CO separators and its unit', () => {
		expect(formatUnits(42, 'stock')).toBe('42 acciones');
		expect(formatUnits(1, 'stock')).toBe('1 acción');
		expect(formatUnits(0.15, 'crypto')).toBe('0,15 unidades');
		expect(formatUnits(9500, 'cash')).toBe('9.500 unidades');
	});
});

describe('transactionsNeedingRate', () => {
	const history = [
		{ id: 'buy-eur', type: 'buy', currency: 'EUR' },
		{ id: 'buy-usd', type: 'buy', currency: 'USD' },
		{ id: 'split-eur', type: 'split', currency: 'EUR' },
		{ id: 'dividend-eur', type: 'dividend', currency: 'eur ' }
	];

	// IBCZ.DE pasada a una cuenta en dólares: todo lo que cotizó en euros pide
	// su tasa, menos el split, que no movió dinero.
	it('asks a rate for every transaction quoted in another currency', () => {
		expect(transactionsNeedingRate(history, 'USD').map((txn) => txn.id)).toEqual([
			'buy-eur',
			'dividend-eur'
		]);
	});

	it('asks none when everything is quoted in the new currency', () => {
		expect(transactionsNeedingRate(history, ' eur').map((txn) => txn.id)).toEqual(['buy-usd']);
	});
});

describe('tradeDateWarnings', () => {
	// PG, comprada el mismo día que se registró y a su precio: nada que decir.
	const base = {
		date: '2026-08-13',
		today: '2026-08-13',
		ticker: 'PG',
		assetType: 'stock',
		price: 145.02,
		currency: 'USD',
		marketPrice: 145.27,
		marketCurrency: 'USD'
	};

	it('stays quiet for a trade that fits its date', () => {
		expect(tradeDateWarnings(base)).toEqual([]);
	});

	// TSM, SPCX y VST entraron fechadas el sábado en que se cargaron.
	it('flags a weekend date for an exchange-traded asset', () => {
		const warnings = tradeDateWarnings({ ...base, date: '2026-08-22', today: '2026-08-22' });

		expect(warnings).toHaveLength(1);
		expect(warnings[0]).toContain('sábado');
	});

	it('knows Sunday too, and lets crypto trade on weekends', () => {
		expect(tradeDateWarnings({ ...base, date: '2026-08-23', today: '2026-08-23' })[0]).toContain(
			'domingo'
		);
		expect(
			tradeDateWarnings({ ...base, assetType: 'crypto', date: '2026-08-22', today: '2026-08-22' })
		).toEqual([]);
	});

	// ADBE: comprada «ayer» a 483,57 cuando valía 252,23.
	it('flags a recent date whose price is far from the market', () => {
		const warnings = tradeDateWarnings({
			...base,
			ticker: 'ADBE',
			today: '2026-08-14',
			price: 483.57,
			marketPrice: 252.23
		});

		expect(warnings).toHaveLength(1);
		expect(warnings[0]).toContain('ADBE');
		expect(warnings[0]).toContain('revisa la fecha');
	});

	it('trusts an old date whatever its price', () => {
		expect(
			tradeDateWarnings({
				...base,
				ticker: 'ADBE',
				date: '2024-04-11',
				today: '2026-08-14',
				price: 483.57,
				marketPrice: 252.23
			})
		).toEqual([]);
	});

	it('compares prices only in the same currency and against a known market price', () => {
		expect(
			tradeDateWarnings({ ...base, price: 213.24, marketPrice: 433.24, marketCurrency: 'EUR' })
		).toEqual([]);
		expect(tradeDateWarnings({ ...base, price: 213.24, marketPrice: null })).toEqual([]);
	});

	// TSM: fecha de sábado y a la mitad de lo que vale.
	it('raises both warnings when both signals are there', () => {
		expect(
			tradeDateWarnings({
				...base,
				ticker: 'TSM',
				date: '2026-08-22',
				today: '2026-08-22',
				price: 213.24,
				marketPrice: 433.24
			})
		).toHaveLength(2);
	});
});
