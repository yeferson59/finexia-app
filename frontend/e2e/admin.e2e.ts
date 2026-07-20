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
});
