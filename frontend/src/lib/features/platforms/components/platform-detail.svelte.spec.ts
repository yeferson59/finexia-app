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
		await expect.element(page.getByText('Bróker').first()).toBeInTheDocument();
	});

	it('labels the invested total with the currency the backend reports', async () => {
		render(PlatformDetail, { platform: { ...platform, displayCurrency: 'COP' } });

		await expect.element(page.getByText(/12\.501/)).toBeInTheDocument();
	});

	it('warns when the total still adds positions no rate could convert', async () => {
		render(PlatformDetail, { platform: { ...platform, positionsUnconverted: 1 } });

		await expect.element(page.getByText(/posición sigue/)).toBeInTheDocument();
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
