<script lang="ts">
	/*
	 * El último artículo, en la banda verde con la que la portada marca lo que
	 * importa. Es lo único llamativo del blog: todo lo demás va en papel.
	 *
	 * Sigue la misma rejilla que las entradas del índice —la fecha al margen,
	 * el titular a la derecha— para que se lea como el primer movimiento del
	 * extracto y no como un anuncio aparte. Solo cambian la escala y el fondo.
	 */
	import { resolve } from '$app/paths';
	import { formatPostDate, formatReadingTime, tagSlug, type PostMeta } from '../blog';

	interface Props {
		post: PostMeta;
	}

	let { post }: Props = $props();
</script>

<article class="featured lp-forest" aria-labelledby="featured-title">
	<div class="lp-wrap grid">
		<p class="when">
			<time datetime={post.date}>{formatPostDate(post.date)}</time>
			<span>{formatReadingTime(post.readingMinutes)}</span>
			{#if post.draft}
				<span class="draft">Borrador</span>
			{/if}
		</p>

		<div class="what">
			<h2 class="title" id="featured-title">
				<a href={resolve('/blog/[slug]', { slug: post.slug })}>{post.title}</a>
			</h2>

			<p class="description">{post.description}</p>

			<ul class="tags" aria-label="Etiquetas">
				{#each post.tags as tag (tag)}
					<li>
						<a href={resolve('/blog/tag/[tag]', { tag: tagSlug(tag) })}>{tag}</a>
					</li>
				{/each}
			</ul>
		</div>
	</div>
</article>

<style>
	.featured {
		padding-block: 72px 80px;
	}

	.grid {
		display: grid;
		grid-template-columns: var(--blog-margin) minmax(0, 1fr);
		column-gap: var(--blog-gap);
	}

	.when {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 2px;
		margin: 0;
		padding-top: 14px;
		font-size: var(--lp-fs-sm);
		line-height: 1.45;
		color: var(--lp-forest-ink-2);
	}

	.when time {
		font-weight: 600;
		color: var(--lp-forest-ink);
	}

	.draft {
		margin-top: 8px;
		padding: 1px 9px;
		border: 1px dashed currentColor;
		border-radius: 999px;
		font-size: 13px;
		font-weight: 600;
	}

	.title {
		max-width: 16ch;
		margin: 0;
		font-size: clamp(40px, 6.2vw, 88px);
		font-stretch: 125%;
		font-weight: 650;
		line-height: 0.96;
		letter-spacing: -0.035em;
		text-wrap: balance;
	}

	.title a {
		color: var(--lp-forest-ink);
		text-decoration: underline;
		text-decoration-color: transparent;
		text-decoration-thickness: 4px;
		text-underline-offset: 10px;
		transition: text-decoration-color 0.18s ease;
	}

	/* El ámbar de la marca, que sobre el verde sí se ve. */
	.title a:hover {
		text-decoration-color: var(--lp-jubilacion);
	}

	.description {
		max-width: 44ch;
		margin: 28px 0 0;
		font-family: var(--blog-serif);
		font-size: clamp(19px, 2vw, 23px);
		font-weight: 350;
		line-height: 1.5;
		color: var(--lp-forest-ink);
		text-wrap: pretty;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 4px 18px;
		margin: 24px 0 0;
		padding: 0;
		list-style: none;
		font-size: var(--lp-fs-sm);
	}

	.tags a {
		color: var(--lp-forest-ink-2);
		text-decoration: underline;
		text-decoration-color: var(--lp-forest-rule);
		text-underline-offset: 3px;
	}

	.tags a:hover {
		color: var(--lp-forest-ink);
		text-decoration-color: currentColor;
	}

	@media (max-width: 760px) {
		.featured {
			padding-block: 40px 48px;
		}
		.grid {
			grid-template-columns: minmax(0, 1fr);
		}
		.when {
			flex-direction: row;
			flex-wrap: wrap;
			align-items: baseline;
			column-gap: 12px;
			margin-bottom: 14px;
			padding-top: 0;
		}
		.draft {
			margin-top: 0;
		}
		.title a {
			text-decoration-thickness: 3px;
			text-underline-offset: 6px;
		}
		.description {
			margin-top: 18px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.title a {
			transition: none;
		}
	}
</style>
