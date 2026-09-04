import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PlatformDetail from './platform-detail.svelte';
import type { Platform } from '../platforms';

const platform: Platform = {
	id: 'p1',
	name: 'Interactive Brokers',
	description: 'Mi bróker principal',
	sourceType: 'broker',
	isActive: true,
	createdAt: '2026-01-15T00:00:00Z',
	investments: 4,
	totalValue: '12500.5',
	displayCurrency: 'USD',
	positionsUnconverted: 0
};

describe('platform-detail.svelte', () => {
	it('renders the platform header with its readable source type', async () => {
		render(PlatformDetail, { platform });

		await expect
			.element(page.getByRole('heading', { name: 'Interactive Brokers' }))
			.toBeInTheDocument();
		// broker -> "Bróker" via PLATFORM_TYPES
		await expect.element(page.getByText(/Bróker/).first()).toBeInTheDocument();
	});

	it('labels the invested total with the currency the backend reports', async () => {
		render(PlatformDetail, { platform: { ...platform, displayCurrency: 'COP' } });

		await expect.element(page.getByText(/12\.501/)).toBeInTheDocument();
	});

	it('warns when the total still adds positions no rate could convert', async () => {
		render(PlatformDetail, { platform: { ...platform, positionsUnconverted: 1 } });

		await expect.element(page.getByText(/posición sigue/)).toBeInTheDocument();
	});

	it('says what the positions are spread over', async () => {
		render(PlatformDetail, { platform: { ...platform, assets: 3, portfolios: 2 } });

		await expect.element(page.getByText('3 activos · 2 portafolios')).toBeInTheDocument();
	});

	// Una ganancia de cero es lo que informa una plataforma plana y también una
	// cuyas posiciones se valoran al coste contra el que se comparan. El detalle
	// tiene que distinguirlas en vez de enseñar el mismo cero.
	it('explains a gain of zero that comes from positions valued at cost', async () => {
		render(PlatformDetail, {
			platform: {
				...platform,
				investments: 4,
				marketValue: '12500.5',
				gainLoss: '0',
				gainLossPct: 0,
				positionsAtCost: 4
			}
		});

		await expect.element(page.getByText(/precio de mercado guardado/)).toBeInTheDocument();
	});

	it('counts the unpriced positions when only some of them are', async () => {
		render(PlatformDetail, {
			platform: {
				...platform,
				investments: 4,
				marketValue: '13000',
				gainLoss: '499.5',
				gainLossPct: 4,
				positionsAtCost: 1
			}
		});

		await expect.element(page.getByText('1 posición valorada a coste')).toBeInTheDocument();
		// Es un matiz sobre una de las cuatro, no el aviso de que la ganancia
		// entera es un artefacto.
		await expect.element(page.getByText(/precio de mercado guardado/)).not.toBeInTheDocument();
	});

	it('switches to the edit form when "Editar" is clicked', async () => {
		render(PlatformDetail, { platform });

		await page.getByRole('button', { name: 'Editar' }).click();

		await expect
			.element(page.getByRole('heading', { name: 'Editar Plataforma' }))
			.toBeInTheDocument();
	});

	it('opens the delete confirmation modal', async () => {
		render(PlatformDetail, { platform });

		await page.getByRole('button', { name: 'Eliminar' }).click();

		await expect
			.element(page.getByRole('heading', { name: 'Confirmar eliminación' }))
			.toBeInTheDocument();
	});
});
