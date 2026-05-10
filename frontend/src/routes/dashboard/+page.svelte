<script lang="ts">
	import NetWorthCard from '$components/dashboard/net-worth-card.svelte';
	import PortfolioOverview from '$components/dashboard/portfolio-overview.svelte';
	import AssetAllocation from '$components/dashboard/asset-allocation.svelte';
	import RecentActivity from '$components/dashboard/recent-activity.svelte';

	const { data } = $props();
</script>

<svelte:head>
	<title>Dashboard - FINEXIA</title>
	<meta name="description" content="Tu panel de control de inversiones y patrimonio" />
</svelte:head>

<header class="dashboard-header-section">
	<h1 id="dashboard-title" class="dashboard-title">Dashboard</h1>
	<p class="dashboard-subtitle">Bienvenido de vuelta, {data.user.name}</p>
</header>

<section class="net-worth-section" aria-labelledby="dashboard-title">
	<NetWorthCard />
</section>

<section class="content-grid" aria-label="Resumen financiero">
	<div class="grid-item full-width">
		<PortfolioOverview />
	</div>

	<div class="grid-item">
		<AssetAllocation />
	</div>

	<div class="grid-item">
		<RecentActivity />
	</div>
</section>

<style>
	.dashboard-header-section {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
	}

	.dashboard-title {
		font-size: 2.5rem;
		font-weight: 700;
		color: #d4af37;
		margin: 0 0 0.5rem 0;
		letter-spacing: 0.5px;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.dashboard-subtitle {
		font-size: 1rem;
		color: rgba(224, 224, 224, 0.6);
		margin: 0;
	}

	.net-worth-section {
		margin-bottom: 2.5rem;
	}

	.content-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 2rem;
	}

	.grid-item {
		animation: fade-in 0.5s ease-out;
		animation-fill-mode: both;
	}

	.grid-item:nth-child(1) {
		animation-delay: 0.1s;
	}

	.grid-item:nth-child(2) {
		animation-delay: 0.2s;
	}

	.grid-item:nth-child(3) {
		animation-delay: 0.3s;
	}

	.grid-item.full-width {
		grid-column: 1 / -1;
	}

	@keyframes fade-in {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@media (max-width: 1024px) {
		.content-grid {
			grid-template-columns: 1fr;
		}

		.dashboard-title {
			font-size: 2rem;
		}
	}

	@media (max-width: 768px) {
		.dashboard-title {
			font-size: 1.75rem;
		}
	}
</style>
