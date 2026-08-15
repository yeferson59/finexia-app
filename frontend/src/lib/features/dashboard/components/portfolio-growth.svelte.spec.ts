import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import PortfolioGrowth from './portfolio-growth.svelte';
import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

// Una cuenta que empezó pequeña y a la que hoy se le añadieron portafolios: el
// valor se multiplicó por cuatro sin haber ganado nada, y la posición está en
// pérdidas. Es el caso en el que la tarjeta cantaba +310% sobre una cartera que
// perdía $342.
const points: GrowthDataPoint[] = [
	{
		date: '2026-06-28',
		totalValue: '242.89',
		totalCostBase: '240.00',
		gainLoss: '2.89',
		gainLossPct: '1.20'
	},
	{
		date: '2026-08-14',
		totalValue: '996.14',
		totalCostBase: '1339.02',
		gainLoss: '-342.88',
		gainLossPct: '-25.61'
	}
];

const summary: GrowthSummary = {
	firstDate: '2026-06-28',
	initialValue: '242.89',
	currentValue: '996.14',
	totalGrowthPct: '310.11',
	gainLoss: '-342.88',
	gainLossPct: '-25.61'
};

describe('portfolio-growth.svelte', () => {
	it('muestra la ganancia real, no la subida del valor desde el primer snapshot', async () => {
		render(PortfolioGrowth, { data: points, summary });

		await expect.element(page.getByText('−$342,88')).toBeInTheDocument();
		await expect.element(page.getByText('-25,61%')).toBeInTheDocument();
		// El +310,11% de crecimiento del valor ya no se presenta como rendimiento.
		expect(document.body.textContent).not.toContain('310,11');
	});

	it('cae al último punto de la serie cuando el backend no manda la ganancia', async () => {
		const { gainLoss, gainLossPct, ...older } = summary;
		void gainLoss;
		void gainLossPct;

		render(PortfolioGrowth, { data: points, summary: older });

		await expect.element(page.getByText('−$342,88')).toBeInTheDocument();
	});

	// El backend convierte la serie entera a una moneda y la nombra. El símbolo
	// no basta —en es-CO el peso y el dólar comparten el "$"—, así que el código
	// tiene que aparecer.
	it('nombra la moneda en la que viene la serie', async () => {
		render(PortfolioGrowth, {
			data: points,
			summary: { ...summary, currency: 'COP', currentValue: '4000000.00' }
		});

		await expect.element(page.getByText('Valor actual · COP')).toBeInTheDocument();
		await expect.element(page.getByText('$4.000.000,00')).toBeInTheDocument();
	});

	it('avisa de las fechas cuyo total incluye portafolios sin convertir', async () => {
		const withGap = points.map((p, i) => (i === 1 ? { ...p, portfoliosUnconverted: 2 } : p));

		render(PortfolioGrowth, { data: withGap, summary });

		await expect
			.element(page.getByText('Faltan tasas para convertir', { exact: false }))
			.toBeInTheDocument();
	});

	it('no avisa cuando todo se convirtió', async () => {
		render(PortfolioGrowth, { data: points, summary });

		await expect
			.element(page.getByText('Faltan tasas para convertir', { exact: false }))
			.not.toBeInTheDocument();
	});
});
