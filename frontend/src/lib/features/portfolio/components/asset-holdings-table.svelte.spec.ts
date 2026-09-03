import { page } from 'vitest/browser';
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AssetHoldingsTable from './asset-holdings-table.svelte';
import type { AssetHoldingRow } from '../asset-holdings';

const row: AssetHoldingRow = {
	assetId: 'a1',
	ticker: 'AAPL',
	name: 'Apple Inc.',
	assetType: 'stock',
	typeLabel: 'Acciones',
	quantity: 42,
	marketPrice: 214.35,
	currency: 'USD',
	value: 9002.7,
	percent: 37.5,
	portfolios: 2,
	priceSource: 'own',
	fxConverted: true
};

const props = {
	displayCurrency: 'USD',
	formatValue: (v: number) => `$${v.toFixed(2)}`,
	onGoToPortfolios: () => {}
};

describe('asset-holdings-table.svelte', () => {
	it('lists the asset with its type, units and weight', async () => {
		render(AssetHoldingsTable, { ...props, rows: [row] });

		await expect.element(page.getByText('AAPL')).toBeInTheDocument();
		await expect.element(page.getByText('Apple Inc.')).toBeInTheDocument();
		await expect.element(page.getByText('Acciones')).toBeInTheDocument();
		await expect.element(page.getByText('42', { exact: true })).toBeInTheDocument();
		await expect.element(page.getByText('37.5%')).toBeInTheDocument();
	});

	// Cuántos portafolios comparten el activo es lo que esta vista sabe y el
	// detalle de cada portafolio no puede decir.
	it('says in how many portfolios the asset is held', async () => {
		render(AssetHoldingsTable, { ...props, rows: [row] });

		await expect.element(page.getByRole('cell', { name: '2', exact: true })).toBeInTheDocument();
	});

	// Sin precio de mercado no hay número que represente al activo: cada
	// entrada pagó el suyo. Un 0 se leería como un activo que no vale nada.
	it('marks a position carried at cost instead of printing a zero price', async () => {
		render(AssetHoldingsTable, {
			...props,
			rows: [{ ...row, marketPrice: null, priceSource: 'cost' }]
		});

		await expect.element(page.getByText('a coste')).toBeInTheDocument();
		await expect.element(page.getByText('—')).toBeInTheDocument();
	});

	// Un importe sin tasa va a valor nominal y mezcla monedas con el resto de
	// la columna: hay que decirlo, no presentarlo como una cifra limpia.
	it('flags a value no rate could convert', async () => {
		render(AssetHoldingsTable, { ...props, rows: [{ ...row, fxConverted: false }] });

		await expect.element(page.getByText('sin convertir')).toBeInTheDocument();
	});

	// Aquí no hay un portafolio al que agregar —la vista los atraviesa todos—,
	// así que la salida del estado vacío es elegir uno.
	it('sends a user with nothing to their portfolios', async () => {
		const onGoToPortfolios = vi.fn();
		render(AssetHoldingsTable, { ...props, rows: [], onGoToPortfolios });

		await expect.element(page.getByText('Todavía no hay nada que listar')).toBeInTheDocument();
		await page.getByRole('button', { name: 'Ir a mis portafolios' }).click();
		expect(onGoToPortfolios).toHaveBeenCalled();
	});
});
