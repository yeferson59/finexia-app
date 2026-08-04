import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('reports', () => {
	test('renders the analytics panels derived from the growth series', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		// Calendario: el primer mes de la serie no tiene con qué compararse.
		await expect(page.getByRole('heading', { name: 'Rentabilidad mensual (%)' })).toBeVisible();
		await expect(page.getByText('2026', { exact: true })).toBeVisible();
		await expect(page.getByRole('img', { name: /\+26,7%, positivo/ })).toBeVisible();

		// Estadísticas: con un solo retorno mensual no hay volatilidad que dar.
		await expect(page.getByRole('heading', { name: 'Estadísticas clave' })).toBeVisible();
		await expect(page.getByText('Max Drawdown')).toBeVisible();
		await expect(page.getByText('N/A')).toBeVisible();

		// Proyección: el historial del stub no llega a medio año.
		await expect(page.getByRole('heading', { name: 'Proyección de crecimiento' })).toBeVisible();
		await expect(
			page.getByText('Proyección disponible con al menos 6 meses de historial.')
		).toBeVisible();
	});

	test('offers the downloadable reports', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		await expect(page.getByRole('heading', { name: 'Resumen mensual' })).toBeVisible();
		await expect(page.getByRole('heading', { name: 'Riesgo y volatilidad' })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Descargar' }).first()).toHaveAttribute(
			'href',
			/\/dashboard\/reports\/download\?type=summary/
		);
	});
});
