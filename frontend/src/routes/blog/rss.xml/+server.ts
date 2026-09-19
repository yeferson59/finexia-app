import { listPosts, renderRssFeed } from '$lib/features/blog';
import { BLOG_DESCRIPTION, BLOG_TITLE } from '$lib/seo';
import type { RequestHandler } from './$types';

/*
 * El feed no se prerenderiza, al contrario que el resto del blog.
 *
 * Un archivo prerenderizado lo sirve el servidor estático, que decide el
 * `Content-Type` por la extensión: `/blog/rss.xml` saldría como `text/xml` y la
 * cabecera de aquí abajo no llegaría a nadie. Generándolo en cada petición sale
 * con su tipo correcto, igual que `sitemap.xml`, y la `s-maxage` de una hora
 * hace que la CDN lo pida una vez por hora como mucho.
 */
export const prerender = false;

export const GET: RequestHandler = async () => {
	const xml = renderRssFeed(await listPosts(), BLOG_TITLE, BLOG_DESCRIPTION);

	return new Response(xml, {
		headers: {
			'Content-Type': 'application/rss+xml; charset=utf-8',
			'Cache-Control': 'max-age=0, s-maxage=3600'
		}
	});
};
