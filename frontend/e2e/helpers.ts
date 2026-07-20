import type { Page } from '@playwright/test';

/** Fixture ids shared with e2e/mocks/mock-api.mjs. */
export const TEST_PORTFOLIO_ID = '11111111-1111-4111-8111-111111111111';
export const TEST_PLATFORM_ID = '33333333-3333-4333-8333-333333333333';

export const USER_EMAIL = 'user@finexia.test';
export const ADMIN_EMAIL = 'admin@finexia.test';
export const PASSWORD = 'Password123!';

/** Logs in through the real /auth form and waits for the dashboard. */
export async function login(page: Page, email: string = USER_EMAIL): Promise<void> {
	await page.goto('/auth');
	await page.fill('#login-email', email);
	await page.fill('#login-password', PASSWORD);
	await page.getByRole('button', { name: 'Iniciar sesión', exact: true }).click();
	await page.waitForURL('**/dashboard');
}
