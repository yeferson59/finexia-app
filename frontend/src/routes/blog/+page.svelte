<script lang="ts">
	import { resolve } from '$app/paths';
	import { FeaturedPost, PostList, TagNav } from '$lib/features/blog';
	import { BLOG_DESCRIPTION, BLOG_TITLE, LOCALE, OG_IMAGE, SITE_NAME, absoluteUrl } from '$lib/seo';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const canonical = absoluteUrl('/blog');

	// El más nuevo abre la página en la banda verde; el resto sigue debajo.
	const featured = $derived(data.posts[0]);
	const rest = $derived(data.posts.slice(1));
</script>

<svelte:head>
	<title>{BLOG_TITLE}</title>
	<meta name="description" content={BLOG_DESCRIPTION} />
	<link rel="canonical" href={canonical} />
	<meta name="robots" content="index,follow" />
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content={SITE_NAME} />
	<meta property="og:locale" content={LOCALE} />
	<meta property="og:title" content={BLOG_TITLE} />
	<meta property="og:description" content={BLOG_DESCRIPTION} />
	<meta property="og:url" content={canonical} />
	<meta property="og:image" content={OG_IMAGE} />
	<meta name="twitter:card" content="summary_large_image" />
	<link
		rel="alternate"
		type="application/rss+xml"
		title={BLOG_TITLE}
		href={absoluteUrl('/blog/rss.xml')}
	/>
</svelte:head>

<div class="lp-wrap intro">
	<h1>El blog de Finexia</h1>
	<div class="aside">
		<p class="lead">{BLOG_DESCRIPTION}</p>
		<a class="lp-link rss" href={resolve('/blog/rss.xml')}>Seguir por RSS</a>
	</div>
</div>

{#if featured}
	<FeaturedPost post={featured} />
{:else}
	<div class="lp-wrap">
		<PostList posts={[]} />
	</div>
{/if}

{#if rest.length > 0}
	<section class="lp-wrap archive" aria-labelledby="archive-title">
		<div class="archive-head">
			<h2 id="archive-title">Anteriores</h2>
			<TagNav tags={data.tags} />
		</div>
		<PostList posts={rest} headingLevel={3} />
	</section>
{/if}

<style>
	.intro {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 400px);
		align-items: end;
		gap: 24px 64px;
		padding-block: 72px 48px;
	}

	h1 {
		max-width: 9ch;
		margin: 0;
		/* Un escalón por debajo del destacado: el titular que se lleva la
		   página es el del artículo, no el del blog. */
		font-size: clamp(40px, 4.8vw, 64px);
		font-stretch: 125%;
		font-weight: 650;
		line-height: 0.94;
		letter-spacing: -0.035em;
	}

	.lead {
		margin: 0;
		font-family: var(--blog-serif);
		font-size: 19px;
		font-weight: 380;
		line-height: 1.5;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	.rss {
		display: inline-block;
		margin-top: 12px;
		font-size: var(--lp-fs-sm);
	}

	.archive {
		padding-top: 72px;
	}

	.archive-head {
		display: grid;
		grid-template-columns: var(--blog-margin) minmax(0, 1fr);
		column-gap: var(--blog-gap);
		align-items: baseline;
		margin-bottom: 20px;
	}

	h2 {
		margin: 0;
		font-size: 21px;
		font-stretch: 114%;
		font-weight: 620;
		letter-spacing: -0.01em;
	}

	@media (max-width: 860px) {
		.intro {
			grid-template-columns: minmax(0, 1fr);
			padding-block: 44px 36px;
		}
	}

	@media (max-width: 760px) {
		.archive {
			padding-top: 48px;
		}
		.archive-head {
			grid-template-columns: minmax(0, 1fr);
			row-gap: 10px;
		}
		.lead {
			font-size: 17px;
		}
	}
</style>
