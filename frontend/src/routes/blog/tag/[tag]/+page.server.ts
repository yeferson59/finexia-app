import { error } from '@sveltejs/kit';
import { collectTags, listPosts, listPostsByTag, listTagSlugs } from '$lib/features/blog';
import type { EntryGenerator, PageServerLoad } from './$types';

// `'auto'` y no el `true` que hereda de `+layout.ts`: si todos los artículos son
// borradores, `entries` sale vacío y SvelteKit aborta el build por una ruta
// prerenderizable que no llegó a prerenderizar. Con `'auto'` se prerenderiza
// lo que haya y la ruta queda además en el servidor, que responde 404.
export const prerender = 'auto';

export const entries: EntryGenerator = async () => (await listTagSlugs()).map((tag) => ({ tag }));

export const load: PageServerLoad = async ({ params }) => {
	const posts = await listPostsByTag(params.tag);
	if (posts.length === 0) error(404, 'No hay artículos con esa etiqueta.');

	// Las etiquetas de la navegación son todas, no solo las de esta página: es
	// la barra por la que se salta de una a otra.
	const tags = collectTags(await listPosts());

	return { posts, tags, tag: tags.find((item) => item.slug === params.tag) ?? null };
};
