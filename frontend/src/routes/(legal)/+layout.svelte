<script lang="ts">
	import { resolve } from '$app/paths';
	import '$lib/ui/public.css';
	import Brand from '$lib/ui/brand.svelte';
	import { Footer } from '$lib/features/landing';

	let { children } = $props();
</script>

<svelte:head>
	<meta name="theme-color" content="#eef0ec" />
</svelte:head>

<div class="lp">
	<header class="legal-header">
		<div class="lp-wrap nav">
			<a href={resolve('/')} class="brand-link" aria-label="Volver al inicio">
				<Brand />
			</a>
			<nav class="legal-nav" aria-label="Páginas legales">
				<a href={resolve('/privacidad')}>Privacidad</a>
				<a href={resolve('/terminos')}>Términos</a>
				<a href={resolve('/cookies')}>Cookies</a>
			</nav>
		</div>
	</header>

	<main class="legal-main">
		<div class="lp-wrap">
			<div class="legal-prose">
				{@render children()}
			</div>
		</div>
	</main>

	<Footer />
</div>

<style>
	.legal-header {
		position: sticky;
		top: 0;
		z-index: 50;
		background: var(--lp-paper);
		border-bottom: 1px solid var(--lp-rule);
	}
	.nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 72px;
	}
	.brand-link {
		display: inline-flex;
		color: var(--lp-ink);
	}
	.legal-nav {
		display: flex;
		align-items: center;
		gap: 28px;
	}
	.legal-nav a {
		font-size: 15px;
		color: var(--lp-ink-2);
		transition: color 0.15s;
	}
	.legal-nav a:hover {
		color: var(--lp-ink);
	}

	.legal-main {
		min-height: 60vh;
		padding: 64px 0 96px;
	}
	.legal-prose {
		max-width: 720px;
	}

	/* Prose styling scoped to legal content rendered via {@render children()}. */
	.legal-prose :global(h1) {
		font-size: clamp(36px, 5.4vw, 60px);
		font-stretch: 116%;
		font-weight: 620;
		line-height: 1;
		letter-spacing: -0.025em;
		margin: 0 0 16px;
	}
	.legal-prose :global(.legal-updated) {
		font-size: 14px;
		color: var(--lp-ink-2);
		margin: 0 0 36px;
	}
	.legal-prose :global(.legal-intro) {
		font-size: 18px;
		color: var(--lp-ink);
		line-height: 1.6;
		margin: 0 0 40px;
	}
	.legal-prose :global(h2) {
		font-size: 24px;
		font-stretch: 110%;
		font-weight: 620;
		letter-spacing: -0.01em;
		line-height: 1.2;
		margin: 48px 0 14px;
	}
	.legal-prose :global(h3) {
		font-weight: 600;
		font-size: 17px;
		margin: 28px 0 8px;
		color: var(--lp-ink);
	}
	.legal-prose :global(p) {
		font-size: 16px;
		line-height: 1.7;
		color: var(--lp-ink-2);
		margin: 0 0 16px;
	}
	.legal-prose :global(ul) {
		margin: 0 0 20px;
		padding-left: 22px;
		list-style: disc;
	}
	.legal-prose :global(li) {
		font-size: 16px;
		line-height: 1.7;
		color: var(--lp-ink-2);
		margin-bottom: 8px;
	}
	.legal-prose :global(a) {
		color: var(--lp-ink);
		text-decoration: underline;
		text-underline-offset: 3px;
	}
	.legal-prose :global(a:hover) {
		text-decoration-thickness: 2px;
	}
	.legal-prose :global(strong) {
		color: var(--lp-ink);
		font-weight: 600;
	}
	.legal-prose :global(.legal-contact) {
		margin-top: 48px;
		padding: 24px 28px;
		border-radius: 12px;
		background: var(--lp-paper-2);
	}
	.legal-prose :global(.legal-contact p) {
		margin-bottom: 6px;
	}

	@media (max-width: 860px) {
		.legal-nav {
			gap: 18px;
		}
	}
	@media (max-width: 480px) {
		.nav {
			height: 58px;
		}
		.legal-nav a {
			font-size: 13px;
		}
		.legal-main {
			padding: 40px 0 64px;
		}
	}
</style>
