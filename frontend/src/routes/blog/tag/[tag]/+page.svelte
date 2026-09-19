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

<div class="lp-wrap">
	<div class="intro">
		<h1>Artículos sobre {label}</h1>
		<TagNav tags={data.tags} active={slug} />
	</div>

	<PostList posts={data.posts} emptyTitle="Todavía no hay artículos con esta etiqueta." />
</div>

<style>
	.intro {
		display: flex;
		flex-direction: column;
		gap: 28px;
		padding-block: 72px 36px;
	}

	h1 {
		max-width: 16ch;
		margin: 0;
		font-size: clamp(40px, 5.6vw, 76px);
		font-stretch: 125%;
		font-weight: 650;
		line-height: 0.96;
		letter-spacing: -0.035em;
		text-wrap: balance;
	}

	@media (max-width: 640px) {
		.intro {
			gap: 20px;
			padding-block: 44px 24px;
		}
	}
</style>
