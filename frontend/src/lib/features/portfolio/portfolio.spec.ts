import { describe, it, expect } from 'vitest';
import {
	groupHoldings,
	computeTypeBreakdown,
	formatPct,
	portfolioBarScale,
	portfolioTotals,
	realReturnPct,
	toPortfolioRows,
	type RawHolding
} from './portfolio';
import type { PortfolioSummary } from '$lib/api/types';
import { portfolioEntrySchema } from './schemas';

// Un portafolio en USD: los totales en moneda base coinciden con
// cantidad × precio porque no hay conversión de por medio.
const raw: RawHolding[] = [
	{
		ticker: 'AAPL',
		name: 'Apple',
		assetType: 'stock',
		quantity: '10',
		price: '100',
		marketPrice: '150',
		currency: 'USD',
		costBasisBase: '1000',
		marketValueBase: '1500',
		fxConverted: true
	},
	// Same ticker in another platform: must aggregate.
	{
		ticker: 'AAPL',
		name: 'Apple',
		assetType: 'stock',
		quantity: '5',
		price: '120',
		marketPrice: '150',
		currency: 'USD',
		costBasisBase: '600',
		marketValueBase: '750',
		fxConverted: true
	},
	{
		ticker: 'BTC',
		name: 'Bitcoin',
		assetType: 'crypto',
		quantity: '1',
		price: '20000',
		marketPrice: '25000',
		currency: 'USD',
		costBasisBase: '20000',
		marketValueBase: '25000',
		fxConverted: true
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

	// Una posición en EUR dentro de un portafolio en USD: los totales vienen
	// convertidos y son los que hay que sumar; el precio unitario se queda en
	// euros porque es lo que cotiza.
	it('uses the converted totals, not quantity times the native price', () => {
		const [holding] = groupHoldings([
			{
				ticker: 'MC.FR',
				name: 'LVMH',
				assetType: 'stock',
				quantity: '2',
				price: '100',
				marketPrice: '110',
				currency: 'EUR',
				costBasisBase: '220',
				marketValueBase: '242',
				fxConverted: true
			}
		]);

		expect(holding.costBasis).toBe(220);
		expect(holding.value).toBe(242);
		expect(holding.marketPrice).toBe(110);
		expect(holding.currency).toBe('EUR');
	});

	// Sin tasa, los importes llegan sin convertir: la fila se pinta igual pero
	// queda marcada para que la vista avise en vez de sumar monedas distintas.
	it('flags a group as unconverted when any of its entries lacked a rate', () => {
		const [holding] = groupHoldings([
			{
				ticker: 'MC.FR',
				name: 'LVMH',
				assetType: 'stock',
				quantity: '2',
				price: '100',
				marketPrice: '110',
				currency: 'EUR',
				costBasisBase: '200',
				marketValueBase: '220',
				fxConverted: false
			}
		]);

		expect(holding.fxConverted).toBe(false);
	});

	// Contra un backend anterior a estos campos no hay nada que avisar: se
	// vuelve al cálculo nativo y la fila no se marca como sospechosa.
	it('falls back to the native calculation when the base totals are absent', () => {
		const [holding] = groupHoldings([
			{
				ticker: 'AAPL',
				name: 'Apple',
				assetType: 'stock',
				quantity: '3',
				price: '100',
				marketPrice: '150',
				currency: 'USD',
				costBasisBase: undefined,
				marketValueBase: undefined,
				fxConverted: undefined
			}
		]);

		expect(holding.costBasis).toBe(300);
		expect(holding.value).toBe(450);
		expect(holding.fxConverted).toBe(true);
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
		expect(formatPct(12.345)).toBe('+12,35%');
		expect(formatPct(-3.2)).toBe('-3,20%');
	});
});

describe('realReturnPct', () => {
	const point = (date: string, totalValue: string, cost: string, netFlow?: string) => ({
		date,
		totalValue,
		totalCostBase: cost,
		gainLoss: String(Number(totalValue) - Number(cost)),
		gainLossPct: '0',
		...(netFlow === undefined ? {} : { netFlow })
	});

	it('encadena los tramos del historial y lo devuelve en porcentaje', () => {
		const growth = [
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '1100', '1000'),
			point('2026-01-03', '1210', '1000')
		];

		expect(realReturnPct(growth)).toBeCloseTo(21, 10);
	});

	// La ganancia sobre coste diría −0 % aquí; la rentabilidad real dice que el
	// mercado subió un 10 % antes de que entrara el dinero nuevo.
	it('no cuenta un aporte como rentabilidad', () => {
		const growth = [
			point('2026-01-01', '1000', '1000'),
			point('2026-01-02', '1100', '1000'),
			point('2026-01-03', '2100', '2000', '1000')
		];

		expect(realReturnPct(growth)).toBeCloseTo(10, 10);
	});

	it('es null mientras el portafolio no tenga dos cierres', () => {
		expect(realReturnPct([point('2026-01-01', '1000', '1000')])).toBeNull();
		expect(realReturnPct([])).toBeNull();
		expect(realReturnPct(undefined)).toBeNull();
	});
});

describe('portfolioEntrySchema', () => {
	const entry = {
		portfolioId: '3f1c1c5e-1f5a-4f1e-9c2a-9a1d0b2f7e11',
		assetId: '5b2d6a4c-2a3b-4c5d-8e9f-0a1b2c3d4e5f',
		sourceId: 'a1b2c3d4-e5f6-4a5b-9c8d-7e6f5a4b3c2d',
		quantity: '0.201065',
		price: '866.60',
		entryDate: '2024-04-11'
	};

	it('normaliza las dos monedas a ISO en mayúsculas', () => {
		const parsed = portfolioEntrySchema.safeParse({
			...entry,
			costCurrency: ' dkk ',
			currency: ' dkk '
		});
		expect(parsed.success).toBe(true);
		expect(parsed.data?.costCurrency).toBe('DKK');
		expect(parsed.data?.currency).toBe('DKK');
	});

	// El formulario la enviaba fija en USD: una acción danesa entraba con su
	// precio en coronas etiquetado como dólares y el coste quedaba sin convertir.
	it('rechaza una moneda ausente o que no sea un código de tres letras', () => {
		const valid = { ...entry, currency: 'DKK' };
		expect(portfolioEntrySchema.safeParse({ ...valid, costCurrency: undefined }).success).toBe(
			false
		);
		expect(portfolioEntrySchema.safeParse({ ...valid, costCurrency: '' }).success).toBe(false);
		expect(portfolioEntrySchema.safeParse({ ...valid, costCurrency: 'corona' }).success).toBe(
			false
		);
	});

	// Una tasa ausente es tasa 1, que es lo que vale toda compra en la que la
	// cuenta y el mercado usan la misma moneda: el formulario no la manda y el
	// backend la acepta porque las dos monedas coinciden.
	it('una tasa ausente vale 1', () => {
		const parsed = portfolioEntrySchema.safeParse({
			...entry,
			costCurrency: 'DKK',
			currency: 'DKK'
		});
		expect(parsed.success).toBe(true);
		expect(parsed.data?.fxRate).toBe(1);
	});

	it('conserva la tasa de una compra convertida por el bróker', () => {
		const parsed = portfolioEntrySchema.safeParse({
			...entry,
			price: '606.60',
			costCurrency: 'USD',
			currency: 'EUR',
			fxRate: '1.0638'
		});
		expect(parsed.success).toBe(true);
		expect(parsed.data?.fxRate).toBe(1.0638);
	});

	// Cero o negativa no convierte nada: borra el coste de la posición. El
	// `.catch(1)` del esquema no debe tragárselas.
	it('rechaza una tasa que no sea positiva', () => {
		const base = { ...entry, costCurrency: 'USD', currency: 'EUR' };
		expect(portfolioEntrySchema.safeParse({ ...base, fxRate: '0' }).success).toBe(false);
		expect(portfolioEntrySchema.safeParse({ ...base, fxRate: '-1.0638' }).success).toBe(false);
	});
});

// ---------------------------------------------------------------------------
// Listado de portafolios
// ---------------------------------------------------------------------------

function summary(over: Partial<PortfolioSummary> = {}): PortfolioSummary {
	return {
		id: over.name ?? 'p1',
		name: 'Cartera Principal',
		description: 'Acciones y ETFs a largo plazo',
		type: 'stocks_etfs',
		baseCurrency: 'USD',
		displayCurrency: 'USD',
		riskName: 'Moderado',
		totalPositions: 5,
		totalCostBase: '1000',
		totalMarketValue: '1200',
		totalGainLoss: '200',
		totalGainLossPct: '20',
		...over
	};
}

describe('toPortfolioRows', () => {
	// El orden es lo que hace legible una escalera: la barra más larga primero
	// y las demás leídas contra ella.
	it('ordena de mayor a menor valor', () => {
		const rows = toPortfolioRows(
			[
				summary({ name: 'B', totalMarketValue: '500' }),
				summary({ name: 'A', totalMarketValue: '900' })
			],
			'USD'
		);

		expect(rows.map((r) => r.name)).toEqual(['A', 'B']);
	});

	// Su importe está en otra moneda: ponerlo entre los demás invitaría a
	// comparar barras que miden cosas distintas.
	it('manda al final lo que no se pudo convertir, aunque valga más', () => {
		const rows = toPortfolioRows(
			[
				summary({ name: 'EUR', displayCurrency: 'EUR', totalMarketValue: '9000' }),
				summary({ name: 'USD', totalMarketValue: '100' })
			],
			'USD'
		);

		expect(rows.map((r) => r.name)).toEqual(['USD', 'EUR']);
		expect(rows.map((r) => r.converted)).toEqual([true, false]);
	});

	// Lo que escribió el dueño es mejor descripción que la etiqueta del tipo;
	// la etiqueta es la reserva para quien no escribió nada.
	it('conserva la descripción del dueño y la etiqueta del tipo como reserva', () => {
		const [written] = toPortfolioRows([summary()], 'USD');
		expect(written.description).toBe('Acciones y ETFs a largo plazo');
		expect(written.typeLabel).toBe('Acciones & ETFs');

		const [blank] = toPortfolioRows([summary({ description: undefined })], 'USD');
		expect(blank.description).toBe('');
		expect(blank.typeLabel).toBe('Acciones & ETFs');
	});

	it('etiqueta el importe con la moneda en la que de verdad está', () => {
		const [row] = toPortfolioRows(
			[summary({ displayCurrency: undefined, baseCurrency: 'COP' })],
			'USD'
		);

		expect(row.currency).toBe('COP');
		expect(row.converted).toBe(false);
	});
});

describe('portfolioBarScale', () => {
	// En pérdida, el capital queda por fuera del extremo de la barra: con el
	// carril escalado solo a los valores, ese tramo se salía del ancho.
	it('cuenta también los costes, no solo los valores', () => {
		const rows = toPortfolioRows(
			[summary({ totalMarketValue: '800', totalCostBase: '1500' })],
			'USD'
		);

		expect(portfolioBarScale(rows)).toBe(1500);
	});

	it('deja fuera de la escala lo que está en otra moneda', () => {
		const rows = toPortfolioRows(
			[
				summary({ name: 'A', totalMarketValue: '100', totalCostBase: '100' }),
				summary({ name: 'B', displayCurrency: 'EUR', totalMarketValue: '9000' })
			],
			'USD'
		);

		expect(portfolioBarScale(rows)).toBe(100);
	});

	it('sin filas no hay escala', () => {
		expect(portfolioBarScale([])).toBe(0);
	});
});

describe('portfolioTotals', () => {
	// Sumar un portafolio en otra moneda daría una cifra que no está en
	// ninguna: se lista, pero no entra en el total.
	it('suma solo lo que está en la moneda pedida y cuenta lo que dejó fuera', () => {
		const rows = toPortfolioRows(
			[
				summary({ name: 'A', totalMarketValue: '1200', totalCostBase: '1000' }),
				summary({ name: 'B', totalMarketValue: '800', totalCostBase: '1000' }),
				summary({ name: 'C', displayCurrency: 'EUR', totalMarketValue: '5000' })
			],
			'USD'
		);
		const totals = portfolioTotals(rows);

		expect(totals.value).toBe(2000);
		expect(totals.cost).toBe(2000);
		expect(totals.gain).toBe(0);
		expect(totals.counted).toBe(2);
		expect(totals.excluded).toBe(1);
	});

	it('calcula el rendimiento sobre lo que costó', () => {
		const rows = toPortfolioRows(
			[summary({ totalMarketValue: '1250', totalCostBase: '1000' })],
			'USD'
		);

		expect(portfolioTotals(rows).gainPct).toBeCloseTo(25, 10);
	});

	// Sin coste no hay porcentaje que calcular, y un 0 se leería como un
	// portafolio que no se movió.
	it('no inventa un rendimiento sin coste', () => {
		const rows = toPortfolioRows(
			[summary({ totalMarketValue: '0', totalCostBase: '0', totalPositions: 0 })],
			'USD'
		);

		expect(portfolioTotals(rows).gainPct).toBe(0);
		expect(portfolioTotals(rows).positions).toBe(0);
	});
});
