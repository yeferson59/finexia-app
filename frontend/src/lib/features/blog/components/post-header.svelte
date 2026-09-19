<script lang="ts">
	/*
	 * La cabecera del artículo: la vuelta al índice, el titular y la ficha.
	 *
	 * El titular manda —es el tamaño del hero de la portada, un escalón por
	 * debajo—, y la fecha, el autor y la lectura van pequeños debajo del filete,
	 * que es donde se miran solo si hacen falta.
	 *
	 * Un borrador solo se ve en `pnpm dev`, y lo dice encima de la ficha.
	 */
	import { resolve } from '$app/paths';
	import { formatPostDate, formatReadingTime, tagSlug, type PostMeta } from '../blog';

	interface Props {
		post: PostMeta;
	}

	let { post }: Props = $props();
</script>

<header class="post-header">
	<a class="back" href={resolve('/blog')}>← Todos los artículos</a>

	<h1>{post.title}</h1>
	<p class="lead">{post.description}</p>

	<div class="byline">
		{#if post.draft}
			<span class="draft">Borrador</span>
		{/if}
		<time datetime={post.date}>{formatPostDate(post.date)}</time>
		<span aria-hidden="true">·</span>
		<span>{post.author}</span>
		<span aria-hidden="true">·</span>
		<span>{formatReadingTime(post.readingMinutes)}</span>
	</div>

	<ul class="tags">
		{#each post.tags as tag (tag)}
			<li>
				<a href={resolve('/blog/tag/[tag]', { tag: tagSlug(tag) })}>{tag}</a>
			</li>
		{/each}
	</ul>
</header>

<style>
	.post-header {
		max-width: 780px;
	}

	.back {
		display: inline-block;
		margin-bottom: 28px;
		font-size: 15px;
		color: var(--lp-ink-2);
	}

	.back:hover {
		color: var(--lp-ink);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	h1 {
		max-width: 20ch;
		margin: 0;
		font-size: clamp(36px, 5vw, 62px);
		font-stretch: 116%;
		font-weight: 630;
		line-height: 1.02;
		letter-spacing: -0.028em;
		text-wrap: balance;
		overflow-wrap: anywhere;
	}

	.lead {
		max-width: 56ch;
		margin: 20px 0 0;
		font-size: clamp(18px, 2.2vw, 21px);
		line-height: 1.5;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	.byline {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-top: 28px;
		padding-top: 20px;
		border-top: 1px solid var(--lp-rule);
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

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin: 14px 0 0;
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
	}

	.tags a:hover {
		border-color: var(--lp-ink);
		color: var(--lp-ink);
	}

	@media (max-width: 640px) {
		.back {
			margin-bottom: 20px;
			font-size: var(--lp-fs-sm);
		}
		.lead {
			margin-top: 14px;
		}
	}
</style>
