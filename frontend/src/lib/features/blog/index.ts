/**
 * Feature `blog` — superficie pública.
 *
 * El blog público de Finexia: los artículos viven en `src/content/blog/*.md`,
 * se convierten en HTML durante el build y se sirven prerenderizados desde
 * `routes/blog/`. Publicar un artículo es añadir un `.md` y desplegar; cómo se
 * escribe está en `docs/BLOG.md`.
 *
 * Las funciones de `posts.ts` solo se llaman desde un `+page.server.ts`: cargan
 * el contenido de forma perezosa para que no acabe en el bundle del cliente.
 */
export { default as PostCard } from './components/post-card.svelte';
export { default as PostList } from './components/post-list.svelte';
export { default as PostHeader } from './components/post-header.svelte';
export { default as PostBody } from './components/post-body.svelte';
export { default as TagNav } from './components/tag-nav.svelte';
export * from './blog';
export * from './posts';
export * from './schemas';
