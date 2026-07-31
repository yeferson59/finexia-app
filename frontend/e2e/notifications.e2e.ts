import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('notifications', () => {
	test('shows the channels with the stored preferences', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/notifications');

		await expect(page.getByRole('heading', { name: 'Correo electrónico' })).toBeVisible();
		// Los toggles llegan del backend: alertas sí, resumen semanal no.
		await expect(page.locator('input[name="emailAlerts"]')).toBeChecked();
		await expect(page.locator('input[name="weeklySummary"]')).not.toBeChecked();

		// El canal en la app está anunciado pero todavía no disponible.
		await expect(page.getByRole('heading', { name: 'Alertas en la app' })).toBeVisible();
		await expect(page.getByText('Próximamente')).toBeVisible();
	});

	test('saves the preferences', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/notifications');

		await page.locator('input[name="weeklySummary"]').check();
		await page.getByRole('button', { name: 'Guardar preferencias' }).click();

		await expect(page.getByText('Preferencias guardadas correctamente.')).toBeVisible();
	});
});
