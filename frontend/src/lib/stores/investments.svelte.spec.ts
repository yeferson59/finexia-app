import { describe, it, expect } from 'vitest';
import { investmentStore, type NewInvestment } from './investments.svelte';

const draft: NewInvestment = {
	name: 'Bono Corporativo AA',
	description: 'Renta fija de grado de inversión.',
	type: 'Bonos',
	category: 'Renta Fija',
	riskLevel: 'Bajo',
	expectedROI: 6.4,
	horizon: 12,
	minimumInvestment: 1000,
	status: 'Activo'
};

describe('investmentStore', () => {
	it('is seeded with the initial products', () => {
		expect(investmentStore.items.length).toBeGreaterThanOrEqual(3);
	});

	it('looks up a seeded investment by id', () => {
		const found = investmentStore.getById('1');
		expect(found?.name).toBe('Fondo Crecimiento Tecnológico');
	});

	it('returns undefined for an unknown id', () => {
		expect(investmentStore.getById('does-not-exist')).toBeUndefined();
	});

	it('adds a product, returns its id, and persists it in the store', () => {
		const before = investmentStore.items.length;

		const id = investmentStore.addInvestment(draft);

		expect(id).toBeTruthy();
		expect(investmentStore.items.length).toBe(before + 1);

		const stored = investmentStore.getById(id);
		expect(stored).toMatchObject(draft);
		expect(stored?.id).toBe(id);
	});

	it('assigns a distinct id to each added product', () => {
		const first = investmentStore.addInvestment(draft);
		const second = investmentStore.addInvestment(draft);
		expect(first).not.toBe(second);
	});
});
