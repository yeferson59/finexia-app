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

		// El importe pasa por `formatCurrency`, que da a cada moneda el locale en
		// el que se lee: el dólar en en-US. Antes esta tarjeta lo componía a mano
		// en es-CO y escribía «$342,88» junto a la cifra grande del panel, que
		// decía «$342.88» para el mismo número.
		await expect.element(page.getByText('−$342.88')).toBeInTheDocument();
		await expect.element(page.getByText('-25,61%')).toBeInTheDocument();
		// El +310,11% de crecimiento del valor ya no se presenta como rendimiento.
		expect(document.body.textContent).not.toContain('310,11');
	});

	it('cae al último punto de la serie cuando el backend no manda la ganancia', async () => {
		const { gainLoss, gainLossPct, ...older } = summary;
		void gainLoss;
		void gainLossPct;

		render(PortfolioGrowth, { data: points, summary: older });

		await expect.element(page.getByText('−$342.88')).toBeInTheDocument();
	});

	// El backend convierte la serie entera a una moneda y la nombra. El símbolo
	// no basta —en es-CO el peso y el dólar comparten el "$"—, así que el código
	// tiene que aparecer.
	it('nombra la moneda en la que viene la serie', async () => {
		render(PortfolioGrowth, {
			data: points,
			summary: { ...summary, currency: 'COP', currentValue: '4000000.00' }
		});

		await expect.element(page.getByText('Valor actual en COP')).toBeInTheDocument();
		// El peso no tiene céntimos en el uso corriente, y `formatCurrency` lo sabe.
		await expect.element(page.getByText(/\$\s?4\.000\.000/)).toBeInTheDocument();
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

	// La cifra que la tarjeta no sabía dar: los $1.099 que entraron el 14 de
	// agosto no son rentabilidad, y el −25,61% sobre costo tampoco los descuenta
	// —divide por lo invertido hoy—. La rentabilidad real encadena tramos.
	it('publica la rentabilidad real del periodo, aparte del rendimiento sobre costo', async () => {
		render(PortfolioGrowth, { data: points, summary });

		await expect.element(page.getByText('Rentabilidad real, Todo')).toBeInTheDocument();
		await expect.element(page.getByText('-43,6%')).toBeInTheDocument();
	});

	it('no publica rentabilidad real sin un tramo que medir', async () => {
		render(PortfolioGrowth, { data: [points[0]], summary });

		await expect.element(page.getByText('Rentabilidad real, Todo')).toBeInTheDocument();
		expect(document.body.textContent).not.toContain('%,');
	});
});

describe('portfolio-growth.svelte · vista de porcentaje', () => {
	/** Una cartera que sube un 10% y luego recibe un aporte que no la mueve. */
	const series: GrowthDataPoint[] = [
		{
			date: '2026-01-01',
			totalValue: '1000.00',
			totalCostBase: '1000.00',
			gainLoss: '0',
			gainLossPct: '0',
			netFlow: '0'
		},
		{
			date: '2026-01-02',
			totalValue: '1100.00',
			totalCostBase: '1000.00',
			gainLoss: '100.00',
			gainLossPct: '10.00',
			netFlow: '0'
		},
		{
			date: '2026-01-03',
			totalValue: '2100.00',
			totalCostBase: '2000.00',
			gainLoss: '100.00',
			gainLossPct: '5.00',
			netFlow: '1000.00'
		}
	];

	const seriesSummary: GrowthSummary = {
		firstDate: '2026-01-01',
		initialValue: '1000.00',
		currentValue: '2100.00',
		totalGrowthPct: '110.00',
		gainLoss: '100.00',
		gainLossPct: '5.00',
		currency: 'USD'
	};

	async function switchToPercent() {
		await page.getByRole('button', { name: '%', exact: true }).click();
	}

	// Las cabeceras de la tabla accesible son las que tienen que cambiar: un
	// porcentaje bajo un «Valor de mercado» sería mentira para quien no ve el SVG.
	it('arranca en dinero y conmuta a porcentaje', async () => {
		render(PortfolioGrowth, { data: series, summary: seriesSummary });

		await expect
			.element(page.getByRole('columnheader', { name: 'Valor de mercado' }))
			.toBeInTheDocument();

		await switchToPercent();

		await expect
			.element(page.getByRole('columnheader', { name: 'Rentabilidad acumulada' }))
			.toBeInTheDocument();
		await expect
			.element(page.getByRole('columnheader', { name: 'Ganancia sobre coste' }))
			.toBeInTheDocument();
		await expect
			.element(page.getByRole('columnheader', { name: 'Valor de mercado' }))
			.not.toBeInTheDocument();
	});

	// El aporte del día 3 duplica el saldo y no mueve la rentabilidad: es lo que
	// esta vista existe para enseñar.
	it('dibuja la rentabilidad limpia de aportes en la tabla accesible', async () => {
		render(PortfolioGrowth, { data: series, summary: seriesSummary });
		await switchToPercent();

		const rows = document.querySelectorAll('table tbody tr');
		const cells = [...rows].map((row) =>
			[...row.querySelectorAll('td')].map((cell) => cell.textContent)
		);

		expect(cells).toEqual([
			['0,0%', '0,0%'],
			['+10,0%', '+10,0%'],
			['+10,0%', '+5,0%']
		]);
	});

	it('explica por qué las dos líneas se separan', async () => {
		render(PortfolioGrowth, { data: series, summary: seriesSummary });
		await switchToPercent();

		await expect
			.element(page.getByText('La rentabilidad descuenta aportes y retiros', { exact: false }))
			.toBeInTheDocument();
	});
});
