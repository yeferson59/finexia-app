import { error } from '@sveltejs/kit';
import { getPost, listPosts, listSlugs, relatedPosts } from '$lib/features/blog';
import type { EntryGenerator, PageServerLoad } from './$types';

// `'auto'` y no el `true` que hereda de `+layout.ts`: si todos los artículos son
// borradores, `entries` sale vacío y SvelteKit aborta el build por una ruta
// prerenderizable que no llegó a prerenderizar. Con `'auto'` se prerenderiza
// lo que haya y la ruta queda además en el servidor, que responde 404.
export const prerender = 'auto';

// Las rutas que tiene que visitar el prerender. SvelteKit también las
// descubriría rastreando los enlaces del índice; declararlas es lo que asegura
// que un artículo sin enlazar se publique igual.
export const entries: EntryGenerator = async () => (await listSlugs()).map((slug) => ({ slug }));

export const load: PageServerLoad = async ({ params }) => {
	const post = await getPost(params.slug);
	if (!post) error(404, 'Ese artículo no existe.');
	return { post, related: relatedPosts(post, await listPosts()) };
};
