import { page } from 'vitest/browser';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AssetCombobox from './asset-combobox.svelte';

function asset(ticker: string, name: string) {
	return {
		id: ticker,
		ticker,
		name,
		assetType: 'stock',
		currency: 'COP',
		currentPrice: null,
		priceUpdatedAt: null,
		isCurated: true
	};
}

// El catálogo busca por fragmento en ticker y nombre, como el backend.
function stubCatalog(catalog: ReturnType<typeof asset>[]) {
	vi.stubGlobal(
		'fetch',
		vi.fn(async (url: string) => {
			const q = (new URL(url, 'http://localhost').searchParams.get('search') ?? '').toUpperCase();
			const data = catalog.filter(
				(a) => a.ticker.toUpperCase().includes(q) || a.name.toUpperCase().includes(q)
			);

			return { ok: true, json: async () => ({ success: true, data }) };
		})
	);
}

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('asset-combobox.svelte', () => {
	it('offers to create a ticker the catalogue only resembles', async () => {
		stubCatalog([asset('GEBR', 'Grupo Energético Brasil')]);

		render(AssetCombobox);
		await page.getByRole('textbox').fill('geb');

		await expect.element(page.getByRole('option', { name: 'GEBR' })).toBeInTheDocument();
		await page.getByRole('button', { name: 'Crear GEB' }).click();

		await expect.element(page.getByText('Nuevo activo')).toBeInTheDocument();
	});

	it('does not offer to create a ticker the catalogue already has', async () => {
		stubCatalog([asset('AAPL', 'Apple Inc.'), asset('AAPLX', 'Apple Tracker')]);

		render(AssetCombobox);
		await page.getByRole('textbox').fill('aapl');

		await expect.element(page.getByRole('option', { name: 'AAPLX' })).toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: 'Crear AAPL' })).not.toBeInTheDocument();
	});

	it('still offers to create when nothing matches', async () => {
		stubCatalog([asset('AAPL', 'Apple Inc.')]);

		render(AssetCombobox);
		await page.getByRole('textbox').fill('zzzq');

		await expect.element(page.getByText('No hay ningún activo que se llame')).toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: 'Crear ZZZQ' })).toBeInTheDocument();
	});
});
