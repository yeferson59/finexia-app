<script lang="ts">
	import NetWorthCard from '$components/dashboard/net-worth-card.svelte';
	import PortfolioOverview from '$components/dashboard/portfolio-overview.svelte';
	import AssetAllocation from '$components/dashboard/asset-allocation.svelte';
	import RecentActivity from '$components/dashboard/recent-activity.svelte';

	import type { PageProps } from './$types';

	const { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Dashboard - FINEXIA</title>
	<meta name="description" content="Tu panel de control de inversiones y patrimonio" />
</svelte:head>

<header class="dashboard-header-section">
	<p class="dashboard-eyebrow">Resumen</p>
	<h1 id="dashboard-title" class="dashboard-title">
		Hola, <em>{data.user.name}</em>
	</h1>
	<p class="dashboard-subtitle">Aquí está tu patrimonio, de un vistazo.</p>
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
		margin-bottom: 2.5rem;
		padding-bottom: 1.75rem;
		border-bottom: 1px solid var(--border);
	}

	.dashboard-eyebrow {
		font-family: var(--font-mono);
		font-size: 0.6875rem;
		font-weight: 500;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--amber);
		margin: 0 0 0.75rem 0;
	}

	.dashboard-title {
		font-family: var(--font-display);
		font-size: clamp(2rem, 4vw, 2.75rem);
		font-weight: 300;
		line-height: 1.05;
		letter-spacing: -0.02em;
		color: var(--text);
		margin: 0 0 0.6rem 0;
	}

	.dashboard-title em {
		font-style: italic;
		font-weight: 500;
		color: var(--amber-light);
	}

	.dashboard-subtitle {
		font-size: 0.95rem;
		font-weight: 300;
		color: var(--text-muted);
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

	@media (prefers-reduced-motion: reduce) {
		.grid-item {
			animation: none;
		}
	}

	@media (max-width: 1024px) {
		.content-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
