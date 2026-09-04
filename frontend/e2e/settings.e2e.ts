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

	test('creates an MCP token, shows the secret once and lists it', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/settings');

		await expect(page.getByRole('heading', { name: 'Acceso para asistentes (MCP)' })).toBeVisible();

		await page.locator('form[action="?/createMcpToken"] input[name="name"]').fill('Claude Desktop');
		await page.getByRole('button', { name: 'Crear token' }).click();

		// El secreto se muestra una sola vez, aquí: no hay endpoint que lo repita.
		// `exact` porque también aparece dentro del ejemplo de configuración, que
		// es el mismo token en otro contexto y no otra aparición que comprobar.
		await expect(page.getByText('fnx_mcp_e2e-secreto', { exact: true })).toBeVisible();

		// Y el token queda en la lista, ya sin secreto.
		await expect(page.getByText('····a3f9')).toBeVisible();

		await page.reload();
		await expect(page.getByText('fnx_mcp_e2e-secreto', { exact: true })).toBeHidden();
		await expect(page.getByText('····a3f9')).toBeVisible();
	});
});
