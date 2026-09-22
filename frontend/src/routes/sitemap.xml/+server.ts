import { SITE_URL } from '$lib/seo';
import { collectTags, listPosts, postPath, tagPath } from '$lib/features/blog';
import type { RequestHandler } from './$types';

interface Entry {
	path: string;
	changefreq: string;
	priority: string;
	/** Fecha real de la última modificación; si falta, la de hoy. */
	lastmod?: string;
}

// Only public, indexable URLs belong here. The dashboard and auth areas are
// private and excluded (also blocked via robots.txt and X-Robots-Tag).
const pages: Entry[] = [
	{ path: '/', changefreq: 'daily', priority: '1.0' },
	{ path: '/blog', changefreq: 'weekly', priority: '0.8' },
	{ path: '/apoyar', changefreq: 'monthly', priority: '0.5' },
	{ path: '/privacidad', changefreq: 'yearly', priority: '0.3' },
	{ path: '/terminos', changefreq: 'yearly', priority: '0.3' },
	{ path: '/cookies', changefreq: 'yearly', priority: '0.3' }
];

export const GET: RequestHandler = async () => {
	const today = new Date().toISOString().split('T')[0];
	const posts = await listPosts();

	// Cada artículo con la fecha que declara: un sitemap que dice que todo
	// cambió hoy no le sirve de nada a quien lo lee.
	const postEntries: Entry[] = posts.map((post) => ({
		path: postPath(post.slug),
		changefreq: 'yearly',
		priority: '0.7',
		lastmod: post.date
	}));

	const tagEntries: Entry[] = collectTags(posts).map((tag) => ({
		path: tagPath(tag.slug),
		changefreq: 'weekly',
		priority: '0.4'
	}));

	const urls = [...pages, ...postEntries, ...tagEntries]
		.map(
			({ path, changefreq, priority, lastmod }) => `  <url>
    <loc>${SITE_URL}${path}</loc>
    <lastmod>${lastmod ?? today}</lastmod>
    <changefreq>${changefreq}</changefreq>
    <priority>${priority}</priority>
  </url>`
		)
		.join('\n');

	const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls}
</urlset>`;

	return new Response(xml, {
		headers: {
			'Content-Type': 'application/xml',
			'Cache-Control': 'max-age=0, s-maxage=3600'
		}
	});
};
