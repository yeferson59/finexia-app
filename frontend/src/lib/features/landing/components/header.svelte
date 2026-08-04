<script lang="ts">
	import { onMount } from 'svelte';
	import Brand from './brand.svelte';

	let headerEl: HTMLElement;
	let menuOpen = $state(false);

	// El `#` va literal en el marcado (`href="#{link.id}"`) y no dentro del dato:
	// así `svelte/no-navigation-without-resolve` reconoce el enlace de fragmento.
	const links = [
		{ id: 'producto', label: 'El producto' },
		{ id: 'beneficios', label: 'Beneficios' },
		{ id: 'como-funciona', label: 'Cómo funciona' },
		{ id: 'seguridad', label: 'Seguridad' },
		{ id: 'faq', label: 'Preguntas' }
	];

	onMount(() => {
		let ticking = false;
		const onScroll = () => {
			if (!ticking) {
				ticking = true;
				requestAnimationFrame(() => {
					headerEl.classList.toggle('scrolled', window.scrollY > 10);
					ticking = false;
				});
			}
		};
		window.addEventListener('scroll', onScroll, { passive: true });
		return () =>
			window.removeEventListener('scroll', onScroll, { passive: true } as EventListenerOptions);
	});
</script>

<svelte:window
	onkeydown={(event) => {
		if (event.key === 'Escape') menuOpen = false;
	}}
/>

<header bind:this={headerEl}>
	<div class="wrap nav">
		<a href="#contenido" class="skip">Saltar al contenido</a>
		<Brand />
		<nav class="nav-links" aria-label="Secciones de la página">
			{#each links as link (link.id)}
				<a href="#{link.id}">{link.label}</a>
			{/each}
		</nav>
		<div class="nav-right">
			<a href="#waitlist" class="nav-cta">Unirme a la lista</a>
			<button
				class="burger"
				type="button"
				aria-label={menuOpen ? 'Cerrar menú' : 'Abrir menú'}
				aria-expanded={menuOpen}
				aria-controls="landing-menu"
				onclick={() => (menuOpen = !menuOpen)}
			>
				<span class="bar" class:x={menuOpen}></span>
				<span class="bar mid" class:hide={menuOpen}></span>
				<span class="bar" class:y={menuOpen}></span>
			</button>
		</div>
	</div>

	<nav id="landing-menu" class="menu" class:open={menuOpen} aria-label="Menú de navegación">
		{#each links as link (link.id)}
			<a href="#{link.id}" onclick={() => (menuOpen = false)} tabindex={menuOpen ? 0 : -1}>
				{link.label}
			</a>
		{/each}
	</nav>
</header>

<style>
	header:global(.scrolled) {
		box-shadow: 0 1px 0 rgba(255, 255, 255, 0.04);
	}
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
		gap: 16px;
		height: 66px;
	}

	/* Visible solo cuando recibe foco con el tabulador. */
	.skip {
		position: absolute;
		left: -9999px;
		top: 12px;
		padding: 10px 16px;
		border-radius: 6px;
		background: var(--amber);
		color: #0d0800;
		font-size: 13.5px;
		font-weight: 600;
		z-index: 60;
	}
	.skip:focus {
		left: 24px;
	}

	.nav-links {
		display: flex;
		align-items: center;
		gap: 28px;
	}
	.nav-links a {
		font-size: 14px;
		color: var(--text-muted);
		font-weight: 400;
		white-space: nowrap;
		transition: color 0.2s;
	}
	.nav-links a:hover {
		color: var(--text);
	}

	.nav-right {
		display: flex;
		align-items: center;
		gap: 10px;
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
		white-space: nowrap;
		transition:
			border-color 0.2s,
			background 0.2s;
	}
	.nav-cta:hover {
		border-color: var(--amber);
		background: rgba(212, 145, 42, 0.06);
	}

	.burger {
		display: none;
		flex-direction: column;
		justify-content: center;
		gap: 4px;
		width: 38px;
		height: 38px;
		padding: 0;
		border: 1px solid var(--border-strong);
		border-radius: 6px;
		background: transparent;
		cursor: pointer;
	}
	.bar {
		display: block;
		width: 15px;
		height: 1.5px;
		margin: 0 auto;
		border-radius: 1px;
		background: var(--text);
		transition:
			transform 0.22s ease,
			opacity 0.18s ease;
	}
	.bar.x {
		transform: translateY(5.5px) rotate(45deg);
	}
	.bar.y {
		transform: translateY(-5.5px) rotate(-45deg);
	}
	.bar.hide {
		opacity: 0;
	}

	.menu {
		display: none;
		flex-direction: column;
		border-top: 1px solid var(--border);
		overflow: hidden;
		max-height: 0;
		transition: max-height 0.28s ease;
	}
	.menu.open {
		max-height: 340px;
	}
	.menu a {
		padding: 14px 24px;
		border-bottom: 1px solid var(--border);
		font-size: 15px;
		color: var(--text-muted);
	}
	.menu a:last-child {
		border-bottom: none;
	}
	.menu a:hover {
		color: var(--text);
		background: var(--surface);
	}

	@media (max-width: 1000px) {
		.nav-links {
			gap: 20px;
		}
		.nav-links a {
			font-size: 13.5px;
		}
	}

	@media (max-width: 900px) {
		.nav-links {
			display: none;
		}
		.burger {
			display: flex;
		}
		.menu {
			display: flex;
		}
	}

	@media (max-width: 480px) {
		.nav {
			height: 58px;
		}
		.nav-cta {
			padding: 8px 14px;
			font-size: 13px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.bar,
		.menu {
			transition: none;
		}
	}
</style>
