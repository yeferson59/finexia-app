import { describe, it, expect } from 'vitest';
import { getPost, listPosts, listPostsByTag, listSlugs, listTagSlugs } from './posts';
import { postFrontmatterSchema } from './schemas';

/*
 * Estos specs leen los artículos de verdad de `src/content/blog/`. No hay
 * fixture a propósito: lo que tiene que estar bien es lo que se publica, y una
 * cabecera mal escrita tiene que romper aquí antes de romper el build.
 */

describe('listPosts', () => {
	it('publica los artículos del repositorio', async () => {
		expect((await listPosts()).length).toBeGreaterThan(0);
	});

	it('los devuelve del más nuevo al más viejo', async () => {
		const dates = (await listPosts()).map((post) => post.date);

		expect([...dates].sort((a, b) => b.localeCompare(a))).toEqual(dates);
	});

	it('cada artículo cumple el schema de la cabecera', async () => {
		for (const post of await listPosts()) {
			expect(postFrontmatterSchema.safeParse(post).success).toBe(true);
		}
	});

	it('no incluye el cuerpo: el índice no lo necesita', async () => {
		expect((await listPosts())[0]).not.toHaveProperty('html');
		expect((await listPosts())[0]).not.toHaveProperty('headings');
	});

	it('le pone a cada uno un slug y sus minutos de lectura', async () => {
		for (const post of await listPosts()) {
			expect(post.slug).toMatch(/^[a-z0-9-]+$/);
			expect(post.readingMinutes).toBeGreaterThanOrEqual(1);
		}
	});
});

describe('getPost', () => {
	it('devuelve el artículo con su HTML ya renderizado', async () => {
		const [first] = await listPosts();
		const post = await getPost(first.slug);

		expect(post?.title).toBe(first.title);
		expect(post?.html).toContain('<p>');
	});

	// `posts.ts` le pone un `id` a cada encabezado para poder enlazar a una
	// sección concreta del artículo.
	it('numera los encabezados con un id', async () => {
		const post = await getPost('por-que-existe-finexia');

		expect(post?.html).toMatch(/<h2 id="[a-z0-9-]+">/);
	});

	// El lateral del artículo enlaza a sus secciones: cada entrada del índice
	// tiene que apuntar a un `id` que exista en el HTML.
	it('apunta las secciones del artículo con el id de su encabezado', async () => {
		const post = await getPost('por-que-existe-finexia');

		expect(post?.headings.length).toBeGreaterThan(0);
		for (const heading of post?.headings ?? []) {
			expect(post?.html).toContain(`<h2 id="${heading.id}">`);
			expect(heading.text).not.toMatch(/[*_`]/);
		}
	});

	it('devuelve null cuando el slug no existe', async () => {
		expect(await getPost('no-existe')).toBeNull();
	});
});

describe('listPostsByTag', () => {
	it('filtra por el slug de la etiqueta', async () => {
		const posts = await listPostsByTag('producto');

		expect(posts.length).toBeGreaterThan(0);
		for (const post of posts) {
			expect(post.tags.map((tag) => tag.toLowerCase())).toContain('producto');
		}
	});

	it('devuelve una lista vacía para una etiqueta que nadie usa', async () => {
		expect(await listPostsByTag('inexistente')).toEqual([]);
	});
});

describe('listSlugs y listTagSlugs', () => {
	it('dan al prerender una entrada por artículo', async () => {
		const posts = await listPosts();

		expect(await listSlugs()).toEqual(posts.map((post) => post.slug));
	});

	it('dan las etiquetas sin repetir y ya en forma de URL', async () => {
		const slugs = await listTagSlugs();

		expect(new Set(slugs).size).toBe(slugs.length);
		for (const slug of slugs) expect(slug).toMatch(/^[a-z0-9-]+$/);
	});
});
