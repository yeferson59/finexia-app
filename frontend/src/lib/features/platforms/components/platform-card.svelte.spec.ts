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
	totalValue: '12500.5',
	displayCurrency: 'USD',
	positionsUnconverted: 0
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

	it('formats the invested total in the currency the backend reports', async () => {
		render(PlatformCard, {
			platform: { ...platform, displayCurrency: 'COP', totalValue: '12500.5' },
			onView: () => {}
		});

		// COP has no minor unit in everyday use, so the amount is whole: a total
		// converted to pesos rendered under a hardcoded "$1,2500.50" was the bug.
		await expect.element(page.getByText(/12\.501/)).toBeInTheDocument();
	});

	it('warns when the total still adds positions no rate could convert', async () => {
		render(PlatformCard, { platform: { ...platform, positionsUnconverted: 2 }, onView: () => {} });

		await expect.element(page.getByText(/posiciones sin tasa/)).toBeInTheDocument();
	});

	it('omits the warning when every position converted', async () => {
		render(PlatformCard, { platform, onView: () => {} });

		await expect.element(page.getByText(/sin tasa/)).not.toBeInTheDocument();
	});

	it('invokes onView with the platform id when "Ver detalles" is clicked', async () => {
		const onView = vi.fn();
		render(PlatformCard, { platform, onView });

		await page.getByRole('button', { name: `Ver detalles de ${platform.name}` }).click();

		expect(onView).toHaveBeenCalledWith('p1');
	});
});
