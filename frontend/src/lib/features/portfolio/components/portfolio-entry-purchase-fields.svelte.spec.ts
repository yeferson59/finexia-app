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

// LVMH: cotiza en euros y se compra desde cuentas en dólares, que es el caso
// que motivó la tasa por transacción.
const lvmh: Asset = {
	id: '7c3e8b5d-1f2a-4b6c-9d0e-3a4b5c6d7e8f',
	ticker: 'MC.FR',
	name: 'LVMH Moet Hennessy Louis Vuitton SE',
	assetType: 'stock',
	currency: 'EUR',
	exchange: 'PAR',
	currentPrice: { value: '429.45', currency: 'EUR' },
	priceUpdatedAt: null
};

const props = {
	asset: novo,
	quantity: '0.201065',
	purchasePrice: '866.60',
	purchaseDate: '2024-04-11',
	fxRate: ''
};

describe('portfolio-entry-purchase-fields.svelte', () => {
	it('ofrece la moneda del activo aunque no esté en la lista de conversión', async () => {
		render(PurchaseFields, { ...props, currency: 'DKK', costCurrency: 'DKK' });

		const select = page.getByLabelText('Moneda de la operación');
		await expect.element(select).toHaveValue('DKK');
		await expect.element(page.getByText('Lo que costó cada una, en DKK.')).toBeInTheDocument();
	});

	it('avisa cuando la moneda elegida no es la de cotización del activo', async () => {
		render(PurchaseFields, { ...props, currency: 'USD', costCurrency: 'USD' });

		await expect
			.element(page.getByText('NOVO-B.CO cotiza en DKK', { exact: false }))
			.toBeInTheDocument();
	});

	// Mientras el interruptor esté apagado no hay nada que decidir: la tasa va
	// oculta en 1 y las dos monedas son la misma, que es el caso de la inmensa
	// mayoría de las compras.
	it('no pide tasa mientras la cuenta liquide en la moneda de la operación', async () => {
		render(PurchaseFields, { ...props, currency: 'DKK', costCurrency: 'DKK' });

		await expect.element(page.getByLabelText('Tasa de la operación')).not.toBeInTheDocument();
	});

	it('pide la tasa al declarar que la cuenta liquidó en otra moneda', async () => {
		render(PurchaseFields, {
			asset: lvmh,
			quantity: '0.0241',
			purchasePrice: '606.60',
			purchaseDate: '2024-12-05',
			currency: 'EUR',
			costCurrency: 'EUR',
			fxRate: ''
		});

		await page.getByText('Mi cuenta liquidó en otra moneda').click();

		await page.getByLabelText('Moneda de la cuenta').selectOptions('USD');

		// La cuenta que sale de esto —el importe que el bróker debitó— la pinta
		// ahora `portfolio-entry-total`, y allí es donde la fija su prueba.
		await expect.element(page.getByLabelText('Tasa de la operación')).toBeInTheDocument();
	});

	// El estado que producía un 400 opaco: el interruptor encendido sin haber
	// cambiado la moneda del precio, así que las dos son la misma y la tasa que
	// se escriba no puede ser cierta. El backend lo rechaza; el formulario tiene
	// que decirlo aquí, y no dejar que se envíe.
	it('no acepta una tasa cuando las dos monedas son la misma', async () => {
		render(PurchaseFields, {
			asset: lvmh,
			quantity: '0.0241',
			purchasePrice: '606.60',
			purchaseDate: '2024-12-05',
			currency: 'USD',
			costCurrency: 'USD',
			fxRate: ''
		});

		await page.getByText('Mi cuenta liquidó en otra moneda').click();

		const rate = page.getByLabelText('Tasa de la operación');
		await expect.element(rate).toBeDisabled();
		await expect
			.element(page.getByText('no hubo conversión y no hay tasa que aplicar', { exact: false }))
			.toBeInTheDocument();
	});
});
