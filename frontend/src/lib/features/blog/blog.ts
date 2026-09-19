/**
 * Helpers puros de la feature `blog`.
 *
 * Aquí no se lee ningún archivo: eso es cosa de `posts.ts`. Esto es lo que se
 * puede probar sin navegador ni disco —cómo se ordena, cómo se cuenta, cómo se
 * escribe una fecha— y lo que usan los componentes del índice y del artículo.
 */

import type { z } from 'zod';
import { SITE_NAME, absoluteUrl } from '$lib/seo';
import type { postFrontmatterSchema } from './schemas';

/** La cabecera de un artículo, ya validada. */
export type PostFrontmatter = z.infer<typeof postFrontmatterSchema>;

/** Un artículo tal y como lo ve el índice: sin su cuerpo. */
export interface PostMeta extends PostFrontmatter {
	/** El nombre del archivo, que es también su URL. */
	slug: string;
	/** Minutos de lectura del cuerpo, redondeados hacia arriba. */
	readingMinutes: number;
}

/** Un artículo entero, con el HTML ya renderizado. */
export interface Post extends PostMeta {
	html: string;
}

/** Una etiqueta del índice: cómo se escribe, cómo se enlaza y cuántos lleva. */
export interface TagSummary {
	slug: string;
	label: string;
	count: number;
}

/** Palabras por minuto de un lector normal leyendo en pantalla. */
const WORDS_PER_MINUTE = 200;

/**
 * La forma de una palabra o una etiqueta en una URL: sin tildes, sin mayúsculas
 * y sin nada que haya que escapar. `guías` y `guias` llevan al mismo sitio.
 */
export function slugify(value: string): string {
	return value
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-+|-+$/g, '');
}

/** El `id` de un encabezado, para poder enlazar a una sección del artículo. */
export const headingId = slugify;

/** Cómo se enlaza una etiqueta. */
export const tagSlug = slugify;

/**
 * Los más nuevos primero; a igualdad de fecha, por título.
 *
 * Dos artículos del mismo día no es raro el día del lanzamiento, y sin el
 * desempate el orden lo decidiría el sistema de archivos.
 */
export function sortByDate<T extends { date: string; title: string }>(posts: T[]): T[] {
	return [...posts].sort(
		(a, b) => b.date.localeCompare(a.date) || a.title.localeCompare(b.title, 'es')
	);
}

/**
 * Los artículos que se enseñan. Un borrador (`draft: true`) solo entra cuando
 * se piden los borradores, que es en `pnpm dev`: el build de producción no lo
 * prerenderiza, ni lo mete en el índice, las etiquetas, el RSS o el sitemap.
 */
export function visiblePosts<T extends Pick<PostMeta, 'draft'>>(
	posts: T[],
	includeDrafts: boolean
): T[] {
	return includeDrafts ? posts : posts.filter((post) => !post.draft);
}

/** Las etiquetas que aparecen en una lista, de la más usada a la menos. */
export function collectTags(posts: Pick<PostMeta, 'tags'>[]): TagSummary[] {
	const counts = new Map<string, TagSummary>();

	for (const post of posts) {
		for (const tag of post.tags) {
			const slug = tagSlug(tag);
			const seen = counts.get(slug);
			if (seen) seen.count += 1;
			else counts.set(slug, { slug, label: tag, count: 1 });
		}
	}

	return [...counts.values()].sort(
		(a, b) => b.count - a.count || a.label.localeCompare(b.label, 'es')
	);
}

/** Si un artículo lleva esa etiqueta, dé igual cómo esté escrita. */
export function hasTag(post: Pick<PostMeta, 'tags'>, slug: string): boolean {
	return post.tags.some((tag) => tagSlug(tag) === slug);
}

/**
 * Minutos de lectura del cuerpo. Nunca cero: un artículo de dos líneas sigue
 * tardando algo en leerse, y «0 min» no es información.
 */
export function readingMinutes(markdown: string): number {
	const words = markdown.trim().split(/\s+/).filter(Boolean).length;
	return Math.max(1, Math.ceil(words / WORDS_PER_MINUTE));
}

/**
 * La fecha en la forma en que se lee en voz alta, con el día sin cero delante:
 * va dentro de una frase, y «el 05 de septiembre» no es como se dice.
 */
export function formatPostDate(iso: string): string {
	// Mediodía UTC: `new Date('2026-09-24')` es medianoche UTC, que en Bogotá
	// todavía es el 23. Con las doce, la fecha es la misma en todo el mundo.
	const date = new Date(`${iso}T12:00:00Z`);
	if (Number.isNaN(date.getTime())) return '—';
	return date.toLocaleDateString('es-CO', { day: 'numeric', month: 'long', year: 'numeric' });
}

/** «4 min de lectura», para poner junto a la fecha. */
export function formatReadingTime(minutes: number): string {
	return `${minutes} min de lectura`;
}

/** La URL pública de un artículo. */
export function postPath(slug: string): string {
	return `/blog/${slug}`;
}

/** La URL pública de una etiqueta. */
export function tagPath(slug: string): string {
	return `/blog/tag/${slug}`;
}

/**
 * El `BlogPosting` de schema.org que va en el `<head>` del artículo: es lo que
 * lee Google para saber que esto es un artículo, de cuándo y de quién.
 */
export function blogPostingJsonLd(post: PostMeta): Record<string, unknown> {
	return {
		'@context': 'https://schema.org',
		'@type': 'BlogPosting',
		headline: post.title,
		description: post.description,
		datePublished: post.date,
		author: { '@type': 'Organization', name: post.author },
		publisher: { '@type': 'Organization', name: SITE_NAME },
		mainEntityOfPage: absoluteUrl(postPath(post.slug)),
		inLanguage: 'es'
	};
}

/** Lo que no puede ir crudo dentro de un XML. */
function escapeXml(value: string): string {
	return value
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;');
}

/**
 * El feed RSS del blog.
 *
 * Lleva el resumen de cada artículo y no su cuerpo: quien sigue el feed decide
 * qué abrir, y el artículo entero se lee en su página, que es donde está
 * maquetado. `pubDate` va en RFC 822 porque es lo que pide RSS 2.0.
 */
export function renderRssFeed(posts: PostMeta[], title: string, description: string): string {
	const self = absoluteUrl('/blog/rss.xml');

	const items = posts
		.map((post) => {
			const url = absoluteUrl(postPath(post.slug));
			return `    <item>
      <title>${escapeXml(post.title)}</title>
      <link>${url}</link>
      <guid isPermaLink="true">${url}</guid>
      <pubDate>${new Date(`${post.date}T12:00:00Z`).toUTCString()}</pubDate>
      <description>${escapeXml(post.description)}</description>
${post.tags.map((tag) => `      <category>${escapeXml(tag)}</category>`).join('\n')}
    </item>`;
		})
		.join('\n');

	return `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>${escapeXml(title)}</title>
    <link>${absoluteUrl('/blog')}</link>
    <description>${escapeXml(description)}</description>
    <language>es</language>
    <atom:link href="${self}" rel="self" type="application/rss+xml" />
${items}
  </channel>
</rss>`;
}
