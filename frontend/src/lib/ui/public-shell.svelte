<script lang="ts">
	/*
	 * El armazón de las pantallas públicas que no son la portada: las de acceso
	 * (entrar, crear cuenta, 2FA, recuperar y restablecer la contraseña,
	 * verificar el correo, activar una invitación), la autorización OAuth y la
	 * página de error.
	 *
	 * Hablan el idioma de la portada —papel, tinta, Archivo— y reparten la
	 * pantalla como su hero: a la izquierda, en grande, dónde estás; a la
	 * derecha, lo que hay que hacer.
	 *
	 * Quien lo usa pinta su titular con `.shell-title` y `.shell-note`, y la
	 * columna de la derecha con `.shell-copy` y `.shell-after` cuando no es un
	 * formulario: las reglas viven aquí y los alcanzan con `:global`.
	 */
	import '$lib/ui/public.css';
	import type { Snippet } from 'svelte';
	import { resolve } from '$app/paths';
	import Brand from '$lib/ui/brand.svelte';

	interface Props {
		/** La columna de la izquierda: el titular de la pantalla. */
		heading: Snippet;
		/** La de la derecha: el formulario o el mensaje. */
		children: Snippet;
	}

	let { heading, children }: Props = $props();

	const year = new Date().getFullYear();
</script>

<svelte:head>
	<meta name="theme-color" content="#eef0ec" />
</svelte:head>

<div class="lp shell">
	<header class="lp-wrap top">
		<a href={resolve('/')} class="home" aria-label="Finexia, volver al inicio">
			<Brand />
		</a>
		<a href={resolve('/')} class="back">Volver al inicio</a>
	</header>

	<main class="lp-wrap body">
		<div class="side">
			{@render heading()}
		</div>
		<div class="action">
			{@render children()}
		</div>
	</main>

	<footer class="lp-wrap bottom">
		<div class="foot">
			<span>© {year} Finexia</span>
			<nav aria-label="Legal">
				<a href={resolve('/privacidad')}>Privacidad</a>
				<a href={resolve('/terminos')}>Términos</a>
				<a href={resolve('/cookies')}>Cookies</a>
			</nav>
		</div>
	</footer>
</div>

<style>
	.shell {
		display: flex;
		flex-direction: column;
		min-height: 100dvh;
	}

	.top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		width: 100%;
		height: 72px;
	}

	.home {
		display: inline-flex;
		color: var(--lp-ink);
	}

	.back {
		font-size: 15px;
		color: var(--lp-ink-2);
	}

	.back:hover {
		color: var(--lp-ink);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.body {
		display: grid;
		flex: 1;
		grid-template-columns: minmax(0, 1fr) minmax(0, 440px);
		gap: 40px 96px;
		align-content: start;
		width: 100%;
		padding-block: clamp(40px, 11vh, 128px) 80px;
	}

	.action {
		min-width: 0;
	}

	/*
	 * El titular de cada pantalla. Lo pintan los componentes que usan el
	 * armazón, así que se alcanza desde aquí con `:global`: el tamaño de la
	 * portada, un escalón por debajo del hero.
	 */
	.side :global(.shell-title) {
		max-width: 12ch;
		margin: 0;
		font-size: clamp(40px, 5.4vw, 72px);
		font-stretch: 118%;
		font-weight: 640;
		line-height: 0.98;
		letter-spacing: -0.03em;
		text-wrap: balance;
		overflow-wrap: anywhere;
	}

	.side :global(.shell-note) {
		max-width: 36ch;
		margin: 24px 0 0;
		font-size: var(--lp-fs-lead);
		line-height: 1.5;
		color: var(--lp-ink-2);
		text-wrap: pretty;
	}

	/* Lo que va a la derecha cuando no es un formulario: un mensaje y su salida. */
	.action :global(.shell-copy) {
		margin: 0;
		font-size: var(--lp-fs-lead);
		line-height: 1.55;
		text-wrap: pretty;
	}

	.action :global(.shell-copy + .shell-copy) {
		margin-top: 12px;
	}

	.action :global(.shell-after) {
		margin: 28px 0 0;
		font-size: 15px;
		color: var(--lp-ink-2);
	}

	.bottom {
		width: 100%;
	}

	.foot {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px 24px;
		flex-wrap: wrap;
		padding-block: 20px 28px;
		border-top: 1px solid var(--lp-rule);
		font-size: var(--lp-fs-sm);
		color: var(--lp-ink-2);
	}

	.foot nav {
		display: flex;
		gap: 20px;
	}

	.foot a {
		color: var(--lp-ink-2);
	}

	.foot a:hover {
		color: var(--lp-ink);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	@media (max-width: 860px) {
		.body {
			grid-template-columns: minmax(0, 1fr);
			gap: 32px;
			padding-block: 28px 56px;
		}
		.side :global(.shell-title) {
			max-width: 14ch;
			font-size: clamp(36px, 9vw, 52px);
		}
		.side :global(.shell-note) {
			margin-top: 16px;
			font-size: 16px;
		}
	}

	@media (max-width: 640px) {
		.top {
			height: 60px;
		}
		.back {
			font-size: var(--lp-fs-sm);
		}
	}
</style>
