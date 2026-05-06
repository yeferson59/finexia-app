<script lang="ts">
	import DashboardHeader from '$components/dashboard/header.svelte';
	import Sidebar from '$components/dashboard/sidebar.svelte';

	let { children } = $props();
	let sidebarOpen = $state(false);
</script>

<div class="dashboard-container">
	<DashboardHeader bind:sidebarOpen />
	<Sidebar {sidebarOpen} />
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
		display: flex;
		min-height: 100dvh;
		background: linear-gradient(135deg, #0f1419 0%, #16191f 50%, #0f1419 100%);
		color: #e0e0e0;
	}

	.dashboard-main {
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
