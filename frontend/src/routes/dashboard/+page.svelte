<script lang="ts">
	import NetWorthCard from '$components/dashboard/net-worth-card.svelte';
	import PortfolioOverview from '$components/dashboard/portfolio-overview.svelte';
	import AssetAllocation from '$components/dashboard/asset-allocation.svelte';
	import RecentActivity from '$components/dashboard/recent-activity.svelte';
	import PortfolioGrowth from '$components/dashboard/portfolio-growth.svelte';
	import PageHeader from '$lib/ui/page-header.svelte';

	import type { PageProps } from './$types';

	const { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Dashboard - FINEXIA</title>
	<meta name="description" content="Tu panel de control de inversiones y patrimonio" />
</svelte:head>

<PageHeader
	eyebrow="Resumen"
	titleId="dashboard-title"
	subtitle="Aquí está tu patrimonio, de un vistazo."
>
	Hola, <em>{data.user.name}</em>
</PageHeader>

<section class="net-worth-section" aria-labelledby="dashboard-title">
	<NetWorthCard summaries={data.portfolioSummaries} currency={data.currency} />
</section>

<section class="growth-section" aria-label="Crecimiento del portafolio">
	<PortfolioGrowth data={data.portfolioGrowth.points} summary={data.portfolioGrowth.summary} />
</section>

<section class="content-grid" aria-label="Resumen financiero">
	<div class="grid-item full-width">
		<PortfolioOverview summaries={data.portfolioSummaries} currency={data.currency} />
	</div>

	<div class="grid-item">
		<AssetAllocation allocation={data.allocation} />
	</div>

	<div class="grid-item">
		<RecentActivity transactions={data.recentTransactions} />
	</div>
</section>

<style>
	.net-worth-section {
		margin-bottom: 2rem;
	}

	.growth-section {
		margin-bottom: 2rem;
		animation: fade-in 0.5s ease-out both;
		animation-delay: 0.05s;
	}

	.content-grid {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
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
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
