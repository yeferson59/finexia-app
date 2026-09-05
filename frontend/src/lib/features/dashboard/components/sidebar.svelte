<script lang="ts">
	/*
	 * Menú lateral del panel.
	 *
	 * Ahora ocupa el alto entero de la ventana y lleva la marca arriba: antes
	 * empezaba 72 px por debajo, bajo una cabecera que cruzaba toda la pantalla,
	 * y esa junta partía la esquina superior izquierda en dos rectángulos que no
	 * se alineaban con nada.
	 *
	 * Los iconos y la lista de secciones salieron a `icons.ts` y `nav.ts`: el
	 * archivo eran 18 bloques de SVG en línea y la navegación se perdía entre
	 * ellos.
	 */
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import Icon from './icon.svelte';
	import { ADMIN_NAV, MAIN_NAV, isActive, type NavItem } from '../nav';

	interface Props {
		sidebarOpen?: boolean;
		user?: { role: string } | null;
	}

	let { sidebarOpen = false, user }: Props = $props();

	const adminItems = $derived(user?.role === 'admin' ? ADMIN_NAV : []);
</script>

<aside id="dashboard-sidebar" class="sidebar" class:open={sidebarOpen}>
	<a class="brand" href={resolve('/dashboard')}>
		<svg
			class="brand-mark"
			width="26"
			height="26"
			viewBox="0 0 30 30"
			fill="none"
			aria-hidden="true"
		>
			<rect width="30" height="30" rx="7" fill="var(--amber)" />
			<path
				d="M7 22L12.5 14.5L16.5 18.5L23 9"
				stroke="#0c0a06"
				stroke-width="2.6"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
		</svg>
		<span class="brand-name">Finexia</span>
	</a>

	<nav class="nav" aria-label="Secciones del panel">
		{@render list(MAIN_NAV)}
	</nav>

	{#if adminItems.length > 0}
		<nav class="nav admin" aria-label="Administración">
			<p class="group">Administración</p>
			{@render list(adminItems)}
		</nav>
	{/if}

	<div class="foot">
		<form action="/dashboard?/logout" method="POST" use:enhance>
			<button class="signout" type="submit">
				<Icon name="logout" size={16} />
				Cerrar sesión
			</button>
		</form>
		<p class="version">Finexia v1.0.0</p>
	</div>
</aside>

<!-- Los dos grupos se pintan igual; lo que cambia es el permiso que hace falta
     para verlos, no la forma de la entrada. -->
{#snippet list(items: NavItem[])}
	<ul>
		{#each items as item (item.href)}
			{@const active = isActive(item.href, page.url.pathname)}
			<li>
				<!-- `nav.ts` ya resolvió cada ruta al construir la lista; la regla no
				     puede seguir el rastro hasta allí. -->
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a href={item.href} class="item" class:active aria-current={active ? 'page' : undefined}>
					<Icon name={item.icon} />
					{item.label}
				</a>
			</li>
		{/each}
	</ul>
{/snippet}

<style>
	.sidebar {
		position: fixed;
		inset: 0 auto 0 0;
		z-index: 30;
		display: flex;
		flex-direction: column;
		width: var(--rail);
		padding: 1.25rem 0.75rem 1rem;
		background: var(--bg);
		border-right: 1px solid var(--border);
		overflow-y: auto;
		overscroll-behavior: contain;
		transition: transform 0.25s ease;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.35rem 0.6rem 1.5rem;
		text-decoration: none;
	}

	/* La única aparición de la serif en el panel: es el logotipo, no un titular. */
	.brand-name {
		font-family: var(--font-display);
		font-size: 1.15rem;
		font-weight: 600;
		letter-spacing: 0.02em;
		color: var(--text);
	}

	.nav ul {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.item {
		position: relative;
		display: flex;
		align-items: center;
		gap: 0.7rem;
		padding: 0.55rem 0.6rem;
		border-radius: 7px;
		color: var(--text-muted);
		font-size: 0.875rem;
		text-decoration: none;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.item:hover {
		background: var(--panel);
		color: var(--text);
	}

	/*
	 * Un solo estado activo, no cuatro. Antes la sección abierta se marcaba a la
	 * vez con fondo ámbar, texto ámbar, borde ámbar y una barra ámbar: cuatro
	 * señales para un hecho, y el ámbar dejaba de significar nada por repetirse.
	 */
	.item.active {
		background: var(--panel);
		color: var(--text);
		font-weight: 500;
	}

	.item.active::before {
		content: '';
		position: absolute;
		left: 0;
		top: 0.55rem;
		bottom: 0.55rem;
		width: 2px;
		border-radius: 0 2px 2px 0;
		background: var(--amber);
	}

	.admin {
		margin-top: 1.25rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
	}

	/* Separa dos permisos distintos, así que la etiqueta dice algo. */
	.group {
		margin: 0 0 0.5rem 0.6rem;
		font-size: 0.75rem;
		color: var(--text-dim);
	}

	.foot {
		margin-top: auto;
		padding-top: 1.25rem;
	}

	.signout {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		width: 100%;
		padding: 0.55rem 0.6rem;
		border: none;
		border-radius: 7px;
		background: transparent;
		color: var(--text-muted);
		font-family: inherit;
		font-size: 0.875rem;
		text-align: left;
		cursor: pointer;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.signout:hover {
		background: var(--panel);
		color: var(--text);
	}

	.version {
		margin: 0.75rem 0 0 0.6rem;
		font-size: 0.7rem;
		color: var(--text-dim);
	}

	@media (max-width: 1024px) {
		.sidebar {
			transform: translateX(-100%);
			box-shadow: 24px 0 48px rgba(0, 0, 0, 0.45);
		}

		.sidebar.open {
			transform: translateX(0);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.sidebar {
			transition: none;
		}
	}
</style>
