import { describe, it, expect } from 'vitest';
import {
	ASSET_SERIES_COLORS,
	OTHERS_COLOR,
	PIE_MAX_SLICES,
	buildConcentration,
	formatQuantity,
	toAssetHoldingRows
} from './asset-holdings';
import type { AssetHolding } from '$lib/api/types';

function holding(over: Partial<AssetHolding> = {}): AssetHolding {
	return {
		assetId: over.ticker ?? 'id',
		ticker: 'AAPL',
		name: 'Apple Inc.',
		assetType: 'stock',
		exchange: 'NASDAQ',
		currency: 'USD',
		quantity: '10',
		marketPrice: '150',
		marketValue: '1500',
		percent: 100,
		displayCurrency: 'USD',
		portfolios: 1,
		priceSource: 'own',
		positionsUnconverted: 0,
		...over
	};
}

describe('toAssetHoldingRows', () => {
	it('resuelve la etiqueta de la clase de activo', () => {
		expect(toAssetHoldingRows([holding({ assetType: 'crypto' })])[0].typeLabel).toBe('Cripto');
	});

	it('conserva el nombre crudo de una clase que no conoce', () => {
		expect(toAssetHoldingRows([holding({ assetType: 'nft' })])[0].typeLabel).toBe('nft');
	});

	// Vacío es «no hay precio que represente al activo»: cada entrada pagó el
	// suyo. Un 0 se leería como un activo que no vale nada.
	it('deja el precio en null cuando la posición va a coste', () => {
		const [row] = toAssetHoldingRows([holding({ marketPrice: '', priceSource: 'cost' })]);
		expect(row.marketPrice).toBeNull();
	});

	it('trata un precio ilegible como ausente, no como cero', () => {
		expect(toAssetHoldingRows([holding({ marketPrice: 'n/a' })])[0].marketPrice).toBeNull();
	});

	it('marca como no convertida la fila con posiciones sin tasa', () => {
		expect(toAssetHoldingRows([holding({ positionsUnconverted: 2 })])[0].fxConverted).toBe(false);
		expect(toAssetHoldingRows([holding()])[0].fxConverted).toBe(true);
	});
});

describe('buildConcentration', () => {
	const many = (n: number) =>
		Array.from({ length: n }, (_, i) =>
			holding({
				ticker: `A${i}`,
				marketValue: String((n - i) * 100),
				percent: 100 / n
			})
		);

	it('ordena por valor y da a cada porción su color, sin repetir', () => {
		const slices = buildConcentration(toAssetHoldingRows(many(4)));
		expect(slices.map((s) => s.label)).toEqual(['A0', 'A1', 'A2', 'A3']);
		expect(slices.map((s) => s.color)).toEqual(ASSET_SERIES_COLORS.slice(0, 4));
	});

	// Pasado un puñado de porciones la torta deja de leerse; la cola va a un
	// «Otros» y el detalle queda en la tabla.
	it('agrupa la cola en «Otros» cuando hay más activos de los que caben', () => {
		const slices = buildConcentration(toAssetHoldingRows(many(20)));
		const others = slices.at(-1);

		expect(slices).toHaveLength(PIE_MAX_SLICES + 1);
		expect(others?.label).toBe('Otros');
		expect(others?.color).toBe(OTHERS_COLOR);
		expect(others?.assets).toBe(20 - PIE_MAX_SLICES);
	});

	// Un solo activo de más no merece una rebanada que lo esconda y ocupe lo
	// mismo. El límite existe para que la torta se lea, no por sí mismo.
	it('no crea «Otros» para un único activo sobrante', () => {
		const slices = buildConcentration(toAssetHoldingRows(many(PIE_MAX_SLICES + 1)));

		expect(slices).toHaveLength(PIE_MAX_SLICES + 1);
		expect(slices.every((s) => s.assets === 1)).toBe(true);
		// Y aun así hay un color propio para cada una: el índice no da la vuelta.
		expect(new Set(slices.map((s) => s.color)).size).toBe(slices.length);
	});

	it('«Otros» suma los valores y los pesos de la cola', () => {
		const rows = toAssetHoldingRows([
			holding({ ticker: 'A', marketValue: '1000', percent: 50 }),
			holding({ ticker: 'B', marketValue: '600', percent: 30 }),
			holding({ ticker: 'C', marketValue: '300', percent: 15 }),
			holding({ ticker: 'D', marketValue: '100', percent: 5 })
		]);
		const others = buildConcentration(rows, 2).at(-1);

		expect(others?.label).toBe('Otros');
		expect(others?.value).toBe(400);
		expect(others?.percent).toBe(20);
		expect(others?.assets).toBe(2);
	});

	// Una posición vendida entera vale 0 y no es una porción: dibujarla mete en
	// la leyenda una entrada de 0% que no se ve en la gráfica.
	it('deja fuera lo que no vale nada', () => {
		const rows = toAssetHoldingRows([
			holding({ ticker: 'A', marketValue: '100', percent: 100 }),
			holding({ ticker: 'B', marketValue: '0', percent: 0 })
		]);

		expect(buildConcentration(rows).map((s) => s.label)).toEqual(['A']);
	});

	it('sin activos no hay porciones', () => {
		expect(buildConcentration([])).toEqual([]);
	});
});

describe('formatQuantity', () => {
	// No es dinero: redondear a dos decimales convierte una posición en cripto
	// en cero.
	it('conserva los decimales de una cantidad pequeña', () => {
		expect(formatQuantity(0.00000123)).toBe('0,00000123');
	});

	it('no inventa decimales en una cantidad entera', () => {
		expect(formatQuantity(15)).toBe('15');
	});
});
