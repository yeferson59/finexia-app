import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PortfolioList from './portfolio-list.svelte';
import { portfolioBarScale, portfolioTotals, toPortfolioRows } from '../portfolio';
import type { PortfolioSummary } from '$lib/api/types';

function summary(over: Partial<PortfolioSummary> = {}): PortfolioSummary {
	return {
		id: over.name ?? 'p1',
		name: 'Cartera Principal',
		description: 'Acciones y ETFs a largo plazo',
		type: 'stocks_etfs',
		baseCurrency: 'USD',
		displayCurrency: 'USD',
		riskName: 'Moderado',
		totalPositions: 5,
		totalCostBase: '1000',
		totalMarketValue: '1200',
		totalGainLoss: '200',
		totalGainLossPct: '20',
		...over
	};
}

/** Monta la lista tal como lo hace la página: filas, escala y totales. */
function mount(summaries: PortfolioSummary[]) {
	const rows = toPortfolioRows(summaries, 'USD');

	return render(PortfolioList, {
		rows,
		totals: portfolioTotals(rows),
		scale: portfolioBarScale(rows),
		displayCurrency: 'USD'
	});
}

describe('portfolio-list.svelte', () => {
	it('lists the portfolio with its risk, value and return', async () => {
		mount([summary()]);

		await expect.element(page.getByRole('link', { name: 'Cartera Principal' })).toBeInTheDocument();
		await expect.element(page.getByText('Moderado')).toBeInTheDocument();
		await expect.element(page.getByText('$1,200.00').first()).toBeInTheDocument();
		await expect.element(page.getByText('+20,00%').first()).toBeInTheDocument();
	});

	// El nombre es un enlace de verdad, no una tarjeta con `onclick`: se abre en
	// otra pestaña, se recorre con el teclado y enseña a dónde lleva.
	it('links the name to the portfolio', async () => {
		mount([summary({ id: 'abc' })]);

		await expect
			.element(page.getByRole('link', { name: 'Cartera Principal' }))
			.toHaveAttribute('href', '/dashboard/portfolios/abc');
	});

	// La tarjeta enseñaba la etiqueta del tipo y escondía lo que el dueño había
	// escrito sobre su propio portafolio.
	it('shows the description its owner wrote, and falls back to the type', async () => {
		mount([
			summary({ name: 'Con descripción', id: 'a' }),
			summary({ name: 'Sin descripción', id: 'b', description: undefined })
		]);

		await expect
			.element(page.getByText('Acciones y ETFs a largo plazo, 5 posiciones'))
			.toBeInTheDocument();
		await expect.element(page.getByText('Acciones & ETFs, 5 posiciones')).toBeInTheDocument();
	});

	// El capital invertido no tenía sitio en la tarjeta: aquí es la barra, y
	// para quien no la ve, el texto que la acompaña.
	it('announces the capital behind the bar', async () => {
		mount([summary()]);

		// Dos veces: junto a la barra de la fila y en el pie, que suma las tres.
		await expect
			.element(page.getByText('Capital invertido: $1,000.00').first())
			.toBeInTheDocument();
	});

	// Se lista —es suyo y tiene que verlo— pero no se suma: su importe está en
	// otra moneda y el total quedaría en ninguna.
	it('keeps an unconvertible portfolio out of the total but still on the list', async () => {
		mount([
			summary({ name: 'A', id: 'a', totalMarketValue: '1200', totalCostBase: '1000' }),
			summary({
				name: 'B',
				id: 'b',
				displayCurrency: undefined,
				baseCurrency: 'COP',
				totalMarketValue: '5000000'
			})
		]);

		await expect.element(page.getByRole('link', { name: 'B' })).toBeInTheDocument();
		await expect.element(page.getByText('en COP, su propia moneda')).toBeInTheDocument();
		// El pie suma solo el convertible.
		await expect.element(page.getByText('1 portafolio, 5 posiciones abiertas')).toBeInTheDocument();
	});

	// Una pantalla vacía es una invitación a actuar: antes era una rejilla sin
	// nada dentro y ni una palabra.
	it('invites a user with no portfolios to create the first one', async () => {
		mount([]);

		await expect.element(page.getByText('Todavía no tienes portafolios')).toBeInTheDocument();
		await expect
			.element(page.getByRole('link', { name: 'Crear el primero' }))
			.toHaveAttribute('href', '/dashboard/portfolios/add');
	});
});
