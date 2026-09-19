import { describe, it, expect } from 'vitest';
import {
	blogPostingJsonLd,
	collectTags,
	formatPostDate,
	formatReadingTime,
	hasTag,
	postPath,
	readingMinutes,
	renderRssFeed,
	slugify,
	sortByDate,
	tagPath,
	visiblePosts,
	type PostMeta
} from './blog';

function post(over: Partial<PostMeta> = {}): PostMeta {
	return {
		slug: 'por-que-existe-finexia',
		title: 'Por qué existe Finexia',
		description: 'Tu dinero está repartido en cinco sitios.',
		date: '2026-09-15',
		tags: ['producto'],
		author: 'Equipo Finexia',
		draft: false,
		readingMinutes: 4,
		...over
	};
}

describe('slugify', () => {
	it('quita tildes, mayúsculas y espacios', () => {
		expect(slugify('Guías Prácticas')).toBe('guias-practicas');
	});

	it('no deja guiones sueltos en los extremos', () => {
		expect(slugify('  ¿Y el efectivo?  ')).toBe('y-el-efectivo');
	});

	it('lleva `guías` y `guias` al mismo sitio', () => {
		expect(slugify('guías')).toBe(slugify('guias'));
	});
});

describe('visiblePosts', () => {
	const posts = [post({ slug: 'publicado' }), post({ slug: 'borrador', draft: true })];

	it('deja fuera los borradores en producción', () => {
		expect(visiblePosts(posts, false).map((item) => item.slug)).toEqual(['publicado']);
	});

	it('los enseña cuando se piden, que es en `pnpm dev`', () => {
		expect(visiblePosts(posts, true).map((item) => item.slug)).toEqual(['publicado', 'borrador']);
	});
});

describe('sortByDate', () => {
	it('pone los más nuevos primero', () => {
		const ordered = sortByDate([
			post({ slug: 'viejo', date: '2026-09-15' }),
			post({ slug: 'nuevo', date: '2026-09-19' })
		]);

		expect(ordered.map((item) => item.slug)).toEqual(['nuevo', 'viejo']);
	});

	it('desempata por título cuando comparten fecha', () => {
		const ordered = sortByDate([
			post({ slug: 'b', title: 'Zeta', date: '2026-09-15' }),
			post({ slug: 'a', title: 'Alfa', date: '2026-09-15' })
		]);

		expect(ordered.map((item) => item.slug)).toEqual(['a', 'b']);
	});

	it('no toca la lista que recibe', () => {
		const original = [
			post({ slug: 'a', date: '2026-09-15' }),
			post({ slug: 'b', date: '2026-09-19' })
		];
		sortByDate(original);

		expect(original.map((item) => item.slug)).toEqual(['a', 'b']);
	});
});

describe('collectTags', () => {
	it('cuenta cada etiqueta una vez por artículo', () => {
		const tags = collectTags([post({ tags: ['producto'] }), post({ tags: ['producto', 'guías'] })]);

		expect(tags).toEqual([
			{ slug: 'producto', label: 'producto', count: 2 },
			{ slug: 'guias', label: 'guías', count: 1 }
		]);
	});

	it('junta las que solo se diferencian en la tilde o la mayúscula', () => {
		const tags = collectTags([post({ tags: ['Guías'] }), post({ tags: ['guias'] })]);

		expect(tags).toHaveLength(1);
		expect(tags[0]).toMatchObject({ slug: 'guias', count: 2 });
	});

	it('devuelve una lista vacía cuando no hay artículos', () => {
		expect(collectTags([])).toEqual([]);
	});
});

describe('hasTag', () => {
	it('encuentra la etiqueta por su slug', () => {
		expect(hasTag(post({ tags: ['Guías'] }), 'guias')).toBe(true);
	});

	it('dice que no cuando el artículo no la lleva', () => {
		expect(hasTag(post({ tags: ['producto'] }), 'seguridad')).toBe(false);
	});
});

describe('readingMinutes', () => {
	it('cuenta a doscientas palabras por minuto, redondeando hacia arriba', () => {
		expect(readingMinutes('palabra '.repeat(201))).toBe(2);
	});

	// «0 min de lectura» no es información.
	it('nunca baja de un minuto', () => {
		expect(readingMinutes('Dos palabras')).toBe(1);
	});

	it('trata un artículo vacío como un minuto', () => {
		expect(readingMinutes('   ')).toBe(1);
	});
});

describe('formatPostDate', () => {
	// Sin la hora, `new Date('2026-09-15')` es medianoche UTC, que en Bogotá
	// todavía es el 14: la fecha publicada no puede retroceder un día.
	it('escribe la fecha en largo y no se corre de día', () => {
		expect(formatPostDate('2026-09-15')).toBe('15 de septiembre de 2026');
	});

	it('devuelve un guion si la fecha no es una fecha', () => {
		expect(formatPostDate('ayer')).toBe('—');
	});
});

describe('formatReadingTime', () => {
	it('lo escribe como se lee', () => {
		expect(formatReadingTime(4)).toBe('4 min de lectura');
	});
});

describe('postPath y tagPath', () => {
	it('construyen las URL públicas', () => {
		expect(postPath('por-que-existe-finexia')).toBe('/blog/por-que-existe-finexia');
		expect(tagPath('guias')).toBe('/blog/tag/guias');
	});
});

describe('blogPostingJsonLd', () => {
	it('describe el artículo como un BlogPosting con su URL absoluta', () => {
		expect(blogPostingJsonLd(post())).toMatchObject({
			'@type': 'BlogPosting',
			headline: 'Por qué existe Finexia',
			datePublished: '2026-09-15',
			mainEntityOfPage: 'https://finexia.me/blog/por-que-existe-finexia',
			inLanguage: 'es'
		});
	});
});

describe('renderRssFeed', () => {
	it('lleva un item por artículo, con su enlace absoluto y su fecha RFC 822', () => {
		const xml = renderRssFeed([post()], 'Blog — Finexia', 'Artículos');

		expect(xml).toContain('<link>https://finexia.me/blog/por-que-existe-finexia</link>');
		expect(xml).toContain('<pubDate>Tue, 15 Sep 2026 12:00:00 GMT</pubDate>');
		expect(xml).toContain('<category>producto</category>');
	});

	it('apunta a sí mismo, como pide un feed', () => {
		const xml = renderRssFeed([], 'Blog — Finexia', 'Artículos');

		expect(xml).toContain('href="https://finexia.me/blog/rss.xml"');
	});

	// Un ampersand crudo rompe el feed entero, no solo ese item.
	it('escapa lo que no puede ir crudo en un XML', () => {
		const xml = renderRssFeed([post({ title: 'Acciones & bonos <hoy>' })], 'Blog', 'Artículos');

		expect(xml).toContain('<title>Acciones &amp; bonos &lt;hoy&gt;</title>');
	});
});
