import { describe, it, expect } from 'vitest';
import {
	breakdownFor,
	platformBreakdown,
	plural,
	portfolioBreakdown,
	typeBreakdown
} from './breakdown';
import type { AllocationItem, Platform, PortfolioSummary } from '$lib/api/types';

const platform = (over: Partial<Platform> & { id: string; name: string }): Platform =>
	({
		description: '',
		sourceType: 'broker',
		isActive: true,
		investments: 1,
		totalValue: '100',
		displayCurrency: 'USD',
		createdAt: '2026-01-01',
		...over
	}) as Platform;

const summary = (
	over: Partial<PortfolioSummary> & { id: string; name: string }
): PortfolioSummary =>
	({
		type: 'stocks_etfs',
		baseCurrency: 'USD',
		displayCurrency: 'USD',
		riskName: 'Moderado',
		totalPositions: 1,
		totalCostBase: '100',
		totalMarketValue: '100',
		totalGainLoss: '0',
		totalGainLossPct: '0',
		...over
	}) as PortfolioSummary;

const allocation = (category: string, marketValue: string, over: Partial<AllocationItem> = {}) =>
	({ category, marketValue, percent: 0, currency: 'USD', ...over }) as AllocationItem;

describe('plural', () => {
	it('concuerda el sustantivo con la cifra', () => {
		expect(plural(1, 'posición', 'posiciones')).toBe('1 posición');
		expect(plural(4, 'posición', 'posiciones')).toBe('4 posiciones');
		expect(plural(0, 'posición', 'posiciones')).toBe('0 posiciones');
	});
});

describe('platformBreakdown', () => {
	it('reparte el peso de cada plataforma sobre el total enseñado', () => {
		const { rows, total } = platformBreakdown(
			[
				platform({ id: 'a', name: 'Degiro', marketValue: '75' }),
				platform({ id: 'b', name: 'Binance', marketValue: '25' })
			],
			'USD'
		);

		expect(total).toBe(100);
		expect(rows.map((r) => r.share)).toEqual([0.75, 0.25]);
	});

	it('ordena de mayor a menor, que es el orden en que se lee un reparto', () => {
		const { rows } = platformBreakdown(
			[
				platform({ id: 'a', name: 'Pequeña', marketValue: '10' }),
				platform({ id: 'b', name: 'Grande', marketValue: '90' })
			],
			'USD'
		);

		expect(rows.map((r) => r.label)).toEqual(['Grande', 'Pequeña']);
	});

	// Sumar euros con dólares da un total que no está en ninguna moneda.
	it('deja fuera lo que no está en la moneda pedida y dice cuánto falta', () => {
		const { rows, excluded, total } = platformBreakdown(
			[
				platform({ id: 'a', name: 'Degiro', marketValue: '100' }),
				platform({ id: 'b', name: 'Revolut', marketValue: '500', displayCurrency: 'EUR' })
			],
			'USD'
		);

		expect(rows).toHaveLength(1);
		expect(excluded).toBe(1);
		expect(total).toBe(100);
	});

	// Una ganancia de cero por falta de precio no es una ganancia de cero.
	it('no informa rendimiento cuando todas las posiciones se valoran a coste', () => {
		const [row] = platformBreakdown(
			[
				platform({
					id: 'a',
					name: 'Banco',
					marketValue: '100',
					investments: 3,
					positionsAtCost: 3,
					gainLossPct: 0
				})
			],
			'USD'
		).rows;

		expect(row.gainPct).toBeNull();
	});

	it('sí lo informa cuando alguna posición tiene precio de mercado', () => {
		const [row] = platformBreakdown(
			[
				platform({
					id: 'a',
					name: 'Banco',
					marketValue: '100',
					investments: 3,
					positionsAtCost: 2,
					gainLossPct: 4.5
				})
			],
			'USD'
		).rows;

		expect(row.gainPct).toBe(4.5);
	});

	// Backend anterior: sin `marketValue` ni `gainLossPct` queda lo invertido, que
	// es el dato que esa versión sabe dar, y ningún rendimiento inventado.
	it('cae a lo invertido cuando el backend no manda valor de mercado', () => {
		const [row] = platformBreakdown(
			[platform({ id: 'a', name: 'Vieja', totalValue: '42' })],
			'USD'
		).rows;

		expect(row.value).toBe(42);
		expect(row.gainPct).toBeNull();
	});

	it('suma las posiciones que van sin convertir dentro de las filas contadas', () => {
		const { unconverted } = platformBreakdown(
			[
				platform({ id: 'a', name: 'Degiro', marketValue: '100', positionsUnconverted: 2 }),
				platform({ id: 'b', name: 'Banco', marketValue: '50', positionsUnconverted: 1 })
			],
			'USD'
		);

		expect(unconverted).toBe(3);
	});
});

describe('portfolioBreakdown', () => {
	it('lleva el tipo del portafolio ya traducido en la segunda línea', () => {
		const [row] = portfolioBreakdown(
			[summary({ id: 'a', name: 'Cartera', type: 'stocks_etfs', totalMarketValue: '10' })],
			'USD'
		).rows;

		expect(row.label).toBe('Cartera');
		expect(row.detail).toBe('Acciones & ETFs');
	});

	it('mide el rendimiento con la cifra del backend', () => {
		const [row] = portfolioBreakdown(
			[summary({ id: 'a', name: 'Cartera', totalMarketValue: '10', totalGainLossPct: '-2.59' })],
			'USD'
		).rows;

		expect(row.gainPct).toBeCloseTo(-2.59);
	});

	it('deja fuera el portafolio que se quedó en su propia moneda', () => {
		const { rows, excluded } = portfolioBreakdown(
			[
				summary({ id: 'a', name: 'Cartera', totalMarketValue: '10' }),
				summary({ id: 'b', name: 'Europa', totalMarketValue: '90', displayCurrency: 'EUR' })
			],
			'USD'
		);

		expect(rows.map((r) => r.label)).toEqual(['Cartera']);
		expect(excluded).toBe(1);
	});
});

describe('typeBreakdown', () => {
	it('mide participación y no rendimiento: el endpoint no devuelve ganancias', () => {
		const { rows, trailing } = typeBreakdown(
			[allocation('stock', '60'), allocation('crypto', '40')],
			'USD'
		);

		expect(trailing).toBe('share');
		expect(rows.map((r) => r.gainPct)).toEqual([null, null]);
		expect(rows.map((r) => r.share)).toEqual([0.6, 0.4]);
	});

	it('traduce la clase de activo', () => {
		const { rows } = typeBreakdown([allocation('bond', '10')], 'USD');
		expect(rows[0].label).toBe('Bonos');
	});

	it('descarta el reparto que viene en otra moneda', () => {
		const { rows, excluded } = typeBreakdown(
			[allocation('stock', '60'), allocation('crypto', '40', { currency: 'COP' })],
			'USD'
		);

		expect(rows).toHaveLength(1);
		expect(excluded).toBe(1);
	});
});

describe('breakdownFor', () => {
	const source = {
		platforms: [platform({ id: 'p', name: 'Degiro', marketValue: '10' })],
		summaries: [summary({ id: 's', name: 'Cartera', totalMarketValue: '20' })],
		allocation: [allocation('stock', '30')]
	};

	it('devuelve el corte pedido', () => {
		expect(breakdownFor('platform', source, 'USD').rows[0].label).toBe('Degiro');
		expect(breakdownFor('portfolio', source, 'USD').rows[0].label).toBe('Cartera');
		expect(breakdownFor('type', source, 'USD').rows[0].label).toBe('Acciones');
	});

	// Sin total, dividir daría NaN y la barra saldría llena o vacía al azar.
	it('no divide por cero cuando todas las filas valen cero', () => {
		const { rows, total } = breakdownFor(
			'platform',
			{ ...source, platforms: [platform({ id: 'p', name: 'Vacía', marketValue: '0' })] },
			'USD'
		);

		expect(total).toBe(0);
		expect(rows[0].share).toBe(0);
	});

	it('no se cae sin datos', () => {
		const empty = { platforms: [], summaries: [], allocation: [] };
		expect(breakdownFor('platform', empty, 'USD').rows).toEqual([]);
		expect(breakdownFor('type', empty, 'USD').total).toBe(0);
	});
});
