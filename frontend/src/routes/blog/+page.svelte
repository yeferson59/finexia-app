<script lang="ts">
	import { PostList, TagNav } from '$lib/features/blog';
	import { BLOG_DESCRIPTION, BLOG_TITLE, LOCALE, OG_IMAGE, SITE_NAME, absoluteUrl } from '$lib/seo';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const canonical = absoluteUrl('/blog');
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

<div class="intro">
	<h1>Blog</h1>
	<p class="lead">{BLOG_DESCRIPTION}</p>
</div>

<TagNav tags={data.tags} />

<PostList posts={data.posts} />

<style>
	.intro {
		max-width: 780px;
		margin-bottom: 36px;
	}

	h1 {
		margin: 0;
		font-size: clamp(40px, 5.4vw, 68px);
		font-stretch: 118%;
		font-weight: 640;
		line-height: 1;
		letter-spacing: -0.03em;
	}

	.lead {
		max-width: 52ch;
		margin: 18px 0 0;
		font-size: var(--lp-fs-lead);
		line-height: 1.55;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	@media (max-width: 640px) {
		.intro {
			margin-bottom: 26px;
		}
		.lead {
			margin-top: 12px;
			font-size: 16px;
		}
	}
</style>
