<script lang="ts">
	import { resolve } from '$app/paths';

	interface Props {
		sidebarOpen?: boolean;
		data: { user: { name: string; email: string; image?: string } };
	}

	let { sidebarOpen = $bindable(false), data }: Props = $props();
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
			<svg
				width="24"
				height="24"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<line x1="3" y1="12" x2="21" y2="12"></line>
				<line x1="3" y1="6" x2="21" y2="6"></line>
				<line x1="3" y1="18" x2="21" y2="18"></line>
			</svg>
		</button>

		<a class="header-brand" href={resolve('/dashboard')} aria-label="Finexia, ir al inicio">
			<svg
				class="brand-icon"
				width="30"
				height="30"
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
			<span class="brand-name">FINEXIA</span>
		</a>

		<div class="header-actions">
			<a href={resolve('/dashboard/notifications')} class="icon-button" aria-label="Notificaciones">
				<svg
					width="20"
					height="20"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
					<path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
				</svg>
				<span class="notification-badge"></span>
			</a>

			<button class="icon-button" aria-label="Configuración">
				<svg
					width="20"
					height="20"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<circle cx="12" cy="12" r="1"></circle>
					<circle cx="19" cy="12" r="1"></circle>
					<circle cx="5" cy="12" r="1"></circle>
				</svg>
			</button>

			<div class="user-profile">
				{#if data.user.image && data.user.image !== 'avatar.png'}
					<img src={data.user.image} alt="Avatar" class="avatar avatar-img" />
				{:else}
					<div class="avatar" aria-hidden="true">
						{data.user.name.trim().charAt(0).toUpperCase()}
					</div>
				{/if}
				<div class="user-info">
					<p class="user-name">{data.user.name}</p>
					<p class="user-email">{data.user.email}</p>
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
		background: rgba(8, 9, 10, 0.82);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
		border-bottom: 1px solid var(--border);
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
		color: var(--text-muted);
		cursor: pointer;
		padding: 0.5rem;
		transition: color 0.2s ease;
	}

	.menu-toggle:hover {
		color: var(--text);
	}

	.header-brand {
		display: flex;
		align-items: center;
		gap: 0.65rem;
		flex-shrink: 0;
	}

	.brand-icon {
		flex-shrink: 0;
	}

	.brand-name {
		font-family: var(--font-display);
		font-weight: 600;
		font-size: 1.25rem;
		letter-spacing: 0.1em;
		color: var(--text);
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
		color: var(--text-muted);
		cursor: pointer;
		padding: 0.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: color 0.2s ease;
		position: relative;
		text-decoration: none;
	}

	.icon-button:hover {
		color: var(--text);
	}

	.notification-badge {
		position: absolute;
		top: 4px;
		right: 4px;
		width: 6px;
		height: 6px;
		background: var(--amber);
		border-radius: 50%;
	}

	.user-profile {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding-left: 1rem;
		border-left: 1px solid var(--border);
	}

	.avatar {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		background: var(--surface-3);
		border: 1px solid var(--border-strong);
		color: var(--amber);
		font-family: var(--font-mono);
		font-size: 0.8rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.avatar-img {
		object-fit: cover;
	}

	.user-info {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
	}

	.user-name {
		font-size: 0.85rem;
		font-weight: 500;
		color: var(--text);
		margin: 0;
	}

	.user-email {
		font-family: var(--font-mono);
		font-size: 0.7rem;
		color: var(--text-dim);
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
