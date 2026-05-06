<script lang="ts">
interface Props {
	sidebarOpen?: boolean;
}

let { sidebarOpen = $bindable(false) }: Props = $props();
</script>

<header class="dashboard-header">
	<div class="header-content">
		<button
			class="menu-toggle"
			onclick={() => (sidebarOpen = !sidebarOpen)}
			aria-label="Toggle menu"
			aria-expanded={sidebarOpen}
			aria-controls="dashboard-sidebar"
		>
			<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<line x1="3" y1="12" x2="21" y2="12"></line>
				<line x1="3" y1="6" x2="21" y2="6"></line>
				<line x1="3" y1="18" x2="21" y2="18"></line>
			</svg>
		</button>

		<div class="header-brand">
			<div class="logo-mark" aria-hidden="true"></div>
			<span class="brand-name">FINEXIA</span>
		</div>

		<div class="header-actions">
			<button class="icon-button" aria-label="Notificaciones">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
					<path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
				</svg>
				<span class="notification-badge"></span>
			</button>

			<button class="icon-button" aria-label="Configuración">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="1"></circle>
					<circle cx="19" cy="12" r="1"></circle>
					<circle cx="5" cy="12" r="1"></circle>
				</svg>
			</button>

			<div class="user-profile">
				<div class="avatar"></div>
				<div class="user-info">
					<p class="user-name">Juan Pérez</p>
					<p class="user-email">juan@email.com</p>
				</div>
			</div>
		</div>
	</div>
</header>

<style>
	.dashboard-header {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 72px;
		background: rgba(15, 20, 25, 0.95);
		backdrop-filter: blur(12px);
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
		z-index: 40;
	}

	.header-content {
		max-width: 1600px;
		margin: 0 auto;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 2rem;
		gap: 2rem;
	}

	.menu-toggle {
		display: none;
		background: none;
		border: none;
		color: #e0e0e0;
		cursor: pointer;
		padding: 0.5rem;
		transition: color 0.25s ease;
	}

	.menu-toggle:hover {
		color: #d4af37;
	}

	.header-brand {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-shrink: 0;
	}

	.logo-mark {
		width: 36px;
		height: 36px;
		border-radius: 8px;
		background: linear-gradient(135deg, #d4af37 0%, #c49e26 100%);
		border: 1px solid rgba(232, 197, 71, 0.4);
		box-shadow: 0 0 12px rgba(212, 175, 55, 0.25);
	}

	.brand-name {
		font-size: 1.25rem;
		font-weight: 700;
		color: #d4af37;
		letter-spacing: 1px;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 1.5rem;
		margin-left: auto;
	}

	.icon-button {
		background: none;
		border: none;
		color: #e0e0e0;
		cursor: pointer;
		padding: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: color 0.25s ease;
		position: relative;
	}

	.icon-button:hover {
		color: #d4af37;
	}

	.notification-badge {
		position: absolute;
		top: 2px;
		right: 2px;
		width: 8px;
		height: 8px;
		background: #e74c3c;
		border-radius: 50%;
	}

	.user-profile {
		display: flex;
		align-items: center;
		gap: 0.875rem;
		padding-left: 1rem;
		border-left: 1px solid rgba(212, 175, 55, 0.1);
	}

	.avatar {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: linear-gradient(135deg, #d4af37, #2ecc71);
		border: 2px solid rgba(212, 175, 55, 0.3);
		flex-shrink: 0;
	}

	.user-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.user-name {
		font-size: 0.85rem;
		font-weight: 600;
		color: #e0e0e0;
		margin: 0;
	}

	.user-email {
		font-size: 0.75rem;
		color: rgba(224, 224, 224, 0.5);
		margin: 0;
	}

	@media (max-width: 1024px) {
		.menu-toggle {
			display: flex;
		}

		.user-info {
			display: none;
		}

		.header-content {
			padding: 0 1.5rem;
			gap: 1rem;
		}
	}

	@media (max-width: 768px) {
		.dashboard-header {
			height: 64px;
		}

		.header-content {
			padding: 0 1rem;
		}

		.header-actions {
			gap: 1rem;
		}

		.brand-name {
			display: none;
		}
	}
</style>
