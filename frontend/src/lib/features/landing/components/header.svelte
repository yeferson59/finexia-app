<script lang="ts">
	import { resolve } from '$app/paths';
	import Brand from './brand.svelte';

	let menuOpen = $state(false);
	let scrolled = $state(false);

	const login = resolve('/auth');

	// El `#` va literal en el marcado (`href="#{link.id}"`) y no dentro del dato:
	// así `svelte/no-navigation-without-resolve` reconoce el enlace de fragmento.
	const links = [
		{ id: 'producto', label: 'El producto' },
		{ id: 'como-funciona', label: 'Cómo funciona' },
		{ id: 'seguridad', label: 'Seguridad' },
		{ id: 'faq', label: 'Preguntas' }
	];
</script>

<svelte:window
	onscroll={() => (scrolled = window.scrollY > 8)}
	onkeydown={(event) => {
		if (event.key === 'Escape') menuOpen = false;
	}}
/>

<header class:scrolled>
	<a href="#contenido" class="skip">Saltar al contenido</a>
	<div class="lp-wrap bar">
		<a href={resolve('/')} class="home" aria-label="Finexia, inicio">
			<Brand />
		</a>

		<nav class="links" aria-label="Secciones de la página">
			{#each links as link (link.id)}
				<a href="#{link.id}">{link.label}</a>
			{/each}
		</nav>

		<div class="actions">
			<!--
				Dos acciones con pesos distintos: entrar es un enlace porque solo lo
				usa quien ya tiene cuenta; la lista de espera, que es a lo que viene la
				mayoría, se lleva el borde.
			-->
			<a href={login} class="login">Iniciar sesión</a>
			<a href="#waitlist" class="join">Unirme a la lista</a>
			<button
				class="burger"
				type="button"
				aria-label={menuOpen ? 'Cerrar menú' : 'Abrir menú'}
				aria-expanded={menuOpen}
				aria-controls="landing-menu"
				onclick={() => (menuOpen = !menuOpen)}
			>
				<span class="line" class:x={menuOpen}></span>
				<span class="line" class:y={menuOpen}></span>
			</button>
		</div>
	</div>

	<nav id="landing-menu" class="menu" class:open={menuOpen} aria-label="Menú de navegación">
		<div class="lp-wrap menu-inner">
			{#each links as link (link.id)}
				<a href="#{link.id}" onclick={() => (menuOpen = false)} tabindex={menuOpen ? 0 : -1}>
					{link.label}
				</a>
			{/each}
			<!-- En móvil la barra solo guarda «Iniciar sesión»: la lista de espera
			     baja aquí para que siga estando a un toque. -->
			<a
				href="#waitlist"
				class="menu-join"
				onclick={() => (menuOpen = false)}
				tabindex={menuOpen ? 0 : -1}
			>
				Unirme a la lista
			</a>
		</div>
	</nav>
</header>

<style>
	header {
		position: sticky;
		top: 0;
		z-index: 50;
		background: var(--lp-paper);
		border-bottom: 1px solid transparent;
		transition: border-color 0.2s ease;
	}

	header.scrolled {
		border-bottom-color: var(--lp-rule);
	}

	.skip {
		position: absolute;
		left: -9999px;
		top: 12px;
		z-index: 60;
		padding: 10px 16px;
		border-radius: 6px;
		background: var(--lp-ink);
		color: var(--lp-paper);
		font-size: var(--lp-fs-sm);
		font-weight: 600;
	}

	.skip:focus {
		left: 16px;
	}

	.bar {
		display: flex;
		align-items: center;
		gap: 40px;
		height: 72px;
	}

	.home {
		display: inline-flex;
		color: var(--lp-ink);
	}

	.links {
		display: flex;
		gap: 28px;
		margin-right: auto;
	}

	.links a,
	.login {
		font-size: 15px;
		color: var(--lp-ink-2);
		white-space: nowrap;
		transition: color 0.15s ease;
	}

	.links a:hover,
	.login:hover {
		color: var(--lp-ink);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 20px;
	}

	.join {
		display: inline-flex;
		align-items: center;
		min-height: 40px;
		padding: 0 16px;
		border: 1px solid var(--lp-ink);
		border-radius: 6px;
		font-size: 15px;
		font-weight: 600;
		color: var(--lp-ink);
		white-space: nowrap;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.join:hover {
		background: var(--lp-ink);
		color: var(--lp-paper);
	}

	.burger {
		display: none;
		flex-direction: column;
		justify-content: center;
		gap: 6px;
		width: 44px;
		height: 44px;
		padding: 0;
		border: none;
		background: transparent;
		cursor: pointer;
	}

	.line {
		display: block;
		width: 22px;
		height: 2px;
		margin: 0 auto;
		background: var(--lp-ink);
		transition: transform 0.22s ease;
	}

	.line.x {
		transform: translateY(4px) rotate(45deg);
	}

	.line.y {
		transform: translateY(-4px) rotate(-45deg);
	}

	.menu {
		display: none;
		overflow: hidden;
		max-height: 0;
		transition: max-height 0.28s ease;
	}

	.menu.open {
		max-height: 420px;
		border-bottom: 1px solid var(--lp-rule);
	}

	.menu-inner {
		display: flex;
		flex-direction: column;
		padding-bottom: 12px;
	}

	.menu a {
		padding: 14px 0;
		border-top: 1px solid var(--lp-rule);
		font-size: var(--lp-fs-lead);
		font-stretch: 108%;
		font-weight: 560;
		color: var(--lp-ink);
	}

	.menu .menu-join {
		text-decoration: underline;
		text-underline-offset: 4px;
	}

	@media (max-width: 960px) {
		.links {
			display: none;
		}
		.actions {
			margin-left: auto;
			gap: 8px;
		}
		.join {
			display: none;
		}
		.login {
			display: inline-flex;
			align-items: center;
			min-height: 40px;
			padding: 0 14px;
			border: 1px solid var(--lp-ink);
			border-radius: 6px;
			font-weight: 600;
			color: var(--lp-ink);
		}
		.burger {
			display: flex;
		}
		.menu {
			display: block;
		}
	}

	@media (max-width: 640px) {
		.bar {
			height: 60px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		header,
		.line,
		.menu {
			transition: none;
		}
	}
</style>
