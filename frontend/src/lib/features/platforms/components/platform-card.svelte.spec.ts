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

describe('platform-card.svelte · métricas', () => {
	const withMetrics = {
		...platform,
		marketValue: '13750.55',
		gainLoss: '1250.05',
		gainLossPct: 10,
		percent: 62.5
	};

	it('muestra la ganancia y su porcentaje sobre lo invertido', async () => {
		render(PlatformCard, { platform: withMetrics, onView: () => {} });

		await expect.element(page.getByText('Ganancia')).toBeInTheDocument();
		await expect.element(page.getByText('+10.00%')).toBeInTheDocument();
	});

	it('dice qué parte de la cuenta vive en esta plataforma', async () => {
		render(PlatformCard, { platform: withMetrics, onView: () => {} });

		// Es lo que hace legible el orden: «la más grande» dice poco hasta que
		// es «la más grande, y tiene el 62,5% del dinero».
		await expect.element(page.getByText('62.5% de la cuenta')).toBeInTheDocument();
	});

	it('marca la pérdida como pérdida', async () => {
		render(PlatformCard, {
			platform: { ...withMetrics, gainLoss: '-830.20', gainLossPct: -6.64 },
			onView: () => {}
		});

		await expect.element(page.getByText('-6.64%')).toBeInTheDocument();
	});

	// Ausente no es cero: un backend anterior a estas métricas no afirma que la
	// plataforma no haya ganado nada, así que la tarjeta no lo afirma tampoco.
	it('calla la ganancia cuando el backend no la manda', async () => {
		render(PlatformCard, { platform, onView: () => {} });

		await expect.element(page.getByText('Invertido')).toBeInTheDocument();
		await expect.element(page.getByText('Ganancia')).not.toBeInTheDocument();
	});
});
