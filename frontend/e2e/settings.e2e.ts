import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('settings', () => {
	test('updates the profile name', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/settings');

		await expect(page.getByRole('heading', { name: 'Perfil' })).toBeVisible();

		const nameInput = page.locator('form[action="?/updateProfile"] input[name="name"]');
		await nameInput.fill('Usuaria Renombrada');
		await page.getByRole('button', { name: 'Guardar perfil' }).click();

		await expect(page.getByText('Perfil actualizado correctamente.')).toBeVisible();
	});

	test('shows security sections with session data', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/settings');

		await expect(page.getByRole('heading', { name: 'Sesiones activas' })).toBeVisible();
		await expect(
			page.getByRole('heading', { name: 'Verificación en dos pasos (2FA)' })
		).toBeVisible();
	});
});
