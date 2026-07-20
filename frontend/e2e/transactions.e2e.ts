import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('transactions', () => {
	test('lists the user transactions', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/transactions');

		await expect(page.getByText('AAPL').first()).toBeVisible();
		await expect(page.getByText('BTC').first()).toBeVisible();
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

		await expect(page.getByRole('heading', { name: 'Asigna tus columnas' })).toBeVisible();
		await expect(page.getByText('2 listas para importar')).toBeVisible();
	});
});
