<script lang="ts">
	import { onMount } from 'svelte';
	import Brand from './brand.svelte';

	let headerEl: HTMLElement;

	onMount(() => {
		const onScroll = () => {
			headerEl.style.boxShadow = window.scrollY > 10 ? '0 1px 0 rgba(255,255,255,0.04)' : 'none';
		};
		window.addEventListener('scroll', onScroll);
		return () => window.removeEventListener('scroll', onScroll);
	});
</script>

<header bind:this={headerEl}>
	<div class="wrap nav">
		<Brand />
		<nav class="nav-links">
			<a href="#beneficios">Beneficios</a>
			<a href="#como-funciona">Cómo funciona</a>
			<a href="#faq">Preguntas</a>
		</nav>
		<a href="#waitlist" class="nav-cta">Unirme a la lista</a>
	</div>
</header>

<style>
	header {
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
	.nav-links {
		display: flex;
		align-items: center;
		gap: 36px;
	}
	.nav-links a {
		font-size: 14px;
		color: var(--text-muted);
		font-weight: 400;
		transition: color 0.2s;
	}
	.nav-links a:hover {
		color: var(--text);
	}
	.nav-cta {
		display: inline-flex;
		align-items: center;
		padding: 9px 18px;
		border-radius: 6px;
		border: 1px solid var(--border-strong);
		font-size: 13.5px;
		font-weight: 500;
		color: var(--text);
		transition:
			border-color 0.2s,
			background 0.2s;
	}
	.nav-cta:hover {
		border-color: var(--amber);
		background: rgba(212, 145, 42, 0.06);
	}
	@media (max-width: 860px) {
		.nav-links {
			display: none;
		}
	}
</style>
