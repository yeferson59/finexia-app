import { expect, test } from '@playwright/test';
import { login, UNVERIFIED_EMAIL, USER_EMAIL } from './helpers';

test.describe('notifications', () => {
	test('shows the channels with the stored preferences', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/notifications');

		await expect(page.getByRole('heading', { name: 'Correo' })).toBeVisible();
		// A dónde llegan los correos: el dato que solo tiene esta página.
		await expect(page.getByText(`Llegan a ${USER_EMAIL}`)).toBeVisible();
		// Los toggles llegan del backend: alertas sí, resumen semanal no.
		await expect(page.locator('input[name="emailAlerts"]')).toBeChecked();
		await expect(page.locator('input[name="weeklySummary"]')).not.toBeChecked();

		// El canal en la app está anunciado pero todavía no disponible.
		await expect(page.getByRole('heading', { name: 'En la app' })).toBeVisible();
		await expect(page.getByText('Todavía no están disponibles')).toBeVisible();
	});

	test('saves the preferences', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/notifications');

		// La fila entera es la etiqueta de la casilla, no solo el cuadrado.
		await page.getByText('Resumen semanal').click();
		await expect(page.locator('input[name="weeklySummary"]')).toBeChecked();

		await page.getByRole('button', { name: 'Guardar preferencias' }).click();

		await expect(page.getByText('Preferencias guardadas correctamente.')).toBeVisible();
	});

	/* Da igual lo que se marque si la dirección no está confirmada: no sale
	   ningún correo, y esta es la única página desde la que se puede saber. */
	test('warns when the address cannot receive anything yet', async ({ page }) => {
		await login(page, UNVERIFIED_EMAIL);
		await page.goto('/dashboard/notifications');

		await expect(page.getByText(`${UNVERIFIED_EMAIL} todavía no está verificada`)).toBeVisible();
		await expect(page.getByRole('link', { name: 'Pide un enlace nuevo' })).toHaveAttribute(
			'href',
			'/auth/verify-email'
		);

		// Las casillas siguen siendo suyas: la preferencia se guarda igual.
		await expect(page.locator('input[name="emailAlerts"]')).toBeEnabled();
	});
});
