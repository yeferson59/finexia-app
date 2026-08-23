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

		// El mes en el que arranca el historial se marca: su cifra es real, pero
		// cubre menos días que un mes entero.
		const calendar2025 = page.getByRole('article').filter({ hasText: '2025' }).first();
		await expect(calendar2025.getByRole('img', { name: /^Jun: .+, mes parcial$/ })).toBeVisible();
		await expect(calendar2025.getByRole('img', { name: 'Ene: sin dato' })).toBeVisible();

		// Y el que está en curso también: la serie del fixture se corta el 28 de
		// julio, así que julio no es comparable con un mes entero tampoco.
		await expect(calendar2026.getByRole('img', { name: /^Jul: .+, mes parcial$/ })).toBeVisible();
		await expect(calendar2026.getByRole('img', { name: /^Abr: .+, mes parcial$/ })).toHaveCount(0);

		// Y el pie deja dicho que las cifras son rendimiento, no saldo.
		await expect(calendar2026.getByText(/no cuentan como rentabilidad/)).toBeVisible();
	});

	test('computes the risk statistics and the projection from the history', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/reports');

		await expect(page.getByRole('heading', { name: 'Estadísticas clave' })).toBeVisible();
		// Los tres bloques, y dentro las métricas que antes no se publicaban.
		// `exact` porque «Riesgo» es prefijo del reporte «Riesgo y volatilidad».
		for (const group of ['Rendimiento', 'Riesgo', 'Historial']) {
			await expect(page.getByRole('heading', { name: group, exact: true })).toBeVisible();
		}
		for (const stat of ['Máxima caída', 'Ratio de Sharpe', 'Capital invertido']) {
			await expect(page.locator('.stat-tile').filter({ hasText: stat })).toBeVisible();
		}

		// Con setenta puntos de historial no queda ninguna métrica sin calcular.
		await expect(page.getByText('N/A')).toHaveCount(0);

		// El Sharpe se publica en gris y con su reparo al lado: es un cociente
		// estimado, y en verde se leía como un sello de calidad.
		const sharpe = page.locator('.stat-tile').filter({ hasText: 'Ratio de Sharpe' });
		await expect(sharpe.locator('dd')).toHaveClass(/neutral/);
		await expect(sharpe.getByText(/margen de error/)).toBeVisible();

		// El mejor y el peor mes salen de meses enteros: junio de 2025 abre el
		// historial y julio de 2026 sigue en curso, así que ninguno compite.
		for (const label of ['Mejor mes', 'Peor mes']) {
			const tile = page.locator('.stat-tile').filter({ hasText: label });
			await expect(tile.locator('dd')).not.toHaveText(/Jun 2025|Jul 2026/);
		}

		// La rentabilidad del periodo es un porcentaje calculado, no la variación
		// del saldo: la serie del fixture crece a base de aportes.
		const period = page.locator('.stat-tile').filter({ hasText: 'Rentabilidad del periodo' });
		await expect(period.locator('dd')).toHaveText(/^[+-]?\d+,\d%$/);

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
