import { collectTags, listPosts } from '$lib/features/blog';
import type { PageServerLoad } from './$types';

// Corre en el build (`prerender` en `+layout.ts`): lee los `.md`, los convierte
// y deja el índice servido como JSON estático.
export const load: PageServerLoad = async () => {
	const posts = await listPosts();
	return { posts, tags: collectTags(posts) };
};
