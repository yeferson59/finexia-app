<script lang="ts">
	/*
	 * Una entrada del índice, escrita como un movimiento del extracto.
	 *
	 * La fecha va en su columna, al margen, y el titular a la derecha: es como
	 * Finexia enseña las transacciones y es como se repasa un archivo, bajando
	 * por las fechas hasta dar con lo que se busca. El titular se lleva el peso,
	 * el resumen va en la serif del texto corrido y las etiquetas cierran en
	 * pequeño.
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
		/** 3 cuando la lista va debajo de un `h2` propio, como «Sigue leyendo». */
		headingLevel?: 2 | 3;
	}

	let { post, headingLevel = 2 }: Props = $props();
</script>

<article class="post-card">
	<p class="when">
		<time datetime={post.date}>{formatPostDate(post.date)}</time>
		<span class="reading">{formatReadingTime(post.readingMinutes)}</span>
		{#if post.draft}
			<span class="draft">Borrador</span>
		{/if}
	</p>

	<div class="what">
		<svelte:element this={`h${headingLevel}`} class="title">
			<a href={resolve('/blog/[slug]', { slug: post.slug })}>{post.title}</a>
		</svelte:element>

		<p class="description">{post.description}</p>

		<ul class="tags" aria-label="Etiquetas">
			{#each post.tags as tag (tag)}
				<li>
					<a href={resolve('/blog/tag/[tag]', { tag: tagSlug(tag) })}>{tag}</a>
				</li>
			{/each}
		</ul>
	</div>
</article>

<style>
	.post-card {
		display: grid;
		grid-template-columns: var(--blog-margin, 220px) minmax(0, 1fr);
		column-gap: var(--blog-gap, 72px);
		padding-block: 36px 40px;
		border-top: 1px solid var(--lp-rule);
	}

	.when {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 2px;
		margin: 0;
		padding-top: 6px;
		font-size: var(--lp-fs-sm);
		line-height: 1.45;
		color: var(--lp-ink-2);
	}

	.when time {
		font-weight: 600;
		color: var(--lp-ink);
	}

	.draft {
		margin-top: 8px;
		padding: 1px 9px;
		border: 1px dashed currentColor;
		border-radius: 999px;
		font-size: 13px;
		font-weight: 600;
	}

	.what {
		max-width: 720px;
	}

	.title {
		margin: 0;
		font-size: clamp(25px, 3vw, 34px);
		font-stretch: 114%;
		font-weight: 620;
		line-height: 1.1;
		letter-spacing: -0.022em;
		text-wrap: balance;
	}

	.title a {
		color: var(--lp-ink);
		text-decoration: underline;
		text-decoration-color: transparent;
		text-decoration-thickness: 2px;
		text-underline-offset: 5px;
		transition: text-decoration-color 0.15s ease;
	}

	.title a:hover {
		text-decoration-color: var(--lp-jubilacion);
	}

	.description {
		max-width: 58ch;
		margin: 12px 0 0;
		font-family: var(--blog-serif);
		font-size: 19px;
		font-weight: 380;
		line-height: 1.55;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 4px 18px;
		margin: 16px 0 0;
		padding: 0;
		list-style: none;
		font-size: var(--lp-fs-sm);
	}

	.tags a {
		color: var(--lp-ink-2);
		text-decoration: underline;
		text-decoration-color: var(--lp-rule);
		text-underline-offset: 3px;
	}

	.tags a:hover {
		color: var(--lp-ink);
		text-decoration-color: currentColor;
	}

	/* Sin sitio para el margen, la fecha sube encima del titular en una línea. */
	@media (max-width: 760px) {
		.post-card {
			grid-template-columns: minmax(0, 1fr);
			padding-block: 28px 32px;
		}
		.when {
			flex-direction: row;
			flex-wrap: wrap;
			align-items: baseline;
			column-gap: 12px;
			margin-bottom: 8px;
			padding-top: 0;
		}
		.draft {
			margin-top: 0;
		}
		.description {
			font-size: 17px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.title a {
			transition: none;
		}
	}
</style>
