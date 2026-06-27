<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { features } from '$/config/features';

	interface Props {
		sidebarOpen?: boolean;
		user?: { role: string } | null;
	}

	let { sidebarOpen = false, user }: Props = $props();

	const menuItems = [
		{ label: 'Dashboard', icon: 'dashboard', href: resolve('/dashboard') },
		{ label: 'Portafolios', icon: 'briefcase', href: resolve('/dashboard/portfolios') },
		// "Inversiones" is hidden until the investments feature flag is enabled.
		...(features.investments
			? [{ label: 'Inversiones', icon: 'trending-up', href: resolve('/dashboard/investments') }]
			: []),
		{ label: 'Plataformas', icon: 'layers', href: resolve('/dashboard/platforms') },
		{ label: 'Transacciones', icon: 'exchange', href: resolve('/dashboard/transactions') },
		{ label: 'Reportes', icon: 'bar-chart', href: resolve('/dashboard/reports') },
		{ label: 'Notificaciones', icon: 'bell', href: resolve('/dashboard/notifications') },
		{ label: 'Configuración', icon: 'settings', href: resolve('/dashboard/settings') }
	];

	const adminItems = $derived(
		user?.role === 'admin'
			? [
					{ label: 'Panel Admin', icon: 'shield', href: resolve('/dashboard/admin') },
					{ label: 'Usuarios', icon: 'users', href: resolve('/dashboard/admin/users') },
					{ label: 'Activos', icon: 'database', href: resolve('/dashboard/admin/assets') }
				]
			: []
	);

	function isActive(href: string): boolean {
		if (href === resolve('/dashboard')) return page.url.pathname === href;
		return page.url.pathname.startsWith(href);
	}
</script>

<aside id="dashboard-sidebar" class="sidebar" class:open={sidebarOpen}>
	<nav class="sidebar-nav" aria-label="Navegación principal del dashboard">
		<h3 class="nav-title">Menú Principal</h3>
		<ul class="nav-list">
			{#each menuItems as item (item.href)}
				<li>
					<a href={item.href} class="nav-link" class:active={isActive(item.href)}>
						<span class="nav-icon">
							{#if item.icon === 'dashboard'}
								<svg
									width="18"
									height="18"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
								>
									<rect x="3" y="3" width="7" height="7"></rect>
									<rect x="14" y="3" width="7" height="7"></rect>
									<rect x="14" y="14" width="7" height="7"></rect>
									<rect x="3" y="14" width="7" height="7"></rect>
								</svg>
							{:else if item.icon === 'briefcase'}
								<svg
									width="18"
									height="18"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
								>
									<rect x="2" y="7" width="20" height="14" rx="2" ry="2"></rect>
									<path d="M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2"></path>
								</svg>
							{:else if item.icon === 'trending-up'}
								<svg
									width="18"
									height="18"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
								>
									<polyline points="23 6 13.5 15.5 8.5 10.5 1 17"></polyline>
									<polyline points="17 6 23 6 23 12"></polyline>
								</svg>
							{:else if item.icon === 'layers'}
								<svg
									width="18"
									height="18"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
								>
									<polygon points="12 2 20 6 20 18 12 22 4 18 4 6 12 2"></polygon>
									<polyline points="12 22 12 12"></polyline>
									<polyline points="4 6 12 12 20 6"></polyline>
									<polyline points="4 18 12 12 20 18"></polyline>
								</svg>
							{:else if item.icon === 'exchange'}
								<svg
									width="18"
									height="18"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
								>
									<line x1="12" y1="5" x2="12" y2="19"></line>
									<polyline points="19 12 12 19 5 12"></polyline>
								</svg>
							{:else if item.icon === 'bar-chart'}
								<svg
									width="18"
									height="18"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
								>
									<line x1="12" y1="20" x2="12" y2="10"></line>
									<line x1="18" y1="20" x2="18" y2="4"></line>
									<line x1="6" y1="20" x2="6" y2="16"></line>
								</svg>
							{:else if item.icon === 'bell'}
								<svg
									width="18"
									height="18"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
								>
									<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
									<path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
								</svg>
							{:else}
								<svg
									width="18"
									height="18"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
								>
									<circle cx="12" cy="12" r="3"></circle>
									<path
										d="M12 1v6m0 6v6M4.22 4.22l4.24 4.24m5.08 5.08l4.24 4.24M1 12h6m6 0h6m-1.78 7.78l-4.24-4.24m-5.08-5.08l-4.24-4.24"
									></path>
								</svg>
							{/if}
						</span>
						<span class="nav-label">{item.label}</span>
					</a>
				</li>
			{/each}
		</ul>
	</nav>

	{#if adminItems.length > 0}
		<nav class="sidebar-nav admin-nav" aria-label="Administración">
			<h3 class="nav-title">Administración</h3>
			<ul class="nav-list">
				{#each adminItems as item (item.href)}
					<li>
						<a href={item.href} class="nav-link" class:active={isActive(item.href)}>
							<span class="nav-icon">
								{#if item.icon === 'shield'}
									<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
									</svg>
								{:else if item.icon === 'users'}
									<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
										<circle cx="9" cy="7" r="4"></circle>
										<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
										<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
									</svg>
								{:else}
									<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
										<path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path>
										<path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
									</svg>
								{/if}
							</span>
							<span class="nav-label">{item.label}</span>
						</a>
					</li>
				{/each}
			</ul>
		</nav>
	{/if}

	<div class="sidebar-footer">
		<form action="?/logout" method="POST" use:enhance>
			<button class="sidebar-button secondary">
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
					<polyline points="16 17 21 12 16 7"></polyline>
					<line x1="21" y1="12" x2="9" y2="12"></line>
				</svg>
				Cerrar Sesión
			</button>
		</form>
		<p class="version">v1.0.0</p>
	</div>
</aside>

<style>
	.sidebar {
		position: fixed;
		left: 0;
		top: 72px;
		width: 280px;
		height: calc(100dvh - 72px);
		background: rgba(8, 9, 10, 0.72);
		border-right: 1px solid var(--border);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
		overflow-y: auto;
		overflow-x: hidden;
		display: flex;
		flex-direction: column;
		padding: 1.5rem 0;
		transition: transform 0.3s ease;
		z-index: 30;
	}

	.sidebar-nav {
		flex: 1;
		padding: 0 1rem;
	}

	.nav-title {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.2em;
		color: var(--text-dim);
		margin: 0 0 1.25rem 1rem;
		padding: 0;
	}

	.nav-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.nav-link {
		position: relative;
		display: flex;
		align-items: center;
		gap: 0.85rem;
		padding: 0.75rem 0.875rem;
		color: var(--text-muted);
		text-decoration: none;
		border-radius: 8px;
		border: 1px solid transparent;
		transition:
			color 0.2s ease,
			background 0.2s ease,
			border-color 0.2s ease;
		font-size: 0.9rem;
		font-weight: 400;
	}

	.nav-link:hover {
		background: var(--surface);
		color: var(--text);
	}

	.nav-link.active {
		background: rgba(212, 145, 42, 0.08);
		color: var(--amber-light);
		border-color: rgba(212, 145, 42, 0.22);
	}

	/* Amber rail marking the active item */
	.nav-link.active::before {
		content: '';
		position: absolute;
		left: -1px;
		top: 50%;
		transform: translateY(-50%);
		width: 2px;
		height: 1.1rem;
		border-radius: 2px;
		background: var(--amber);
	}

	.nav-icon {
		width: 20px;
		height: 20px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.nav-label {
		flex: 1;
	}

	.admin-nav {
		border-top: 1px solid var(--border);
		padding-top: 1.25rem;
		margin-top: 0.5rem;
	}

	.sidebar-footer {
		border-top: 1px solid var(--border);
		padding: 1.5rem 1rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.sidebar-button {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.65rem;
		padding: 0.75rem 1rem;
		background: var(--surface-2);
		border: 1px solid var(--border-strong);
		color: var(--text);
		border-radius: 6px;
		cursor: pointer;
		font-weight: 500;
		font-size: 0.85rem;
		transition:
			border-color 0.2s ease,
			background 0.2s ease,
			color 0.2s ease;
		font-family: var(--font-body);
	}

	.sidebar-button:hover {
		background: rgba(212, 145, 42, 0.06);
		color: var(--amber-light);
		border-color: rgba(212, 145, 42, 0.4);
	}

	.sidebar-button.secondary {
		width: 100%;
	}

	.version {
		text-align: center;
		font-family: var(--font-mono);
		font-size: 0.65rem;
		letter-spacing: 0.1em;
		color: var(--text-dim);
		margin: 0;
	}

	/* Scrollbar Styling */
	.sidebar::-webkit-scrollbar {
		width: 6px;
	}

	.sidebar::-webkit-scrollbar-track {
		background: transparent;
	}

	.sidebar::-webkit-scrollbar-thumb {
		background: var(--border-strong);
		border-radius: 3px;
	}

	.sidebar::-webkit-scrollbar-thumb:hover {
		background: rgba(212, 145, 42, 0.4);
	}

	@media (max-width: 1024px) {
		.sidebar {
			transform: translateX(-100%);
		}

		.sidebar.open {
			transform: translateX(0);
		}
	}

	@media (max-width: 768px) {
		.sidebar {
			top: 64px;
			height: calc(100dvh - 64px);
		}
	}
</style>
