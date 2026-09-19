<script lang="ts">
	/*
	 * El armazón del blog, hermano del de las páginas legales: la misma cabecera
	 * en papel con la marca a la izquierda, y el pie verde de la portada.
	 *
	 * La navegación lleva a la portada y al inicio de sesión, que es lo que
	 * puede querer quien llega aquí desde una búsqueda y no conoce el producto.
	 *
	 * Cada página pone su propio `.lp-wrap`: el índice abre con una banda a
	 * sangre que no cabe dentro del ancho del texto. Aquí se declaran los dos
	 * tokens que solo usa el blog: la serif del texto corrido y la columna del
	 * margen, que es donde van la fecha, la ficha y el índice del artículo.
	 */
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
	<header class="blog-header">
		<div class="lp-wrap nav">
			<a href={resolve('/')} class="brand-link" aria-label="Finexia, volver al inicio">
				<Brand />
			</a>
			<nav class="blog-nav" aria-label="Navegación del blog">
				<a href={resolve('/blog')}>Blog</a>
				<a href={resolve('/')}>El producto</a>
				<a href={resolve('/auth')}>Iniciar sesión</a>
			</nav>
		</div>
	</header>

	<main class="blog-main">
		{@render children()}
	</main>

	<Footer />
</div>

<style>
	.blog-header {
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
		gap: 16px;
		height: 72px;
	}

	.brand-link {
		display: inline-flex;
		color: var(--lp-ink);
	}

	.blog-nav {
		display: flex;
		align-items: center;
		gap: 28px;
	}

	.blog-nav a {
		font-size: 15px;
		color: var(--lp-ink-2);
		transition: color 0.15s;
	}

	.blog-nav a:hover {
		color: var(--lp-ink);
	}

	.blog-main {
		/* Fraunces con su eje óptico: a 19px dibuja los trazos para leer, no
		   para titular. */
		--blog-serif: 'Fraunces', Georgia, serif;
		--blog-margin: 220px;
		--blog-gap: 72px;
		--blog-header: 72px;

		min-height: 60vh;
		padding-bottom: 112px;
	}

	@media (max-width: 1100px) {
		.blog-main {
			--blog-margin: 180px;
			--blog-gap: 48px;
		}
	}

	@media (max-width: 860px) {
		.blog-nav {
			gap: 18px;
		}
	}

	@media (max-width: 480px) {
		.nav {
			height: 58px;
		}
		.blog-nav a {
			font-size: 13px;
		}
		.blog-main {
			--blog-header: 58px;
			padding-bottom: 72px;
		}
	}
</style>
