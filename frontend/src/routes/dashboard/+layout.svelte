<script lang="ts">
	import { DashboardHeader, Sidebar } from '$lib/features/dashboard';

	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
	let sidebarOpen = $state(false);
</script>

<div class="shell">
	<Sidebar {sidebarOpen} user={data.user} />
	<DashboardHeader bind:sidebarOpen {data} />

	{#if sidebarOpen}
		<button class="scrim" onclick={() => (sidebarOpen = false)} aria-label="Cerrar el menú"
		></button>
	{/if}

	<main class="main">
		<div class="content">
			{@render children()}
		</div>
	</main>
</div>

<style>
	/*
	 * Medidas del chrome. Viven aquí y no en `layout.css` porque solo existen en
	 * el panel: la portada pública y el login comparten los colores y las
	 * tipografías, pero no tienen ni menú lateral ni barra superior.
	 */
	.shell {
		--rail: 232px;
		--header-h: 60px;
		/*
		 * Dos niveles de superficie, no cuatro. Uno para el fondo y otro para lo
		 * que de verdad es un objeto suelto —una fila de movimiento, la sección
		 * abierta del menú—; el panel tenía cuatro y cinco tarjetas idénticas
		 * flotando sobre él, así que ninguna decía nada por estar levantada.
		 */
		--panel: rgba(255, 255, 255, 0.04);
		--panel-2: rgba(255, 255, 255, 0.08);
		/*
		 * El capital invertido, frente al ámbar del valor de mercado: el dinero
		 * que pusiste se lee frío y lo que el mercado hizo con él, cálido.
		 * Validado contra el ámbar — ΔE 20,9 en visión normal y 18,4 en protanopia
		 * —, que es lo que hace que las dos líneas de la gráfica se distingan.
		 */
		--cost: #73819c;

		min-height: 100dvh;
		background: var(--bg);
		color: var(--text);
	}

	.main {
		margin-left: var(--rail);
		padding: calc(var(--header-h) + 2.25rem) 2.25rem 4rem;
	}

	/* Centrada en el hueco que deja el menú, no pegada a él: con el ancho tope
	   aplicado al propio `main` una pantalla de 1920 dejaba medio metro de vacío
	   solo a la derecha. */
	.content {
		max-width: 1180px;
		margin: 0 auto;
	}

	.scrim {
		display: none;
	}

	@media (max-width: 1024px) {
		.main {
			margin-left: 0;
			padding: calc(var(--header-h) + 1.5rem) 1.25rem 3rem;
		}

		.scrim {
			display: block;
			position: fixed;
			inset: 0;
			z-index: 25;
			border: none;
			background: rgba(4, 5, 6, 0.6);
		}
	}
</style>
