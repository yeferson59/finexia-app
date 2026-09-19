import { error } from '@sveltejs/kit';
import { collectTags, listPosts, listPostsByTag, listTagSlugs } from '$lib/features/blog';
import type { EntryGenerator, PageServerLoad } from './$types';

export const entries: EntryGenerator = async () => (await listTagSlugs()).map((tag) => ({ tag }));

export const load: PageServerLoad = async ({ params }) => {
	const posts = await listPostsByTag(params.tag);
	if (posts.length === 0) error(404, 'No hay artículos con esa etiqueta.');

	// Las etiquetas de la navegación son todas, no solo las de esta página: es
	// la barra por la que se salta de una a otra.
	const tags = collectTags(await listPosts());

	return { posts, tags, tag: tags.find((item) => item.slug === params.tag) ?? null };
};
