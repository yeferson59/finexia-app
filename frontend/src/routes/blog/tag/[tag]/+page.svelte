<script lang="ts">
	import { PostList, TagNav, tagPath } from '$lib/features/blog';
	import { LOCALE, OG_IMAGE, SITE_NAME, absoluteUrl } from '$lib/seo';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const label = $derived(data.tag?.label ?? data.posts[0].tags[0]);
	const slug = $derived(data.tag?.slug ?? '');
	const canonical = $derived(absoluteUrl(tagPath(slug)));
	const title = $derived(`Artículos sobre ${label} — ${SITE_NAME}`);
	const description = $derived(
		`Todo lo que hemos escrito sobre ${label} en el blog de ${SITE_NAME}.`
	);
</script>

<svelte:head>
	<title>{title}</title>
	<meta name="description" content={description} />
	<link rel="canonical" href={canonical} />
	<meta name="robots" content="index,follow" />
	<meta property="og:type" content="website" />
	<meta property="og:site_name" content={SITE_NAME} />
	<meta property="og:locale" content={LOCALE} />
	<meta property="og:url" content={canonical} />
	<meta property="og:title" content={title} />
	<meta property="og:description" content={description} />
	<meta property="og:image" content={OG_IMAGE} />
	<meta name="twitter:card" content="summary_large_image" />
</svelte:head>

<div class="intro">
	<p class="eyebrow">Etiqueta</p>
	<h1>{label}</h1>
</div>

<TagNav tags={data.tags} active={slug} />

<PostList posts={data.posts} emptyTitle="Todavía no hay artículos con esta etiqueta." />

<style>
	.intro {
		max-width: 780px;
		margin-bottom: 36px;
	}

	.eyebrow {
		margin: 0 0 10px;
		font-size: var(--lp-fs-sm);
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--lp-ink-3);
	}

	h1 {
		margin: 0;
		font-size: clamp(40px, 5.4vw, 68px);
		font-stretch: 118%;
		font-weight: 640;
		line-height: 1;
		letter-spacing: -0.03em;
	}

	@media (max-width: 640px) {
		.intro {
			margin-bottom: 26px;
		}
	}
</style>
