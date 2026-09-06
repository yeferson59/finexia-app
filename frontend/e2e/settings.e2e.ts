import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('settings', () => {
	test('opens with what the account is and how it is protected', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/settings');

		// Los dos grupos abren diciendo cómo está lo suyo: sin esto había que
		// recorrer la página entera para saber si la 2FA estaba puesta.
		await expect(page.getByRole('heading', { name: 'Tu perfil' })).toBeVisible();
		await expect(page.getByText(/Tu cuenta está a nombre de/)).toBeVisible();

		await expect(page.getByRole('heading', { name: 'Cómo entras' })).toBeVisible();
		await expect(page.getByText(/la verificación en dos pasos está desactivada/)).toBeVisible();

		await expect(page.getByRole('heading', { name: 'Lo que tienes conectado' })).toBeVisible();

		// El correo se dice, no se edita: era un campo de formulario desactivado.
		await expect(page.locator('input[name="email"]')).toHaveCount(0);
		// Y «Apariencia» ya no existe: solo servía para decir que no hay temas.
		await expect(page.getByRole('heading', { name: 'Apariencia' })).toHaveCount(0);
	});

	test('updates the profile name', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/settings');

		await expect(page.getByRole('heading', { name: 'Foto, nombre y moneda' })).toBeVisible();

		const nameInput = page.locator('form[action="?/updateProfile"] input[name="name"]');
		await nameInput.fill('Usuaria Renombrada');
		await page.getByRole('button', { name: 'Guardar cambios' }).click();

		await expect(page.getByText('Perfil actualizado correctamente.')).toBeVisible();
	});

	test('states the security sections without uppercase badges', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/settings');

		await expect(page.getByRole('heading', { name: 'Sesiones abiertas' })).toBeVisible();
		await expect(page.getByRole('heading', { name: 'Verificación en dos pasos' })).toBeVisible();

		// El estado se dice en una frase, no en una píldora en versalitas.
		await expect(page.getByText('Todavía no la tienes activada.')).toBeVisible();

		// Y la sesión propia se reconoce dentro de la frase que la describe.
		await expect(page.getByText(/el que estás usando/)).toBeVisible();
	});

	test('creates an MCP token, shows the secret once and lists it', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/settings');

		await expect(page.getByRole('heading', { name: 'Asistentes' })).toBeVisible();

		await page.locator('form[action="?/createMcpToken"] input[name="name"]').fill('Claude Desktop');
		await page.getByRole('button', { name: 'Crear token' }).click();

		// El secreto se muestra una sola vez, aquí: no hay endpoint que lo repita.
		// `exact` porque también aparece dentro del ejemplo de configuración, que
		// es el mismo token en otro contexto y no otra aparición que comprobar.
		await expect(page.getByText('fnx_mcp_e2e-secreto', { exact: true })).toBeVisible();

		// Y el token queda en la lista, ya sin secreto, con su estado en prosa.
		await expect(page.getByText('····a3f9')).toBeVisible();
		await expect(page.getByText(/Sin usar todavía\./)).toBeVisible();

		await page.reload();
		await expect(page.getByText('fnx_mcp_e2e-secreto', { exact: true })).toBeHidden();
		await expect(page.getByText('····a3f9')).toBeVisible();
	});
});
