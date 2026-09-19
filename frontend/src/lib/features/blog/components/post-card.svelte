<script lang="ts">
	/*
	 * Una entrada del índice.
	 *
	 * En papel y tinta, no en tarjeta: una fecha pequeña, el titular grande y el
	 * resumen debajo, separados del siguiente por un filete. Es como se lee un
	 * sumario, y deja que el titular —lo único por lo que alguien decide entrar—
	 * se lleve todo el peso.
	 *
	 * El enlace envuelve el titular y no la entrada entera para que las etiquetas
	 * sigan siendo enlaces propios.
	 *
	 * Un borrador solo llega aquí en `pnpm dev`; la marca es para quien escribe,
	 * que así no lo confunde con uno publicado.
	 */
	import { resolve } from '$app/paths';
	import { formatPostDate, formatReadingTime, tagSlug, type PostMeta } from '../blog';

	interface Props {
		post: PostMeta;
	}

	let { post }: Props = $props();
</script>

<article class="post-card">
	<p class="meta">
		{#if post.draft}
			<span class="draft">Borrador</span>
		{/if}
		<time datetime={post.date}>{formatPostDate(post.date)}</time>
		<span aria-hidden="true">·</span>
		<span>{formatReadingTime(post.readingMinutes)}</span>
	</p>

	<h2 class="title">
		<a href={resolve('/blog/[slug]', { slug: post.slug })}>{post.title}</a>
	</h2>

	<p class="description">{post.description}</p>

	<ul class="tags">
		{#each post.tags as tag (tag)}
			<li>
				<a href={resolve('/blog/tag/[tag]', { tag: tagSlug(tag) })}>{tag}</a>
			</li>
		{/each}
	</ul>
</article>

<style>
	.post-card {
		padding-block: 36px;
		border-top: 1px solid var(--lp-rule);
	}

	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin: 0 0 10px;
		font-size: var(--lp-fs-sm);
		color: var(--lp-ink-2);
	}

	.draft {
		padding: 1px 8px;
		border: 1px dashed currentColor;
		border-radius: 999px;
		font-size: 12px;
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	.title {
		margin: 0;
		font-size: clamp(26px, 3.4vw, 36px);
		font-stretch: 112%;
		font-weight: 620;
		line-height: 1.12;
		letter-spacing: -0.02em;
		text-wrap: balance;
	}

	.title a {
		color: var(--lp-ink);
	}

	.title a:hover {
		text-decoration: underline;
		text-underline-offset: 4px;
	}

	.description {
		max-width: 62ch;
		margin: 12px 0 0;
		font-size: var(--lp-fs-lead);
		line-height: 1.55;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin: 18px 0 0;
		padding: 0;
		list-style: none;
	}

	.tags a {
		display: inline-flex;
		padding: 4px 12px;
		border: 1px solid var(--lp-rule);
		border-radius: 999px;
		font-size: 13px;
		color: var(--lp-ink-2);
		transition:
			border-color 0.15s ease,
			color 0.15s ease;
	}

	.tags a:hover {
		border-color: var(--lp-ink);
		color: var(--lp-ink);
	}

	@media (max-width: 640px) {
		.post-card {
			padding-block: 28px;
		}
		.description {
			font-size: 16px;
		}
	}
</style>
