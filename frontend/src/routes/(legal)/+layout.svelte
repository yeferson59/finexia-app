<script lang="ts">
	import { resolve } from '$app/paths';
	import '$lib/features/landing/landing.css';
	import { Brand, Footer } from '$lib/features/landing';

	let { children } = $props();
</script>

<header class="legal-header">
	<div class="wrap nav">
		<a href={resolve('/')} class="brand-link" aria-label="Volver al inicio">
			<Brand />
		</a>
		<nav class="legal-nav">
			<a href={resolve('/privacidad')}>Privacidad</a>
			<a href={resolve('/terminos')}>Términos</a>
			<a href={resolve('/cookies')}>Cookies</a>
		</nav>
	</div>
</header>

<main class="legal-main">
	<div class="wrap legal-prose">
		{@render children()}
	</div>
</main>

<Footer />

<style>
	.legal-header {
		position: sticky;
		top: 0;
		z-index: 50;
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
		background: rgba(8, 9, 10, 0.82);
		border-bottom: 1px solid var(--border);
	}
	.nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		height: 66px;
	}
	.brand-link {
		display: inline-flex;
	}
	.legal-nav {
		display: flex;
		align-items: center;
		gap: 28px;
	}
	.legal-nav a {
		font-size: 14px;
		color: var(--text-muted);
		transition: color 0.2s;
	}
	.legal-nav a:hover {
		color: var(--text);
	}

	.legal-main {
		min-height: 60vh;
		padding: 56px 0 40px;
	}
	.legal-prose {
		max-width: 760px;
	}

	/* Prose styling scoped to legal content rendered via {@render children()}. */
	.legal-prose :global(h1) {
		font-family: var(--font-display);
		font-weight: 400;
		font-size: clamp(30px, 5vw, 46px);
		line-height: 1.1;
		letter-spacing: -0.02em;
		margin-bottom: 10px;
	}
	.legal-prose :global(.legal-updated) {
		font-family: var(--font-mono);
		font-size: 12px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin-bottom: 36px;
	}
	.legal-prose :global(.legal-intro) {
		font-size: 17px;
		color: var(--text-muted);
		font-weight: 300;
		line-height: 1.7;
		margin-bottom: 40px;
	}
	.legal-prose :global(h2) {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: clamp(20px, 3vw, 26px);
		letter-spacing: -0.01em;
		margin: 40px 0 14px;
	}
	.legal-prose :global(h3) {
		font-weight: 600;
		font-size: 16px;
		margin: 24px 0 8px;
		color: var(--text);
	}
	.legal-prose :global(p) {
		font-size: 15.5px;
		line-height: 1.75;
		color: var(--text-muted);
		margin: 0 0 16px;
	}
	.legal-prose :global(ul) {
		margin: 0 0 20px;
		padding-left: 22px;
		list-style: disc;
	}
	.legal-prose :global(li) {
		font-size: 15.5px;
		line-height: 1.7;
		color: var(--text-muted);
		margin-bottom: 8px;
	}
	.legal-prose :global(a) {
		color: var(--amber-light);
		text-decoration: underline;
		text-underline-offset: 2px;
	}
	.legal-prose :global(a:hover) {
		color: var(--amber);
	}
	.legal-prose :global(strong) {
		color: var(--text);
		font-weight: 600;
	}
	.legal-prose :global(.legal-contact) {
		margin-top: 40px;
		padding: 24px 26px;
		border: 1px solid var(--border-strong);
		border-radius: 12px;
		background: var(--surface);
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
			padding: 36px 0 28px;
		}
	}
</style>
