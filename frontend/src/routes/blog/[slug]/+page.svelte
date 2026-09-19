<script lang="ts">
	import { resolve } from '$app/paths';
	import {
		PostAside,
		PostBody,
		PostHeader,
		PostList,
		blogPostingJsonLd,
		postPath
	} from '$lib/features/blog';
	import { LOCALE, OG_IMAGE, SITE_NAME, absoluteUrl } from '$lib/seo';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const canonical = $derived(absoluteUrl(postPath(data.post.slug)));
	const title = $derived(`${data.post.title} — ${SITE_NAME}`);
	const jsonLd = $derived(JSON.stringify(blogPostingJsonLd(data.post)));
</script>

<svelte:head>
	<title>{title}</title>
	<meta name="description" content={data.post.description} />
	<link rel="canonical" href={canonical} />
	<!-- Un borrador solo existe en `pnpm dev`; aun así, que nadie lo indexe. -->
	<meta name="robots" content={data.post.draft ? 'noindex,nofollow' : 'index,follow'} />

	<!-- Open Graph -->
	<meta property="og:type" content="article" />
	<meta property="og:site_name" content={SITE_NAME} />
	<meta property="og:locale" content={LOCALE} />
	<meta property="og:url" content={canonical} />
	<meta property="og:title" content={data.post.title} />
	<meta property="og:description" content={data.post.description} />
	<meta property="og:image" content={OG_IMAGE} />
	<meta property="article:published_time" content={data.post.date} />
	<meta property="article:author" content={data.post.author} />

	<!-- Twitter -->
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content={data.post.title} />
	<meta name="twitter:description" content={data.post.description} />
	<meta name="twitter:image" content={OG_IMAGE} />

	<!-- Structured data (JSON-LD del artículo, sin entrada de usuario) -->
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html `<script type="application/ld+json">${jsonLd}</scr` + 'ipt>'}
</svelte:head>

<!--
	La rejilla del extracto: el margen a la izquierda con la vuelta al índice,
	la ficha y las secciones; el texto a la derecha.
-->
<article class="lp-wrap post">
	<a class="back" href={resolve('/blog')}>Todos los artículos</a>
	<div class="head">
		<PostHeader post={data.post} />
	</div>
	<PostAside post={data.post} />
	<div class="body">
		<PostBody html={data.post.html} />
	</div>
</article>

{#if data.related.length > 0}
	<section class="lp-wrap more" aria-labelledby="more-title">
		<h2 id="more-title">Sigue leyendo</h2>
		<div class="more-list">
			<PostList posts={data.related} headingLevel={3} />
		</div>
	</section>
{/if}

<style>
	.post {
		display: grid;
		grid-template-columns: var(--blog-margin) minmax(0, 1fr);
		grid-template-areas:
			'back head'
			'aside body';
		column-gap: var(--blog-gap);
		row-gap: 56px;
		padding-top: 72px;
	}

	.back {
		grid-area: back;
		align-self: start;
		padding-top: 12px;
		font-size: 15px;
		color: var(--lp-ink-2);
		text-decoration: underline;
		text-decoration-color: var(--lp-rule);
		text-underline-offset: 4px;
	}

	.back:hover {
		color: var(--lp-ink);
		text-decoration-color: currentColor;
	}

	.head {
		grid-area: head;
	}

	.post > :global(.post-aside) {
		grid-area: aside;
	}

	.body {
		grid-area: body;
		min-width: 0;
		padding-top: 6px;
	}

	/*
	 * «Sigue leyendo» cuelga de un filete en tinta a todo el ancho: marca que el
	 * artículo terminó. Las entradas son las del índice, con su propio margen.
	 */
	.more {
		margin-top: 96px;
	}

	.more h2 {
		margin: 0;
		padding: 20px 0 8px;
		border-top: 2px solid var(--lp-ink);
		font-size: 21px;
		font-stretch: 114%;
		font-weight: 620;
		letter-spacing: -0.01em;
	}

	@media (max-width: 1023px) {
		.post {
			grid-template-columns: minmax(0, 1fr);
			grid-template-areas: 'back' 'head' 'aside' 'body';
			row-gap: 0;
		}
		.back {
			padding-top: 0;
			margin-bottom: 24px;
			justify-self: start;
		}
		.post > :global(.post-aside) {
			margin: 28px 0 40px;
			padding-top: 18px;
			border-top: 1px solid var(--lp-rule);
		}
		.body {
			padding-top: 0;
		}
	}

	@media (max-width: 640px) {
		.post {
			padding-top: 36px;
		}
		.more {
			margin-top: 64px;
		}
	}
</style>
