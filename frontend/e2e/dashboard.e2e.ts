import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('dashboard charts', () => {
	test('la gráfica de crecimiento se recorre con el teclado', async ({ page }) => {
		await login(page);

		const chart = page.getByRole('slider', {
			name: 'Recorrer la gráfica de crecimiento del portafolio'
		});
		await expect(chart).toBeVisible();
		await expect(page.getByText(/Pasa el cursor por la gráfica/)).toBeVisible();

		await chart.focus();
		await chart.press('End');

		// El punto señalado se anuncia y además se pinta en el detalle de arriba.
		await expect(chart).toHaveAttribute('aria-valuetext', /Valor de mercado \$/);
		await expect(page.getByText(/^Capital invertido \$/)).toBeVisible();
		await expect(page.getByText(/^Ganancia [+−]\$/)).toBeVisible();
	});

	test('la gráfica se puede leer en rentabilidad en vez de en dinero', async ({ page }) => {
		await login(page);

		const card = page.locator('section[aria-label="Crecimiento del portafolio"]');
		const units = card.getByRole('group', { name: 'Unidad de la gráfica' });
		await expect(units.getByRole('button', { name: 'Valor' })).toHaveAttribute(
			'aria-pressed',
			'true'
		);

		await units.getByRole('button', { name: '%', exact: true }).click();

		// La serie del fixture crece a base de aportes: en dinero la curva sube y
		// en rentabilidad no, que es justo lo que esta vista viene a enseñar.
		const series = card.getByRole('table', {
			name: /Rentabilidad acumulada y ganancia sobre coste/
		});
		await expect(
			series.getByRole('columnheader', { name: 'Rentabilidad acumulada' })
		).toBeAttached();
		await expect(series.getByRole('columnheader', { name: 'Ganancia sobre coste' })).toBeAttached();
		await expect(series.getByRole('cell').first()).toHaveText(/^[+-]?\d+,\d%$/);

		// Y la cifra del periodo va con las otras métricas, no escondida en el SVG.
		await expect(card.getByText('Rentabilidad real · Todo')).toBeVisible();
	});

	test('la gráfica ofrece sus dos series como tabla para el lector de pantalla', async ({
		page
	}) => {
		await login(page);

		const series = page.getByRole('table', {
			name: /Valor de mercado y capital invertido del portafolio/
		});
		await expect(series.getByRole('columnheader', { name: 'Valor de mercado' })).toBeAttached();
		await expect(series.getByRole('columnheader', { name: 'Capital invertido' })).toBeAttached();

		// Cada fila se identifica con la fecha completa: el eje abrevia a mes y
		// año en los rangos largos, y así varias filas se llamarían igual.
		const firstRow = series.getByRole('rowheader').first();
		await expect(firstRow).toHaveText(/^\d{2} de \p{L}+ de \d{4}$/u);
		await expect(firstRow.locator('time')).toHaveAttribute('datetime', /^\d{4}-\d{2}-\d{2}$/);
	});

	test('la asignación de activos enlaza leyenda y porción', async ({ page }) => {
		await login(page);

		// El fixture reparte el patrimonio entre cinco categorías.
		const legendEntry = page.getByRole('button', { name: /^Acciones/ });
		await expect(legendEntry).toBeVisible();

		await legendEntry.click();
		await expect(legendEntry).toHaveAttribute('aria-pressed', 'true');

		// Al fijar una categoría, el centro del donut deja de mostrar el total.
		await expect(page.locator('.hole-label')).toHaveText('ACCIONES');
	});
});

test.describe('dashboard', () => {
	test('renders the main widgets with data from the backend', async ({ page }) => {
		await login(page);

		// Net worth card.
		await expect(page.getByText('Patrimonio Neto')).toBeVisible();

		// Growth and summary sections.
		await expect(page.locator('section[aria-label="Crecimiento del portafolio"]')).toBeVisible();
		await expect(page.locator('section[aria-label="Resumen financiero"]')).toBeVisible();

		// Portfolio summary and recent activity fed by the API fixtures.
		await expect(page.getByText('Cartera Principal').first()).toBeVisible();
		await expect(page.getByText('AAPL').first()).toBeVisible();

		// Los tres portafolios del fixture, con su tipo ya traducido.
		await expect(page.getByRole('link', { name: 'Cripto' })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Reserva' })).toBeVisible();
		await expect(page.getByText('Acciones & ETFs')).toBeVisible();
	});

	// Enseñar la tasa es el motivo de que la aplicación la traiga: una cifra en
	// pesos que nadie puede cuadrar con su banco no sirve de mucho.
	test('enseña la tasa con la que convierte al cambiar de moneda', async ({ page }) => {
		await login(page);

		// En dólares no hay conversión que enseñar, así que la línea no está.
		await expect(page.getByText(/^1\s/)).toBeHidden();

		await page.getByLabel('Moneda de visualización').selectOption('COP');

		await expect(page.getByText('1 USD = 4.123,46 COP')).toBeVisible();
		await expect(page.getByText(/TRM · dolarapi\.com/)).toBeVisible();
	});
});
