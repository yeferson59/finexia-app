import { expect, test } from '@playwright/test';

test.describe('landing page', () => {
	test('renders the hero headline', async ({ page }) => {
		await page.goto('/');
		await expect(page.getByRole('heading', { level: 1 })).toContainText('patrimonio');
	});

	test('shows the waitlist email field', async ({ page }) => {
		await page.goto('/');
		await expect(page.locator('input[type="email"]').first()).toBeVisible();
	});

	test('exposes a canonical link and page title', async ({ page }) => {
		await page.goto('/');
		await expect(page).toHaveTitle(/Finexia/);
		await expect(page.locator('link[rel="canonical"]')).toHaveAttribute(
			'href',
			'https://finexia.me/'
		);
	});
});
