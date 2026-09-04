import { page } from 'vitest/browser';
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PlatformRow from './platform-row.svelte';
import type { Platform, PlatformShare } from '../platforms';

const platform: Platform = {
	id: 'p1',
	name: 'Interactive Brokers',
	description: '',
	sourceType: 'broker',
	isActive: true,
	createdAt: '2026-01-15T00:00:00Z',
	investments: 4,
	totalValue: '12500.5',
	displayCurrency: 'USD',
	positionsUnconverted: 0
};

const entry = (over: Partial<Platform> = {}, share = 62.5): PlatformShare => ({
	platform: { ...platform, ...over },
	share,
	rank: 0
});

describe('platform-row.svelte', () => {
	it('renders the platform name and its readable source type', async () => {
		render(PlatformRow, { entry: entry(), count: 3, onView: () => {} });

		await expect.element(page.getByRole('button', { name: /Interactive Brokers/ })).toBeVisible();
		// El listado enseñaba el `sourceType` crudo del backend —«broker»—
		// mientras el detalle traducía: la misma plataforma con dos nombres.
		await expect.element(page.getByText(/Bróker/)).toBeInTheDocument();
	});

	it('dice qué parte de la cuenta vive en esta plataforma', async () => {
		// Es lo que hace legible el orden: «la más grande» dice poco hasta que
		// es «la más grande, y tiene el 62,5% del dinero».
		render(PlatformRow, { entry: entry(), count: 3, onView: () => {} });

		await expect.element(page.getByText(/62\.5% de la cuenta/)).toBeInTheDocument();
	});

	it('marca sólo las plataformas inactivas, y en gris', async () => {
		render(PlatformRow, { entry: entry({ isActive: false }), count: 3, onView: () => {} });

		await expect.element(page.getByText('Inactiva')).toBeInTheDocument();
	});

	it('no marca nada cuando la plataforma está activa', async () => {
		render(PlatformRow, { entry: entry(), count: 3, onView: () => {} });

		await expect.element(page.getByText('Inactiva')).not.toBeInTheDocument();
	});

	it('formats the invested total in the currency the backend reports', async () => {
		// COP has no minor unit in everyday use, so the amount is whole: a total
		// converted to pesos rendered under a hardcoded "$1,2500.50" was the bug.
		render(PlatformRow, {
			entry: entry({ displayCurrency: 'COP' }),
			count: 3,
			onView: () => {}
		});

		await expect.element(page.getByText(/12\.501/)).toBeInTheDocument();
	});

	it('warns when the total still adds positions no rate could convert', async () => {
		render(PlatformRow, { entry: entry({ positionsUnconverted: 2 }), count: 3, onView: () => {} });

		await expect.element(page.getByText(/posiciones sin tasa/)).toBeInTheDocument();
	});

	it('omits the warning when every position converted', async () => {
		render(PlatformRow, { entry: entry(), count: 3, onView: () => {} });

		await expect.element(page.getByText(/sin tasa/)).not.toBeInTheDocument();
	});

	it('invokes onView with the platform id when the name is clicked', async () => {
		const onView = vi.fn();
		render(PlatformRow, { entry: entry(), count: 3, onView });

		await page.getByRole('button', { name: `Ver detalles de ${platform.name}` }).click();

		expect(onView).toHaveBeenCalledWith('p1');
	});
});

describe('platform-row.svelte · ganancia', () => {
	const withGain = { gainLoss: '1250.05', gainLossPct: 10, marketValue: '13750.55' };

	it('muestra la ganancia y su porcentaje sobre lo invertido', async () => {
		render(PlatformRow, { entry: entry(withGain), count: 3, onView: () => {} });

		await expect.element(page.getByText('+10.00%')).toBeInTheDocument();
	});

	it('marca la pérdida como pérdida', async () => {
		render(PlatformRow, {
			entry: entry({ ...withGain, gainLoss: '-830.20', gainLossPct: -6.64 }),
			count: 3,
			onView: () => {}
		});

		await expect.element(page.getByText('-6.64%')).toBeInTheDocument();
	});

	// Ausente no es cero: un backend anterior a estas métricas no afirma que la
	// plataforma no haya ganado nada, así que la fila no lo afirma tampoco.
	it('calla la ganancia cuando el backend no la manda', async () => {
		render(PlatformRow, { entry: entry(), count: 3, onView: () => {} });

		await expect.element(page.getByText('—')).toBeInTheDocument();
	});
});
