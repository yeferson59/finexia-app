import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PurchaseFields from './portfolio-entry-purchase-fields.svelte';
import type { Asset } from '$lib/api/types';

// Novo Nordisk: cotiza en coronas danesas, una moneda que no está en la lista
// de conversión de la app. Es el caso que rompía el coste cuando el formulario
// mandaba USD fijo.
const novo: Asset = {
	id: '5b2d6a4c-2a3b-4c5d-8e9f-0a1b2c3d4e5f',
	ticker: 'NOVO-B.CO',
	name: 'Novo Nordisk B A/S',
	assetType: 'stock',
	currency: 'DKK',
	exchange: 'CPH',
	currentPrice: { value: '298.25', currency: 'DKK' },
	priceUpdatedAt: null
};

const props = {
	asset: novo,
	quantity: '0.201065',
	purchasePrice: '866.60',
	purchaseDate: '2024-04-11',
	totalValue: 174.24,
	formatCurrency: (value: number) => `kr ${value.toFixed(2)}`
};

describe('portfolio-entry-purchase-fields.svelte', () => {
	it('ofrece la moneda del activo aunque no esté en la lista de conversión', async () => {
		render(PurchaseFields, { ...props, costCurrency: 'DKK' });

		const select = page.getByLabelText('Moneda de la compra');
		await expect.element(select).toHaveValue('DKK');
		await expect.element(page.getByText('Precio por unidad en DKK')).toBeInTheDocument();
	});

	it('avisa cuando la moneda elegida no es la de cotización del activo', async () => {
		render(PurchaseFields, { ...props, costCurrency: 'USD' });

		await expect
			.element(page.getByText('NOVO-B.CO cotiza en DKK', { exact: false }))
			.toBeInTheDocument();
	});
});
