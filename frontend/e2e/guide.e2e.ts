import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('guía de usuario', () => {
	test('ofrece el manual para verlo o descargarlo', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/guia');

		await expect(page.getByRole('heading', { name: 'Guía de usuario' })).toBeVisible();
		await expect(page.getByRole('heading', { name: 'Manual de Usuario de Finexia' })).toBeVisible();

		// La descarga apunta al PDF que sirve la aplicación.
		const download = page.getByRole('link', { name: 'Descargar PDF' });
		await expect(download).toHaveAttribute('href', /manual-usuario\.pdf$/);
		await expect(download).toHaveAttribute('download', 'finexia-manual-de-usuario.pdf');
	});

	test('el PDF se incrusta solo cuando se pide', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/guia');

		// No se carga de entrada: son varios megas que quien solo descarga no paga.
		await expect(page.locator('iframe')).toHaveCount(0);

		await page.getByRole('button', { name: 'Ver la guía aquí' }).click();
		await expect(page.locator('iframe')).toHaveAttribute('src', /manual-usuario\.pdf$/);

		await page.getByRole('button', { name: 'Ocultar la guía' }).click();
		await expect(page.locator('iframe')).toHaveCount(0);
	});

	test('el PDF se sirve realmente', async ({ request }) => {
		const response = await request.get('/manual-usuario.pdf');

		expect(response.status()).toBe(200);
		expect(response.headers()['content-type']).toContain('application/pdf');
	});

	test('el menú lateral lleva a la guía', async ({ page }) => {
		await login(page);

		await page.getByRole('link', { name: 'Guía de usuario' }).click();
		await expect(page).toHaveURL(/\/dashboard\/guia$/);
	});
});
