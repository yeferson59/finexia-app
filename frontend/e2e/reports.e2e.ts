import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('reports', () => {
	test('renders the analytics panels derived from the growth series', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		// Un panel por año: el fixture cubre de junio de 2025 a julio de 2026.
		const calendars = page.getByRole('heading', { name: 'Rentabilidad mensual (%)' });
		await expect(calendars).toHaveCount(2);

		const calendar2026 = page.getByRole('article').filter({ hasText: '2026' }).first();
		await expect(calendar2026.getByText('Acumulado del año')).toBeVisible();
		// El color no es lo único que dice el signo: cada celda lo lleva en su
		// nombre accesible.
		await expect(
			calendar2026.getByRole('img', { name: /^Ene: \+\d+,\d%, positivo$/ })
		).toBeVisible();
		await expect(
			calendar2026.getByRole('img', { name: /^Abr: −?-?\d+,\d%, negativo$/ })
		).toBeVisible();

		// El primer mes de la serie no tiene con qué compararse.
		const calendar2025 = page.getByRole('article').filter({ hasText: '2025' }).first();
		await expect(calendar2025.getByRole('img', { name: 'Jun: sin dato' })).toBeVisible();
	});

	test('computes the risk statistics and the projection from the history', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		// Con trece retornos mensuales ya hay volatilidad que calcular.
		await expect(page.getByRole('heading', { name: 'Estadísticas clave' })).toBeVisible();
		await expect(page.getByText('Max Drawdown')).toBeVisible();
		await expect(page.getByText('N/A')).toHaveCount(0);

		// Y más de medio año de historial, así que la proyección se dibuja.
		await expect(page.getByRole('heading', { name: 'Proyección de crecimiento' })).toBeVisible();
		await expect(
			page.getByText('Proyección disponible con al menos 6 meses de historial.')
		).toHaveCount(0);
		await expect(page.getByRole('table', { name: /Valor proyectado/ })).toBeAttached();
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
