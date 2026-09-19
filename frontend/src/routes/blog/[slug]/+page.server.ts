import { error } from '@sveltejs/kit';
import { getPost, listSlugs } from '$lib/features/blog';
import type { EntryGenerator, PageServerLoad } from './$types';

// Las rutas que tiene que visitar el prerender. SvelteKit también las
// descubriría rastreando los enlaces del índice; declararlas es lo que asegura
// que un artículo sin enlazar se publique igual.
export const entries: EntryGenerator = async () => (await listSlugs()).map((slug) => ({ slug }));

export const load: PageServerLoad = async ({ params }) => {
	const post = await getPost(params.slug);
	if (!post) error(404, 'Ese artículo no existe.');
	return { post };
};
