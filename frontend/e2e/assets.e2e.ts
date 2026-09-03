import { expect, test } from '@playwright/test';
import { login } from './helpers';

/*
 * La vista consolidada de activos: una fila por activo con lo que el usuario
 * tiene de él sumando todos sus portafolios, y la torta con los que más pesan.
 */
test.describe('mis activos', () => {
	test('lists every asset with its type, units and weight', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/assets');

		await expect(page.getByRole('heading', { level: 1 })).toContainText('Mis Activos');

		const table = page.getByRole('table');
		await expect(table).toBeVisible();

		// Una fila por activo del fixture, no una por posición: BTC vive en el
		// portafolio de cripto y aparece una sola vez, con sus unidades.
		const btc = table.getByRole('row').filter({ hasText: 'Bitcoin' });
		await expect(btc).toHaveCount(1);
		await expect(btc).toContainText('Cripto');
		await expect(btc).toContainText('0,15');

		// Las diez posiciones del fixture son diez activos distintos.
		await expect(table.locator('tbody tr')).toHaveCount(10);
	});

	test('the pie names the heaviest assets and reads out the one you point at', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/assets');

		await expect(page.getByRole('heading', { name: 'Concentración por activo' })).toBeVisible();

		// El centro arranca con el total de la cartera.
		const chart = page.locator('svg.pie');
		await expect(chart).toContainText('VALOR TOTAL');

		// La leyenda es una lista de botones: señalar una entrada lleva su
		// activo al centro, que es lo que hace legible un reparto de colores.
		const legendEntry = page.getByRole('button', { name: /AAPL/ });
		await legendEntry.hover();
		await expect(chart).toContainText('AAPL');
	});
});
