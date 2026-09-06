import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('transactions', () => {
	test('lists the user transactions', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/transactions');

		await expect(page.getByText('AAPL').first()).toBeVisible();
		await expect(page.getByText('BTC').first()).toBeVisible();

		// La primera columna dice qué pasó, no un UUID cortado que no se puede
		// buscar ni cuadrar con el extracto del bróker.
		await expect(page.getByRole('columnheader', { name: 'Movimiento' })).toBeVisible();
		await expect(page.getByText(/^TRX-/)).toHaveCount(0);

		// La nota de cada movimiento no se veía en ninguna parte de esta pantalla.
		await expect(page.getByText('Aporte mensual')).toBeVisible();

		// Un precio por unidad por debajo del céntimo —el interés de la cuenta— se
		// escribe entero: redondeado a dos decimales salía «$0.00» junto a su
		// propio total.
		await expect(page.getByText('$0.0021')).toBeVisible();
	});

	test('import wizard reaches the mapping step with a preview', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/transactions/import');

		// Portfolio and platform selects come pre-seeded from the loader.
		await expect(page.locator('#portfolio')).toHaveValue(/.+/);
		await expect(page.locator('#platform')).toHaveValue(/.+/);

		// Uploading a file posts to the preview proxy and moves to the map step.
		await page.setInputFiles('input[type="file"]', {
			name: 'transacciones.csv',
			mimeType: 'text/csv',
			buffer: Buffer.from(
				'Fecha,Tipo,Ticker,Cantidad,Precio\n2026-05-01,buy,AAPL,10,150\n2026-06-01,sell,BTC,0.01,65000\n'
			)
		});

		await expect(page.getByRole('heading', { name: 'Qué es cada columna' })).toBeVisible();
		await expect(page.getByText('6 listas para importar')).toBeVisible();
		// Las filas que no se pueden interpretar se anuncian antes de confirmar,
		// y el motivo va debajo de su propia fila.
		await expect(page.getByText('2 con errores')).toBeVisible();
		await expect(page.getByText('Fecha no reconocida')).toBeVisible();

		// Cada campo enseña los primeros valores de la columna que tiene asignada:
		// es con lo que se comprueba que la columna elegida es la correcta sin
		// bajar a la vista previa.
		const tickerRow = page.locator('li').filter({ has: page.locator('#map-ticker') });
		await expect(tickerRow).toContainText('NVDA');
		await expect(tickerRow).toContainText('AAPL');

		// Las columnas opcionales que el archivo no trae no se pintan llenas de
		// guiones: no se pintan.
		await expect(page.getByRole('columnheader', { name: 'Comisión' })).toHaveCount(0);
	});
});
