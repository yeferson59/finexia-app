<script lang="ts">
	import DashboardHeader from '$components/dashboard/header.svelte';
	import Sidebar from '$components/dashboard/sidebar.svelte';

	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
	let sidebarOpen = $state(false);
</script>

<div class="dashboard-container">
	<div class="dashboard-scanlines" aria-hidden="true"></div>
	<DashboardHeader bind:sidebarOpen {data} />
	<Sidebar {sidebarOpen} user={data.user} />
	{#if sidebarOpen}
		<button
			class="sidebar-backdrop"
			onclick={() => (sidebarOpen = false)}
			aria-label="Cerrar menú lateral"
		></button>
	{/if}

	<main class="dashboard-main">
		<div class="dashboard-content">
			{@render children()}
		</div>
	</main>
</div>

<style>
	.dashboard-container {
		position: relative;
		display: flex;
		min-height: 100dvh;
		overflow-x: hidden;
		background:
			radial-gradient(ellipse 80% 50% at 70% -10%, rgba(212, 145, 42, 0.06), transparent 60%),
			var(--bg);
		color: var(--text);
	}

	/* Subtle film-grain scanline texture carried over from the landing page */
	.dashboard-scanlines {
		position: fixed;
		inset: 0;
		z-index: 0;
		pointer-events: none;
		background: repeating-linear-gradient(
			0deg,
			transparent,
			transparent 3px,
			rgba(255, 255, 255, 0.006) 3px,
			rgba(255, 255, 255, 0.006) 4px
		);
	}

	.dashboard-main {
		position: relative;
		z-index: 1;
		flex: 1;
		display: flex;
		flex-direction: column;
		min-width: 0;
		margin-top: 72px;
		margin-left: 280px;
		transition: margin-left 0.3s ease;
	}

	.dashboard-content {
		flex: 1;
		padding: 2rem;
		max-width: 1600px;
		width: 100%;
		margin: 0 auto;
	}

	.sidebar-backdrop {
		display: none;
	}

	@media (max-width: 1024px) {
		.dashboard-main {
			margin-left: 0;
		}

		.dashboard-content {
			padding: 1.5rem;
		}

		.sidebar-backdrop {
			display: block;
			position: fixed;
			inset: 72px 0 0;
			z-index: 20;
			border: none;
			background: rgba(8, 10, 14, 0.5);
			backdrop-filter: blur(2px);
		}
	}

	@media (max-width: 768px) {
		.dashboard-main {
			margin-top: 64px;
		}

		.dashboard-content {
			padding: 1rem;
		}

		.sidebar-backdrop {
			inset: 64px 0 0;
		}
	}
</style>
