import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PlatformDetail from './platform-detail.svelte';

const platform = {
	id: 'p1',
	name: 'Interactive Brokers',
	description: 'Mi bróker principal',
	sourceType: 'broker',
	isActive: true,
	createdAt: '2026-01-15T00:00:00Z',
	updatedAt: '2026-02-01T00:00:00Z',
	investments: 4,
	totalValue: '12500.5'
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
