import { expect, test, type Page } from '@playwright/test';

// SEO snapshot of the public pages (Fase 0 of the frontend migration).
// The migration must not change rendered titles, descriptions, canonicals or
// robots directives; these assertions freeze today's values so any regression
// while extracting components fails the suite.

const SITE_URL = 'https://finexia.me';
const DEFAULT_TITLE = 'Finexia — Tu patrimonio en un solo mapa';
const DEFAULT_DESCRIPTION =
	'Registra manualmente dónde tienes tus activos y agrúpalos en los portafolios que tú imaginas, aunque estén en distintas plataformas. Sin conectar cuentas. Lanzamiento 1 oct 2026.';

function meta(page: Page, name: string) {
	return page.locator(`meta[name="${name}"]`);
}

function og(page: Page, property: string) {
	return page.locator(`meta[property="${property}"]`);
}

test.describe('SEO snapshot', () => {
	test('landing page head', async ({ page }) => {
		await page.goto('/');

		await expect(page).toHaveTitle(DEFAULT_TITLE);
		await expect(meta(page, 'description')).toHaveAttribute('content', DEFAULT_DESCRIPTION);
		await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', `${SITE_URL}/`);

		await expect(og(page, 'og:type')).toHaveAttribute('content', 'website');
		await expect(og(page, 'og:site_name')).toHaveAttribute('content', 'Finexia');
		await expect(og(page, 'og:locale')).toHaveAttribute('content', 'es_ES');
		await expect(og(page, 'og:url')).toHaveAttribute('content', `${SITE_URL}/`);
		await expect(og(page, 'og:title')).toHaveAttribute('content', DEFAULT_TITLE);
		await expect(og(page, 'og:image')).toHaveAttribute('content', `${SITE_URL}/og-image.png`);
	});

	const legalPages = [
		{ path: '/privacidad', title: 'Política de Tratamiento de Datos Personales — Finexia' },
		{ path: '/terminos', title: 'Términos y Condiciones — Finexia' },
		{ path: '/cookies', title: 'Aviso de Cookies — Finexia' }
	];

	for (const { path, title } of legalPages) {
		test(`legal page head: ${path}`, async ({ page }) => {
			await page.goto(path);

			await expect(page).toHaveTitle(title);
			await expect(page.locator('link[rel="canonical"]')).toHaveAttribute(
				'href',
				`${SITE_URL}${path}`
			);
			await expect(meta(page, 'robots')).toHaveAttribute('content', 'index,follow');
			await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
		});
	}

	// El blog es superficie pública nueva: mismas garantías que las legales, y
	// un artículo que declara además que es un artículo.
	const blogPages = [
		{ path: '/blog', title: 'Blog — Finexia' },
		{ path: '/blog/bienvenida-a-finexia', title: 'Te damos la bienvenida a Finexia — Finexia' },
		{ path: '/blog/tag/producto', title: 'Artículos sobre producto — Finexia' }
	];

	for (const { path, title } of blogPages) {
		test(`blog page head: ${path}`, async ({ page }) => {
			await page.goto(path);

			await expect(page).toHaveTitle(title);
			await expect(page.locator('link[rel="canonical"]')).toHaveAttribute(
				'href',
				`${SITE_URL}${path}`
			);
			await expect(meta(page, 'robots')).toHaveAttribute('content', 'index,follow');
			await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
		});
	}

	test('a blog post declares itself as an article', async ({ page }) => {
		await page.goto('/blog/bienvenida-a-finexia');

		await expect(og(page, 'og:type')).toHaveAttribute('content', 'article');
		await expect(og(page, 'article:published_time')).toHaveAttribute('content', '2026-09-19');

		const jsonLd = await page.locator('script[type="application/ld+json"]').textContent();
		expect(JSON.parse(jsonLd ?? '{}')).toMatchObject({
			'@type': 'BlogPosting',
			headline: 'Te damos la bienvenida a Finexia',
			mainEntityOfPage: `${SITE_URL}/blog/bienvenida-a-finexia`
		});
	});

	// La otra cara del noindex de abajo: el blog está para que lo indexen.
	test('the blog is indexable', async ({ page }) => {
		const response = await page.goto('/blog');
		expect(response?.headers()['x-robots-tag']).toBeUndefined();
	});

	test('private areas send X-Robots-Tag noindex', async ({ page }) => {
		const authResponse = await page.goto('/auth');
		expect(authResponse?.headers()['x-robots-tag']).toBe('noindex, nofollow');

		// Even the logged-out redirect from /dashboard must carry the directive.
		const dashboardResponse = await page.request.get('/dashboard', { maxRedirects: 0 });
		expect(dashboardResponse.headers()['x-robots-tag']).toBe('noindex, nofollow');
	});
});
