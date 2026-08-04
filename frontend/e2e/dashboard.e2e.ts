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
		await expect(chart).toHaveAttribute('aria-valuetext', /valor \$/);
		await expect(page.getByText(/^Invertido \$/)).toBeVisible();
		await expect(page.getByText(/^Ganancia [+−]\$/)).toBeVisible();
	});

	test('la asignación de activos enlaza leyenda y porción', async ({ page }) => {
		await login(page);

		const legendEntry = page.getByRole('button', { name: /stock/ });
		await expect(legendEntry).toBeVisible();

		await legendEntry.click();
		await expect(legendEntry).toHaveAttribute('aria-pressed', 'true');
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
	});
});
