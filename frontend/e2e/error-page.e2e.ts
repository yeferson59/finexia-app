import { expect, test } from '@playwright/test';

test.describe('error page', () => {
	test('explains a missing page and offers the way back home', async ({ page }) => {
		const response = await page.goto('/esta-ruta-no-existe');
		expect(response?.status()).toBe(404);

		await expect(page.getByRole('heading', { level: 1 })).toHaveText('No encontramos esta página');
		await expect(page.getByText('/esta-ruta-no-existe')).toBeVisible();
		// El «Not Found» en inglés de SvelteKit no llega a la pantalla.
		await expect(page.getByText('Not Found')).toHaveCount(0);

		const home = page.getByRole('main').getByRole('link', { name: 'Volver al inicio' });
		await expect(home).toHaveAttribute('href', '/');
	});
});
