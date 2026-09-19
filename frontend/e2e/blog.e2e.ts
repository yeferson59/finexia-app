import { expect, test } from '@playwright/test';

/*
 * El blog público. No hace falta sesión ni el stub de la API: los artículos son
 * archivos del repositorio y las páginas salen prerenderizadas del build.
 */

test.describe('blog', () => {
	test('el índice lista los artículos publicados', async ({ page }) => {
		await page.goto('/blog');

		await expect(page.getByRole('heading', { level: 1, name: 'Blog' })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Por qué existe Finexia' })).toBeVisible();
		await expect(
			page.getByRole('link', { name: 'Por qué no te pedimos las claves de tus cuentas' })
		).toBeVisible();
		await expect(
			page.getByRole('link', { name: 'Cómo organizar tus portafolios por objetivo' })
		).toBeVisible();
	});

	test('los más nuevos van primero', async ({ page }) => {
		await page.goto('/blog');

		const titles = await page.locator('article .title a').allTextContents();
		expect(titles[0]).toContain('Cómo organizar tus portafolios por objetivo');
	});

	test('desde el índice se entra al artículo y se vuelve', async ({ page }) => {
		await page.goto('/blog');
		await page.getByRole('link', { name: 'Por qué existe Finexia' }).click();

		await expect(page).toHaveURL('/blog/por-que-existe-finexia');
		await expect(
			page.getByRole('heading', { level: 1, name: 'Por qué existe Finexia' })
		).toBeVisible();
		await expect(page.getByText('15 de septiembre de 2026')).toBeVisible();
		await expect(
			page.getByRole('heading', { level: 2, name: 'El problema no es la falta de datos' })
		).toBeVisible();

		await page.getByRole('link', { name: 'Todos los artículos' }).click();
		await expect(page).toHaveURL('/blog');
	});

	test('las etiquetas filtran el índice', async ({ page }) => {
		await page.goto('/blog/tag/seguridad');

		await expect(page.getByRole('heading', { level: 1, name: 'seguridad' })).toBeVisible();
		await expect(
			page.getByRole('link', { name: 'Por qué no te pedimos las claves de tus cuentas' })
		).toBeVisible();
		await expect(page.getByRole('link', { name: 'Por qué existe Finexia' })).toHaveCount(0);
	});

	// La tilde no puede entrar en la URL, pero la etiqueta se sigue leyendo con ella.
	test('una etiqueta con tilde se enlaza sin ella', async ({ page }) => {
		await page.goto('/blog');
		await page.getByRole('link', { name: 'guías' }).first().click();

		await expect(page).toHaveURL('/blog/tag/guias');
		await expect(page.getByRole('heading', { level: 1, name: 'guías' })).toBeVisible();
	});

	test('una etiqueta que nadie usa devuelve 404', async ({ page }) => {
		const response = await page.goto('/blog/tag/inexistente');
		expect(response?.status()).toBe(404);
	});

	test('un artículo que no existe devuelve 404', async ({ page }) => {
		const response = await page.goto('/blog/articulo-que-no-existe');
		expect(response?.status()).toBe(404);
	});

	test('el feed RSS lleva un item por artículo', async ({ page }) => {
		const response = await page.request.get('/blog/rss.xml');

		expect(response.headers()['content-type']).toContain('application/rss+xml');
		const xml = await response.text();
		expect(xml.match(/<item>/g)).toHaveLength(3);
		expect(xml).toContain('<link>https://finexia.me/blog/por-que-existe-finexia</link>');
	});

	test('el sitemap incluye el blog, sus artículos y sus etiquetas', async ({ page }) => {
		const xml = await (await page.request.get('/sitemap.xml')).text();

		expect(xml).toContain('<loc>https://finexia.me/blog</loc>');
		expect(xml).toContain('<loc>https://finexia.me/blog/por-que-existe-finexia</loc>');
		expect(xml).toContain('<loc>https://finexia.me/blog/tag/producto</loc>');
	});

	test('se llega al blog desde la portada y desde el pie', async ({ page }) => {
		await page.goto('/');

		await expect(page.getByRole('link', { name: 'Blog' }).first()).toHaveAttribute('href', '/blog');
		await page.getByRole('link', { name: 'Blog' }).first().click();
		await expect(page).toHaveURL('/blog');
	});

	/*
	 * `posts.ts` carga los artículos de forma perezosa para que el conversor de
	 * Markdown se quede en el servidor. Si alguien lo importa de forma estática
	 * nada falla, solo engorda el bundle: esto es lo que lo nota.
	 */
	test('el cliente no descarga el conversor de Markdown', async ({ page }) => {
		const scripts: string[] = [];
		page.on('response', (response) => {
			if (response.url().includes('/_app/') && response.url().endsWith('.js')) {
				scripts.push(response.url());
			}
		});

		await page.goto('/blog/por-que-existe-finexia');
		await page.waitForLoadState('networkidle');

		const bodies = await Promise.all(
			scripts.map(async (url) => (await page.request.get(url)).text())
		);
		expect(bodies.some((body) => body.includes('marked'))).toBe(false);
	});
});
