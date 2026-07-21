import { expect, test } from '@playwright/test';
import { PASSWORD, USER_EMAIL, login } from './helpers';

test.describe('auth', () => {
	test('redirects protected routes to /auth when logged out', async ({ page }) => {
		await page.goto('/dashboard');
		await expect(page).toHaveURL(/\/auth$/);

		await page.goto('/dashboard/settings');
		await expect(page).toHaveURL(/\/auth$/);
	});

	test('rejects invalid credentials with a visible error', async ({ page }) => {
		await page.goto('/auth');
		await page.fill('#login-email', USER_EMAIL);
		await page.fill('#login-password', 'WrongPassword1!');
		await page.getByRole('button', { name: 'Iniciar sesión', exact: true }).click();

		await expect(page.getByRole('alert')).toContainText('Credenciales incorrectas');
		await expect(page).toHaveURL(/\/auth/);
	});

	test('logs in and out', async ({ page }) => {
		await login(page);
		await expect(page).toHaveURL(/\/dashboard$/);

		await page.getByRole('button', { name: 'Cerrar Sesión' }).click();
		await page.waitForURL('**/auth');

		// The session cookies are gone: protected routes bounce back to /auth.
		await page.goto('/dashboard');
		await expect(page).toHaveURL(/\/auth$/);
	});

	test('login form validates before hitting the backend', async ({ page }) => {
		await page.goto('/auth');
		await page.fill('#login-email', USER_EMAIL);
		// Backend contract: password min 8 chars; shorter must fail client/server-side.
		await page.fill('#login-password', 'short');
		await page.getByRole('button', { name: 'Iniciar sesión', exact: true }).click();
		await expect(page).toHaveURL(/\/auth/);
	});

	test('valid password constant matches backend DTO limits', () => {
		// LoginRequestDTO: min=8, max=20. Keep the shared fixture inside those bounds.
		expect(PASSWORD.length).toBeGreaterThanOrEqual(8);
		expect(PASSWORD.length).toBeLessThanOrEqual(20);
	});
});
