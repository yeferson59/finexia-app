import { page } from 'vitest/browser';
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AssetHoldingsList from './asset-holdings-list.svelte';
import type { AssetHoldingRow } from '../asset-holdings';

const row: AssetHoldingRow = {
	assetId: 'a1',
	ticker: 'AAPL',
	name: 'Apple Inc.',
	assetType: 'stock',
	typeLabel: 'Acciones',
	sectorLabel: 'Tecnología',
	sectorDetail: '',
	quantity: 42,
	marketPrice: 214.35,
	currency: 'USD',
	value: 9002.7,
	percent: 37.5,
	portfolios: 2,
	priceSource: 'own',
	priceProvider: 'finnhub',
	priceFetchedAt: '2026-09-05T10:00:00Z',
	fxConverted: true
};

const props = {
	maxValue: 9002.7,
	displayCurrency: 'USD',
	formatValue: (v: number) => `$${v.toFixed(2)}`,
	onGoToPortfolios: () => {},
	onOpen: () => {}
};

describe('asset-holdings-list.svelte', () => {
	// Aquí no hay un portafolio al que agregar —la vista los atraviesa todos—,
	// así que la salida del estado vacío es elegir uno.
	it('sends a user with nothing to their portfolios', async () => {
		const onGoToPortfolios = vi.fn();
		render(AssetHoldingsList, { ...props, rows: [], onGoToPortfolios });

		await expect.element(page.getByText('Todavía no hay nada que listar')).toBeInTheDocument();
		await page.getByRole('button', { name: 'Ir a mis portafolios' }).click();
		expect(onGoToPortfolios).toHaveBeenCalled();
	});

	// Una búsqueda sin resultados no es una cartera vacía: invitar a ir a los
	// portafolios diría que no hay nada, cuando solo no hay nada con ese nombre.
	it('tells a search with no match apart from an empty portfolio', async () => {
		render(AssetHoldingsList, { ...props, rows: [row] });

		await page.getByRole('searchbox').fill('zzz');

		await expect.element(page.getByText('Ningún activo se llama así.')).toBeInTheDocument();
		await expect.element(page.getByText('Todavía no hay nada que listar')).not.toBeInTheDocument();
	});
});
