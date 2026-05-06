<script lang="ts">
	import { page } from '$app/state';

	interface Props {
		sidebarOpen?: boolean;
	}

	let { sidebarOpen = false }: Props = $props();

	const menuItems = [
		{ label: 'Dashboard', icon: 'dashboard', href: '/dashboard' },
		{ label: 'Portafolio', icon: 'briefcase', href: '/dashboard/portfolio' },
		{ label: 'Inversiones', icon: 'trending-up', href: '/dashboard/investments' },
		{ label: 'Transacciones', icon: 'exchange', href: '/dashboard/transactions' },
		{ label: 'Reportes', icon: 'bar-chart', href: '/dashboard/reports' },
		{ label: 'Configuración', icon: 'settings', href: '/dashboard/settings' }
	];

	function isActive(href: string): boolean {
		return href === '/dashboard' ? page.url.pathname === href : page.url.pathname.startsWith(href);
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
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<rect x="3" y="3" width="7" height="7"></rect>
									<rect x="14" y="3" width="7" height="7"></rect>
									<rect x="14" y="14" width="7" height="7"></rect>
									<rect x="3" y="14" width="7" height="7"></rect>
								</svg>
							{:else if item.icon === 'briefcase'}
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<rect x="2" y="7" width="20" height="14" rx="2" ry="2"></rect>
									<path d="M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2"></path>
								</svg>
							{:else if item.icon === 'trending-up'}
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polyline points="23 6 13.5 15.5 8.5 10.5 1 17"></polyline>
									<polyline points="17 6 23 6 23 12"></polyline>
								</svg>
							{:else if item.icon === 'exchange'}
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<line x1="12" y1="5" x2="12" y2="19"></line>
									<polyline points="19 12 12 19 5 12"></polyline>
								</svg>
							{:else if item.icon === 'bar-chart'}
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<line x1="12" y1="20" x2="12" y2="10"></line>
									<line x1="18" y1="20" x2="18" y2="4"></line>
									<line x1="6" y1="20" x2="6" y2="16"></line>
								</svg>
							{:else}
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="3"></circle>
									<path d="M12 1v6m0 6v6M4.22 4.22l4.24 4.24m5.08 5.08l4.24 4.24M1 12h6m6 0h6m-1.78 7.78l-4.24-4.24m-5.08-5.08l-4.24-4.24"></path>
								</svg>
							{/if}
						</span>
						<span class="nav-label">{item.label}</span>
					</a>
				</li>
			{/each}
		</ul>
	</nav>

	<div class="sidebar-footer">
		<button class="sidebar-button secondary">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
				<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
				<polyline points="16 17 21 12 16 7"></polyline>
				<line x1="21" y1="12" x2="9" y2="12"></line>
			</svg>
			Cerrar Sesión
		</button>
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
		background: rgba(26, 31, 46, 0.8);
		border-right: 1px solid rgba(212, 175, 55, 0.1);
		backdrop-filter: blur(12px);
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
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 1px;
		color: rgba(224, 224, 224, 0.4);
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
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 0.875rem 1rem;
		color: #e0e0e0;
		text-decoration: none;
		border-radius: 10px;
		transition: all 0.25s ease;
		font-size: 0.95rem;
		font-weight: 500;
		letter-spacing: 0.3px;
	}

	.nav-link:hover {
		background: rgba(212, 175, 55, 0.1);
		color: #d4af37;
	}

	.nav-link.active {
		background: linear-gradient(135deg, rgba(212, 175, 55, 0.2), rgba(46, 204, 113, 0.05));
		color: #d4af37;
		border: 1px solid rgba(212, 175, 55, 0.2);
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

	.sidebar-footer {
		border-top: 1px solid rgba(212, 175, 55, 0.1);
		padding: 1.5rem 1rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.sidebar-button {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		padding: 0.875rem 1rem;
		background: transparent;
		border: 1.5px solid rgba(212, 175, 55, 0.2);
		color: #e0e0e0;
		border-radius: 8px;
		cursor: pointer;
		font-weight: 600;
		font-size: 0.9rem;
		transition: all 0.25s ease;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.sidebar-button:hover {
		background: rgba(212, 175, 55, 0.1);
		color: #d4af37;
		border-color: rgba(212, 175, 55, 0.3);
	}

	.sidebar-button.secondary {
		width: 100%;
	}

	.version {
		text-align: center;
		font-size: 0.75rem;
		color: rgba(224, 224, 224, 0.3);
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
		background: rgba(212, 175, 55, 0.2);
		border-radius: 3px;
	}

	.sidebar::-webkit-scrollbar-thumb:hover {
		background: rgba(212, 175, 55, 0.3);
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
