import { expect, test } from '@playwright/test';
import { login } from './helpers';

test.describe('dashboard', () => {
	test('renders the main widgets with data from the backend', async ({ page }) => {
		await login(page);

		// Net worth card.
		await expect(page.getByText('Patrimonio Neto')).toBeVisible();

		// Growth and summary sections.
		await expect(page.locator('section[aria-label="Crecimiento del portafolio"]')).toBeVisible();
		await expect(page.locator('section[aria-label="Resumen financiero"]')).toBeVisible();

		// Portfolio summary and recent activity fed by the API fixtures.
		await expect(page.getByText('Cartera Principal').first()).toBeVisible();
		await expect(page.getByText('AAPL').first()).toBeVisible();
	});
});
