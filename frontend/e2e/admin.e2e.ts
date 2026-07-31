import { expect, test } from '@playwright/test';
import { ADMIN_EMAIL, login } from './helpers';

test.describe('admin', () => {
	test('lists registered users for an admin', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/users');

		await expect(page.getByRole('heading', { name: 'Usuarios registrados' })).toBeVisible();
		await expect(page.getByText('user@finexia.test').first()).toBeVisible();
		await expect(page.getByRole('heading', { name: 'Invitaciones pendientes' })).toBeVisible();
		await expect(page.getByText('espera@finexia.test').first()).toBeVisible();
	});

	test('redirects non-admin users back to the dashboard', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/admin/users');
		await expect(page).toHaveURL(/\/dashboard$/);
	});

	test('lists the shared asset catalogue with its manual price', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/assets');

		const row = page.locator('tr', { hasText: 'AAPL' });
		await expect(row.getByText('Apple Inc.')).toBeVisible();
		await expect(row.getByText('$190.00')).toBeVisible();
		// El precio manual se edita en la propia fila.
		await expect(row.locator('input[name="price"]')).toHaveValue('190.00');
	});

	test('opens the asset create and import forms', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/assets');

		await page.getByRole('button', { name: '+ Nuevo Activo' }).click();
		await expect(page.getByRole('heading', { name: 'Nuevo activo' })).toBeVisible();
		await expect(page.locator('#ticker')).toBeVisible();

		await page.getByRole('button', { name: 'Importar CSV/Excel' }).click();
		await expect(
			page.getByRole('heading', { name: 'Importar activos desde CSV/Excel' })
		).toBeVisible();
		await expect(page.locator('input[type="file"][name="file"]')).toBeVisible();
	});

	test('lists the exchange rates and opens the create form', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/exchange-rates');

		const row = page.locator('tr', { hasText: 'USD/COP' });
		await expect(row.getByText('4,123.456789')).toBeVisible();
		await expect(row.locator('input[name="rate"]')).toHaveValue('4123.456789');

		await page.getByRole('button', { name: '+ Nueva Tasa' }).click();
		await expect(page.getByRole('heading', { name: 'Nueva tasa de cambio' })).toBeVisible();
		await expect(page.locator('#fromCurrency')).toBeVisible();
	});
});
