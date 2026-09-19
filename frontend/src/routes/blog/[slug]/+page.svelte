<script lang="ts">
	import { PostBody, PostHeader, blogPostingJsonLd, postPath } from '$lib/features/blog';
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

<article>
	<PostHeader post={data.post} />
	<PostBody html={data.post.html} />
</article>
