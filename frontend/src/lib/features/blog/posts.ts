/**
 * Los artículos del blog, leídos de `src/content/blog/*.md`.
 *
 * **La carga es perezosa a propósito.** `import.meta.glob` sin `eager` y
 * `await import('marked')` hacen que Vite deje cada artículo y el propio
 * conversor de Markdown en chunks aparte. Quien llama a estas funciones es un
 * `+page.server.ts` prerenderizado, o sea que todo esto corre en el build y al
 * navegador solo le llega el JSON del artículo que está viendo.
 *
 * Si alguien pone `eager: true` aquí, o importa `marked` arriba del archivo,
 * nada falla: simplemente el bundle del cliente empieza a llevar el texto de
 * todos los artículos y 40 KB de conversor. Lo comprueba `pnpm test:e2e`
 * («el cliente no carga marked») antes de que llegue a producción.
 */

import { dev } from '$app/environment';
import {
	headingId,
	readingMinutes,
	sortByDate,
	visiblePosts,
	hasTag,
	tagSlug,
	type Post,
	type PostMeta
} from './blog';
import { splitFrontmatter } from './frontmatter';
import { postFrontmatterSchema } from './schemas';

/** Un importador por archivo; ninguno se ejecuta hasta que se le llama. */
const sources = import.meta.glob('/src/content/blog/*.md', {
	query: '?raw',
	import: 'default'
}) as Record<string, () => Promise<string>>;

/** El build lee los artículos muchas veces; el disco, una. */
let cache: Promise<Post[]> | null = null;

/** `/src/content/blog/por-que-existe-finexia.md` → `por-que-existe-finexia`. */
function slugOf(path: string): string {
	return path.slice(path.lastIndexOf('/') + 1).replace(/\.md$/, '');
}

/**
 * Markdown a HTML, con un `id` en cada encabezado para poder enlazar a una
 * sección concreta del artículo.
 */
async function renderMarkdown(body: string): Promise<string> {
	const { marked } = await import('marked');

	const renderer = new marked.Renderer();
	renderer.heading = ({ text, depth }) =>
		`<h${depth} id="${headingId(text)}">${marked.parseInline(text)}</h${depth}>\n`;

	return await marked.parse(body, { renderer, gfm: true });
}

/** Un archivo del contenido convertido en artículo, o un error que nombra cuál. */
async function readPost(path: string, load: () => Promise<string>): Promise<Post> {
	const slug = slugOf(path);
	const { data, body } = splitFrontmatter(await load());

	const parsed = postFrontmatterSchema.safeParse(data);
	if (!parsed.success) {
		const detail = parsed.error.issues.map((issue) => `${issue.path.join('.')}: ${issue.message}`);
		throw new Error(`Cabecera inválida en src/content/blog/${slug}.md — ${detail.join('; ')}`);
	}

	return {
		...parsed.data,
		slug,
		readingMinutes: readingMinutes(body),
		html: await renderMarkdown(body)
	};
}

/** Todos los artículos publicables, ordenados. Los borradores solo en `dev`. */
function readAll(): Promise<Post[]> {
	cache ??= Promise.all(Object.entries(sources).map(([path, load]) => readPost(path, load))).then(
		(posts) => sortByDate(visiblePosts(posts, dev))
	);

	return cache;
}

/** Sin el cuerpo: lo que necesitan el índice, las etiquetas y el sitemap. */
function toMeta({ html: _html, ...meta }: Post): PostMeta {
	return meta;
}

/** El índice del blog, del más nuevo al más viejo. */
export async function listPosts(): Promise<PostMeta[]> {
	return (await readAll()).map(toMeta);
}

/** Un artículo por su slug, o `null` si no existe. */
export async function getPost(slug: string): Promise<Post | null> {
	return (await readAll()).find((post) => post.slug === slug) ?? null;
}

/** Los artículos de una etiqueta, en el mismo orden que el índice. */
export async function listPostsByTag(slug: string): Promise<PostMeta[]> {
	return (await listPosts()).filter((post) => hasTag(post, slug));
}

/** Las rutas que el prerender tiene que visitar en `/blog/[slug]`. */
export async function listSlugs(): Promise<string[]> {
	return (await readAll()).map((post) => post.slug);
}

/** Las rutas que el prerender tiene que visitar en `/blog/tag/[tag]`. */
export async function listTagSlugs(): Promise<string[]> {
	const slugs = new Set<string>();
	for (const post of await readAll()) {
		for (const tag of post.tags) slugs.add(tagSlug(tag));
	}
	return [...slugs].sort();
}
