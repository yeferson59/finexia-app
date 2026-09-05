import { describe, it, expect } from 'vitest';
import {
	BAND_MAX,
	REST_COLOR,
	buildBand,
	formatQuantity,
	halfValueCount,
	rankColor,
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

describe('rankColor', () => {
	// El puesto es un orden, no una categoría: lo dice la luminosidad, no el
	// matiz. Con un solo matiz no hay pareja de colores que confundir.
	it('aclara el color según el puesto, sin repetir ninguno', () => {
		const ramp = Array.from({ length: 6 }, (_, i) => rankColor(i, 6));

		expect(new Set(ramp).size).toBe(6);
		expect(ramp.every((color) => color.includes('71)'))).toBe(true);
	});

	it('no divide por cero con un solo puesto', () => {
		expect(rankColor(0, 1)).toBe(rankColor(0, 1));
		expect(rankColor(0, 1)).toContain('oklch(');
	});
});

describe('buildBand', () => {
	const many = (n: number) =>
		Array.from({ length: n }, (_, i) =>
			holding({
				ticker: `A${i}`,
				marketValue: String((n - i) * 100),
				percent: 100 / n
			})
		);

	it('ordena de mayor a menor y da a cada franja su color, sin repetir', () => {
		const band = buildBand(toAssetHoldingRows(many(4)));

		expect(band.map((s) => s.label)).toEqual(['A0', 'A1', 'A2', 'A3']);
		expect(new Set(band.map((s) => s.color)).size).toBe(4);
	});

	// Los anchos son la barra: si no suman 100 queda un hueco al final que se
	// lee como dinero que no está en ninguna parte.
	it('reparte los anchos hasta sumar cien', () => {
		const total = buildBand(toAssetHoldingRows(many(7))).reduce((sum, s) => sum + s.width, 0);

		expect(total).toBeCloseTo(100, 6);
	});

	// El ancho sale del valor y el peso viene del backend: son cosas distintas
	// y no se calculan la una desde la otra.
	it('imprime el peso del backend aunque el ancho lo reparta él', () => {
		const rows = toAssetHoldingRows([
			holding({ ticker: 'A', marketValue: '300', percent: 20 }),
			holding({ ticker: 'B', marketValue: '100', percent: 6.7 })
		]);
		const band = buildBand(rows);

		expect(band.map((s) => s.width)).toEqual([75, 25]);
		expect(band.map((s) => s.percent)).toEqual([20, 6.7]);
	});

	// Pasado un puñado de franjas la cola es un serrucho de menos de un píxel:
	// se agrupa, y el detalle queda en la lista de al lado.
	it('agrupa la cola cuando hay más activos de los que caben', () => {
		const band = buildBand(toAssetHoldingRows(many(20)));
		const rest = band.at(-1);

		expect(band).toHaveLength(BAND_MAX + 1);
		expect(rest?.label).toBe('Resto');
		expect(rest?.color).toBe(REST_COLOR);
		expect(rest?.assets).toBe(20 - BAND_MAX);
	});

	it('no crea un «Resto» para un único activo sobrante', () => {
		const band = buildBand(toAssetHoldingRows(many(BAND_MAX + 1)));

		expect(band).toHaveLength(BAND_MAX + 1);
		expect(band.every((s) => s.assets === 1)).toBe(true);
	});

	it('el «Resto» suma los valores y los pesos de la cola', () => {
		const rows = toAssetHoldingRows([
			holding({ ticker: 'A', marketValue: '1000', percent: 50 }),
			holding({ ticker: 'B', marketValue: '600', percent: 30 }),
			holding({ ticker: 'C', marketValue: '300', percent: 15 }),
			holding({ ticker: 'D', marketValue: '100', percent: 5 })
		]);
		const rest = buildBand(rows, 2).at(-1);

		expect(rest?.label).toBe('Resto');
		expect(rest?.value).toBe(400);
		expect(rest?.percent).toBe(20);
		expect(rest?.assets).toBe(2);
	});

	// Una posición vendida entera vale 0 y no es una franja: dibujarla mete en
	// la barra un ancho de cero que no se ve y una entrada que no dice nada.
	it('deja fuera lo que no vale nada', () => {
		const rows = toAssetHoldingRows([
			holding({ ticker: 'A', marketValue: '100', percent: 100 }),
			holding({ ticker: 'B', marketValue: '0', percent: 0 })
		]);

		expect(buildBand(rows).map((s) => s.label)).toEqual(['A']);
	});

	it('sin activos no hay barra', () => {
		expect(buildBand([])).toEqual([]);
	});
});

describe('halfValueCount', () => {
	// La lectura de la barra puesta en número: cuántas franjas caben en la
	// mitad izquierda, que es la mitad del dinero.
	it('cuenta los mayores activos que hacen falta para llegar a la mitad', () => {
		const rows = toAssetHoldingRows([
			holding({ ticker: 'A', marketValue: '40' }),
			holding({ ticker: 'B', marketValue: '30' }),
			holding({ ticker: 'C', marketValue: '20' }),
			holding({ ticker: 'D', marketValue: '10' })
		]);

		// 40 no llega a 50; 40 + 30 sí.
		expect(halfValueCount(rows)).toBe(2);
	});

	it('cuenta uno cuando un solo activo pasa de la mitad', () => {
		const rows = toAssetHoldingRows([
			holding({ ticker: 'A', marketValue: '90' }),
			holding({ ticker: 'B', marketValue: '10' })
		]);

		expect(halfValueCount(rows)).toBe(1);
	});

	// No depende del orden en el que lleguen las filas: ordena por su cuenta,
	// igual que la barra.
	it('no depende del orden de entrada', () => {
		const rows = toAssetHoldingRows([
			holding({ ticker: 'B', marketValue: '10' }),
			holding({ ticker: 'A', marketValue: '90' })
		]);

		expect(halfValueCount(rows)).toBe(1);
	});

	it('sin valor no hay mitad que contar', () => {
		expect(halfValueCount([])).toBe(0);
		expect(halfValueCount(toAssetHoldingRows([holding({ marketValue: '0' })]))).toBe(0);
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
