import { expect, test } from '@playwright/test';
import { TEST_PORTFOLIO_ID, login } from './helpers';

test.describe('portfolio detail', () => {
	test('renders the portfolio with its holdings', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}`);

		await expect(page.getByRole('heading', { level: 1 })).toContainText('Cartera Principal');
		await expect(page.getByText('AAPL').first()).toBeVisible();
		await expect(page.getByText('Moderado').first()).toBeVisible();
	});

	test('adds an entry through the add-asset form', async ({ page }) => {
		await login(page);
		await page.goto(`/dashboard/portfolios/${TEST_PORTFOLIO_ID}/add`);

		// Platform.
		await page.selectOption('select[name="platformId"]', { label: 'Broker Demo' });

		// Asset combobox: focus triggers the suggestion fetch (via /api/assets).
		await page.click('#asset-search');
		const suggestion = page.getByRole('option').filter({ hasText: 'AAPL' });
		await expect(suggestion).toBeVisible();
		await suggestion.click();

		// Purchase details; the date picker defaults to today.
		await page.fill('input[name="quantity"]', '5');
		await page.fill('input[name="purchasePrice"]', '100');

		await page.getByRole('button', { name: 'Agregar Activo' }).click();

		// The action redirects back to the portfolio detail on success.
		await page.waitForURL(`**/dashboard/portfolios/${TEST_PORTFOLIO_ID}`);
		await expect(page.getByRole('heading', { level: 1 })).toContainText('Cartera Principal');
	});
});
