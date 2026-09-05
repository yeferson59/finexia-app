import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PortfolioPositions from './portfolio-positions.svelte';
import { computeTypeBreakdown, type HoldingView } from '../portfolio';

function holding(over: Partial<HoldingView> = {}): HoldingView {
	return {
		symbol: 'AAPL',
		name: 'Apple Inc.',
		assetType: 'stock',
		quantity: 10,
		marketPrice: 150,
		currency: 'USD',
		costBasis: 1000,
		value: 1500,
		gainLoss: 500,
		gainLossPct: 50,
		allocation: 60,
		fxConverted: true,
		...over
	};
}

function mount(holdings: HoldingView[], over: Record<string, unknown> = {}) {
	return render(PortfolioPositions, {
		holdings,
		typeBreakdown: computeTypeBreakdown(holdings),
		topTransaction: null,
		portfolioId: 'p1',
		baseCurrency: 'USD',
		...over
	});
}

describe('portfolio-positions.svelte', () => {
	it('lists each position with its class, weight, value and return', async () => {
		mount([holding()]);

		await expect.element(page.getByRole('link', { name: 'AAPL' })).toBeInTheDocument();
		await expect.element(page.getByText('Apple Inc.')).toBeInTheDocument();
		await expect.element(page.getByText('Acciones')).toBeInTheDocument();
		await expect.element(page.getByText('$1,500.00')).toBeInTheDocument();
		await expect.element(page.getByText('+50,00%')).toBeInTheDocument();
	});

	// La fila era un botón con `aria-label`, que tapaba su contenido: un lector
	// de pantalla decía «ver detalles de AAPL» y ni el valor ni el rendimiento.
	it('links the symbol to the asset detail', async () => {
		mount([holding()]);

		await expect
			.element(page.getByRole('link', { name: 'AAPL' }))
			.toHaveAttribute('href', '/dashboard/portfolios/p1/assets/AAPL');
	});

	// De mayor a menor peso, así que la primera fila es la posición dominante:
	// era una tarjeta llamada «Concentración».
	it('ranks the positions by weight', async () => {
		mount([
			holding({ symbol: 'SMALL', value: 100, allocation: 10 }),
			holding({ symbol: 'BIG', value: 900, allocation: 90 })
		]);

		const symbols = page.getByRole('rowheader');
		await expect.element(symbols.first()).toHaveTextContent('BIG');
	});

	// Dos tarjetas —«mejor activo» y «peor activo»— decían cuál es la primera y
	// cuál la última de esta misma lista ordenada por rendimiento.
	it('names the best and worst performers in one line', async () => {
		mount([
			holding({ symbol: 'WIN', gainLossPct: 30, allocation: 50, value: 500 }),
			holding({ symbol: 'LOSS', gainLossPct: -10, allocation: 50, value: 500 })
		]);

		await expect
			.element(page.getByText(/WIN es la que más ha rendido .*LOSS, la que menos/))
			.toBeInTheDocument();
	});

	// Con una sola posición no hay extremos que nombrar: «la mejor y la peor»
	// serían la misma fila.
	it('says nothing about standouts when there is only one position', async () => {
		mount([holding()]);

		await expect.element(page.getByText(/la que más ha rendido/)).not.toBeInTheDocument();
	});

	// El donut repartía el portafolio entre dos o tres clases: eso cabe en una
	// línea, y la columna «Clase» deja comprobarla fila a fila.
	it('states the asset-class mix in a sentence', async () => {
		mount([
			holding({ symbol: 'AAPL', assetType: 'stock', value: 750, allocation: 75 }),
			holding({ symbol: 'VWCE', assetType: 'etf', value: 250, allocation: 25 })
		]);

		await expect.element(page.getByText('Acciones 75,0%, ETFs 25,0%.')).toBeInTheDocument();
	});

	it('invites a user with an empty portfolio to add the first asset', async () => {
		mount([]);

		await expect
			.element(page.getByText('Este portafolio aún no tiene activos'))
			.toBeInTheDocument();
		await expect
			.element(page.getByRole('link', { name: 'Agregar tu primer activo' }))
			.toHaveAttribute('href', '/dashboard/portfolios/p1/add');
	});
});
