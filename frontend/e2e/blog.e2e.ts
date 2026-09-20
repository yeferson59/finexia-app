import { expect, test } from '@playwright/test';

/*
 * El blog público. No hace falta sesión ni el stub de la API: los artículos son
 * archivos del repositorio y las páginas salen prerenderizadas del build.
 *
 * Estos tests corren contra el build de producción, así que ven lo que verá
 * quien entre: solo los artículos publicados. Los que están en `draft: true` no
 * existen aquí, y eso es justo lo que comprueba «un borrador no se publica».
 */

/** El artículo publicado contra el que se miran el índice, la etiqueta y el feed. */
const POST = {
	slug: 'bienvenida-a-finexia',
	title: 'Te damos la bienvenida a Finexia',
	date: '19 de septiembre de 2026',
	heading: 'Qué es Finexia',
	tag: 'producto'
};

/** Un borrador cualquiera: no entra en el índice, ni en las etiquetas, ni en el feed. */
const DRAFT = { slug: 'por-que-existe-finexia', title: 'Por qué existe Finexia' };

test.describe('blog', () => {
	test('el índice lista los artículos publicados', async ({ page }) => {
		await page.goto('/blog');

		await expect(page.getByRole('heading', { level: 1, name: 'El blog de Finexia' })).toBeVisible();
		await expect(page.getByRole('link', { name: POST.title })).toBeVisible();
	});

	/*
	 * El orden del índice lo prueban `blog.spec.ts` (`sortByDate`) y
	 * `posts.spec.ts` sobre los artículos de verdad; aquí lo que importa es que
	 * el artículo abre la página como destacado.
	 */
	test('el más nuevo abre el índice', async ({ page }) => {
		await page.goto('/blog');

		const titles = await page.locator('article .title a').allTextContents();
		expect(titles[0]).toContain(POST.title);
	});

	test('un borrador no se publica', async ({ page }) => {
		await page.goto('/blog');
		await expect(page.getByRole('link', { name: DRAFT.title })).toHaveCount(0);

		const response = await page.goto(`/blog/${DRAFT.slug}`);
		expect(response?.status()).toBe(404);
	});

	test('desde el índice se entra al artículo y se vuelve', async ({ page }) => {
		await page.goto('/blog');
		await page.getByRole('link', { name: POST.title }).click();

		await expect(page).toHaveURL(`/blog/${POST.slug}`);
		await expect(page.getByRole('heading', { level: 1, name: POST.title })).toBeVisible();
		await expect(page.getByText(POST.date)).toBeVisible();
		await expect(page.getByRole('heading', { level: 2, name: POST.heading })).toBeVisible();

		await page.getByRole('link', { name: 'Todos los artículos' }).click();
		await expect(page).toHaveURL('/blog');
	});

	// La etiqueta se enlaza desde el propio artículo del índice. `exact` porque
	// «producto» es también parte de «El producto», que es el enlace a la portada.
	test('las etiquetas filtran el índice', async ({ page }) => {
		await page.goto('/blog');
		await page.getByRole('link', { name: POST.tag, exact: true }).first().click();

		await expect(page).toHaveURL(`/blog/tag/${POST.tag}`);
		await expect(page.getByRole('heading', { level: 1, name: POST.tag })).toBeVisible();
		await expect(page.getByRole('link', { name: POST.title })).toBeVisible();
		await expect(page.getByRole('link', { name: DRAFT.title })).toHaveCount(0);
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
		expect(xml.match(/<item>/g)).toHaveLength(1);
		expect(xml).toContain(`<link>https://finexia.me/blog/${POST.slug}</link>`);
		expect(xml).not.toContain(DRAFT.slug);
	});

	test('el sitemap incluye el blog, sus artículos y sus etiquetas', async ({ page }) => {
		const xml = await (await page.request.get('/sitemap.xml')).text();

		expect(xml).toContain('<loc>https://finexia.me/blog</loc>');
		expect(xml).toContain(`<loc>https://finexia.me/blog/${POST.slug}</loc>`);
		expect(xml).toContain(`<loc>https://finexia.me/blog/tag/${POST.tag}</loc>`);
		expect(xml).not.toContain(DRAFT.slug);
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

		await page.goto(`/blog/${POST.slug}`);
		await page.waitForLoadState('networkidle');

		const bodies = await Promise.all(
			scripts.map(async (url) => (await page.request.get(url)).text())
		);
		expect(bodies.some((body) => body.includes('marked'))).toBe(false);
	});
});
