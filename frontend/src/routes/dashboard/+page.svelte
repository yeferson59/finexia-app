<script lang="ts">
	import {
		WealthHeadline,
		Breakdown,
		RecentActivity,
		PortfolioGrowth,
		MarketKeyNotice
	} from '$lib/features/dashboard';

	import type { PageProps } from './$types';

	const { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Panel - FINEXIA</title>
	<meta name="description" content="Tu panel de control de inversiones y patrimonio" />
</svelte:head>

<MarketKeyNotice hasUsableKey={data.hasUsableKey} hasBrokenKey={data.hasBrokenKey} />

<WealthHeadline
	summaries={data.portfolioSummaries}
	currency={data.currency}
	displayRate={data.displayRate}
	series={data.portfolioGrowth.points}
/>

<Breakdown
	platforms={data.platforms}
	summaries={data.portfolioSummaries}
	allocation={data.allocation}
	currency={data.currency}
/>

<div class="lower">
	<section class="growth" aria-label="Crecimiento del portafolio">
		<PortfolioGrowth
			bare
			data={data.portfolioGrowth.points}
			summary={data.portfolioGrowth.summary}
		/>
	</section>

	<RecentActivity transactions={data.recentTransactions} />
</div>

<style>
	/*
	 * Las dos lecturas secundarias, una al lado de otra: cómo ha ido el
	 * patrimonio en el tiempo y qué ha pasado esta semana. La gráfica manda,
	 * así que se lleva el ancho que sobra; el extracto tiene un ancho fijo
	 * porque son cifras cortas y crecer no lo mejora.
	 */
	.lower {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 300px;
		gap: 3rem;
		padding-top: 2rem;
	}

	.growth {
		min-width: 0;
	}

	@media (max-width: 1080px) {
		.lower {
			grid-template-columns: minmax(0, 1fr);
			gap: 2.5rem;
		}
	}
</style>
