import { describe, it, expect } from 'vitest';
import {
	ASSET_TYPE_COLORS,
	ASSET_TYPE_FALLBACK_COLOR,
	ASSET_TYPE_LABELS,
	assetTypeColor,
	formatAssetType
} from './asset-type';

/*
 * El vocabulario de `market.AssetType` (backend/internal/market/asset.go).
 * Está escrito a mano a propósito: si el backend añade o renombra una clase de
 * activo, este test es el que lo dice. La versión anterior de estas tablas
 * estaba tecleada con el vocabulario plural de `portfolio.type` y no acertaba
 * ni una clave, así que el donut perdía a la vez la etiqueta y el color.
 */
const ASSET_TYPES = ['stock', 'etf', 'crypto', 'bond', 'cash', 'real_estate', 'commodity', 'other'];

describe('tablas de clases de activo', () => {
	it('cubre exactamente el enum del backend', () => {
		expect(Object.keys(ASSET_TYPE_LABELS).sort()).toEqual([...ASSET_TYPES].sort());
		expect(Object.keys(ASSET_TYPE_COLORS).sort()).toEqual([...ASSET_TYPES].sort());
	});

	it('no repite color entre clases: dos porciones iguales no se distinguen', () => {
		const colors = Object.values(ASSET_TYPE_COLORS);
		expect(new Set(colors).size).toBe(colors.length);
	});

	it('no usa el vocabulario plural de portfolio.type', () => {
		for (const plural of ['stocks', 'etfs', 'bonds', 'cryptos', 'real_estates', 'commodities']) {
			expect(ASSET_TYPE_LABELS[plural]).toBeUndefined();
		}
	});
});

describe('formatAssetType', () => {
	it('traduce las clases conocidas', () => {
		expect(formatAssetType('stock')).toBe('Acciones');
		expect(formatAssetType('real_estate')).toBe('Inmobiliario');
	});

	it('conserva el nombre crudo de una clase que el backend añada', () => {
		expect(formatAssetType('nft')).toBe('nft');
	});
});

describe('assetTypeColor', () => {
	it('da el color de la clase conocida', () => {
		expect(assetTypeColor('stock')).toBe('#d4912a');
	});

	it('cae al color de reserva en una clase desconocida', () => {
		expect(assetTypeColor('nft')).toBe(ASSET_TYPE_FALLBACK_COLOR);
	});
});
