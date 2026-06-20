import { SITE_URL } from '$lib/seo';
import type { RequestHandler } from './$types';

// Only public, indexable URLs belong here. The dashboard and auth areas are
// private and excluded (also blocked via robots.txt and X-Robots-Tag).
const pages: { path: string; changefreq: string; priority: string }[] = [
	{ path: '/', changefreq: 'daily', priority: '1.0' }
];

export const GET: RequestHandler = () => {
	const lastmod = new Date().toISOString().split('T')[0];

	const urls = pages
		.map(
			({ path, changefreq, priority }) => `  <url>
    <loc>${SITE_URL}${path}</loc>
    <lastmod>${lastmod}</lastmod>
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
