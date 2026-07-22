import { page } from 'vitest/browser';
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PlatformCard from './platform-card.svelte';

const platform = {
	id: 'p1',
	name: 'Interactive Brokers',
	sourceType: 'broker',
	isActive: true,
	investments: 4,
	totalValue: '12500.5'
};

describe('platform-card.svelte', () => {
	it('renders the platform name, type and active status', async () => {
		render(PlatformCard, { platform, onView: () => {} });

		await expect
			.element(page.getByRole('heading', { name: 'Interactive Brokers' }))
			.toBeInTheDocument();
		await expect.element(page.getByText('broker', { exact: true })).toBeInTheDocument();
		await expect.element(page.getByText('Activo', { exact: true })).toBeInTheDocument();
	});

	it('shows the inactive status when the platform is disabled', async () => {
		render(PlatformCard, { platform: { ...platform, isActive: false }, onView: () => {} });

		await expect.element(page.getByText('Inactivo')).toBeInTheDocument();
	});

	it('invokes onView with the platform id when "Ver detalles" is clicked', async () => {
		const onView = vi.fn();
		render(PlatformCard, { platform, onView });

		await page.getByRole('button', { name: `Ver detalles de ${platform.name}` }).click();

		expect(onView).toHaveBeenCalledWith('p1');
	});
});
