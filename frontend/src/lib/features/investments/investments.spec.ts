import { describe, it, expect } from 'vitest';
import { investmentStore } from './state/investments.svelte';
import { findInvestmentProduct, fromStoredInvestment, getRiskColor } from './investments';

describe('findInvestmentProduct', () => {
	it('resolves a product from the mock catalogue', () => {
		expect(findInvestmentProduct('1')?.name).toBe('Fondo Crecimiento Tecnológico');
	});

	it('falls back to the shared store for products created in this session', () => {
		const id = investmentStore.addInvestment({
			name: 'Bono Verde',
			description: 'Renta fija sostenible',
			type: 'Bonos',
			category: 'Energía Renovable',
			riskLevel: 'Bajo',
			expectedROI: 6.5,
			horizon: 12,
			minimumInvestment: 500,
			status: 'Activo'
		});

		const product = findInvestmentProduct(id);
		expect(product?.name).toBe('Bono Verde');
		// Los campos que el formulario no recoge caen a sus valores por defecto.
		expect(product?.currentROI).toBe(0);
		expect(product?.investors).toBe(0);
	});

	it('returns null for an unknown or empty id', () => {
		expect(findInvestmentProduct('no-existe')).toBeNull();
		expect(findInvestmentProduct('')).toBeNull();
	});
});

describe('fromStoredInvestment', () => {
	it('derives highlights and a maturity date from the horizon', () => {
		const detail = fromStoredInvestment({
			id: 'x1',
			name: 'Fondo X',
			description: 'desc',
			type: 'Fondos',
			category: 'Tecnología',
			riskLevel: 'Alto',
			expectedROI: 12,
			horizon: 6,
			minimumInvestment: 1000,
			status: 'Activo'
		});

		expect(detail.highlights).toContain('Tipo de instrumento: Fondos');
		expect(detail.highlights).toContain('Categoría: Tecnología');
		expect(detail.highlights.some((h) => h.startsWith('Inversión mínima:'))).toBe(true);
		expect(new Date(detail.maturityDate) > new Date(detail.startDate)).toBe(true);
	});

	it('omits the minimum-investment highlight when it is zero', () => {
		const detail = fromStoredInvestment({
			id: 'x2',
			name: 'Fondo Y',
			description: 'desc',
			type: 'ETF',
			category: 'Oro',
			riskLevel: 'Medio',
			expectedROI: 4,
			horizon: 3,
			minimumInvestment: 0,
			status: 'Activo'
		});

		expect(detail.highlights.some((h) => h.startsWith('Inversión mínima:'))).toBe(false);
	});
});

describe('getRiskColor', () => {
	it('maps each risk level to its theme colour', () => {
		expect(getRiskColor('Bajo')).toBe('var(--green)');
		expect(getRiskColor('Medio')).toBe('var(--amber-light)');
		expect(getRiskColor('Alto')).toBe('var(--amber)');
		expect(getRiskColor('Muy Alto')).toBe('var(--red)');
		expect(getRiskColor('Desconocido')).toBe('var(--text)');
	});
});
