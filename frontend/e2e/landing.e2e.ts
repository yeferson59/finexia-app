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

	test('switches between the dashboard preview tabs', async ({ page }) => {
		await page.goto('/');

		const resumen = page.getByRole('tab', { name: 'Resumen' });
		await expect(resumen).toHaveAttribute('aria-selected', 'true');
		await expect(page.getByText('Tu patrimonio agregado, en una pantalla')).toBeVisible();

		await page.getByRole('tab', { name: 'Reportes' }).click();
		await expect(resumen).toHaveAttribute('aria-selected', 'false');
		await expect(page.getByText('Reportes que puedes llevarte')).toBeVisible();
	});

	test('expands a FAQ answer and reports it to assistive tech', async ({ page }) => {
		await page.goto('/');

		const question = page.getByRole('button', { name: '¿Qué podré hacer desde el panel?' });
		await expect(question).toHaveAttribute('aria-expanded', 'false');

		await question.click();
		await expect(question).toHaveAttribute('aria-expanded', 'true');
		await expect(page.getByText(/Ver tu patrimonio agregado/)).toBeVisible();
	});

	test('offers a login button that reaches the auth page', async ({ page }) => {
		await page.goto('/');

		const login = page.locator('header').getByRole('link', { name: 'Iniciar sesión' });
		await expect(login).toBeVisible();
		await expect(login).toHaveAttribute('href', '/auth');

		await login.click();
		await expect(page).toHaveURL(/\/auth$/);
	});

	test('keeps the login button reachable on small screens', async ({ page }) => {
		await page.setViewportSize({ width: 390, height: 844 });
		await page.goto('/');

		// En móvil la barra reserva su botón para entrar; la lista de espera baja
		// al menú, porque su formulario ya está a la vista en el hero.
		await expect(
			page.locator('header').getByRole('link', { name: 'Iniciar sesión' })
		).toBeVisible();

		await page.getByRole('button', { name: 'Abrir menú' }).click();
		const menu = page.getByRole('navigation', { name: 'Menú de navegación' });
		await expect(menu.getByRole('link', { name: 'Unirme a la lista' })).toBeVisible();
	});

	test('opens the navigation menu on small screens', async ({ page }) => {
		await page.setViewportSize({ width: 390, height: 844 });
		await page.goto('/');

		const burger = page.getByRole('button', { name: 'Abrir menú' });
		await expect(burger).toBeVisible();

		await burger.click();
		const menu = page.getByRole('navigation', { name: 'Menú de navegación' });
		await expect(menu).toBeVisible();
		await expect(menu.getByRole('link', { name: 'Seguridad' })).toBeVisible();
	});
});
