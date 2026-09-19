<script lang="ts">
	/*
	 * El sumario: una entrada debajo de otra, sin rejilla.
	 *
	 * Un blog que empieza tiene tres artículos y una lista los muestra mejor que
	 * una parrilla de tarjetas a medio llenar. Cuando no hay ninguno —una
	 * etiqueta que se queda sin artículos tras despublicar uno— lo dice en vez
	 * de dejar el hueco en blanco.
	 */
	import PostCard from './post-card.svelte';
	import type { PostMeta } from '../blog';

	interface Props {
		posts: PostMeta[];
		/** Qué decir cuando la lista está vacía. */
		emptyTitle?: string;
		/** El nivel de cada titular; 3 si la lista cuelga de un `h2`. */
		headingLevel?: 2 | 3;
	}

	let {
		posts,
		emptyTitle = 'Todavía no hay artículos publicados.',
		headingLevel = 2
	}: Props = $props();
</script>

{#if posts.length === 0}
	<p class="empty">{emptyTitle}</p>
{:else}
	<div class="post-list">
		{#each posts as post (post.slug)}
			<PostCard {post} {headingLevel} />
		{/each}
	</div>
{/if}

<style>
	.post-list {
		border-bottom: 1px solid var(--lp-rule);
	}

	.empty {
		margin: 0;
		padding-block: 40px;
		border-top: 1px solid var(--lp-rule);
		font-size: var(--lp-fs-lead);
		color: var(--lp-ink-2);
	}
</style>
