<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { PortfolioCard, formatPct } from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const portfolios = $derived(data.portfolios ?? []);

	const PER_PAGE = 9;
	let page = $state(1);
	const pagedPortfolios = $derived(portfolios.slice((page - 1) * PER_PAGE, page * PER_PAGE));

	// Totales agregados de todos los portafolios
	const totalMarketValue = $derived(
		portfolios.reduce((s, p) => s + (parseFloat(p.totalMarketValue) || 0), 0)
	);
	const totalCostBase = $derived(
		portfolios.reduce((s, p) => s + (parseFloat(p.totalCostBase) || 0), 0)
	);
	const totalGainLoss = $derived(totalMarketValue - totalCostBase);
	const totalGainLossPct = $derived(totalCostBase > 0 ? (totalGainLoss / totalCostBase) * 100 : 0);

	function fmt(value: number, currency = 'USD'): string {
		return privacy.money(
			new Intl.NumberFormat('es-CO', {
				style: 'currency',
				currency,
				minimumFractionDigits: 0,
				maximumFractionDigits: 0
			}).format(value)
		);
	}

	function openPortfolio(id: string) {
		goto(resolve('/dashboard/portfolios/[id]', { id }));
	}

	function createPortfolio() {
		goto(resolve('/dashboard/portfolios/add'));
	}

	// Peso de cada portafolio sobre el valor de mercado total (barra de progreso)
	function allocation(marketValue: string): number {
		const v = parseFloat(marketValue) || 0;
		return totalMarketValue > 0 ? (v / totalMarketValue) * 100 : 0;
	}
</script>

<svelte:head>
	<title>Portafolios - FINEXIA</title>
	<meta name="description" content="Gestión de múltiples portafolios de inversión" />
</svelte:head>

<PageHeader
	title="Portafolios"
	subtitle="Gestiona tus múltiples portafolios de inversión en un solo lugar."
>
	{#snippet actions()}
		<button onclick={createPortfolio} class="btn-create-portfolio">
			<svg
				width="18"
				height="18"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<path d="M12 5v14M5 12h14" />
			</svg>
			Crear Portafolio
		</button>
	{/snippet}
</PageHeader>

<section class="summary-cards">
	<article class="panel summary-card">
		<p class="eyebrow">Valor de mercado</p>
		<h2 class="hero-value">{fmt(totalMarketValue)}</h2>
		<p class="hero-delta">Costo base: {fmt(totalCostBase)}</p>
	</article>

	<article class="panel summary-card">
		<p class="eyebrow">Portafolios activos</p>
		<h2 class="hero-value">{portfolios.length}</h2>
		<p class="hero-delta">
			{portfolios.reduce((s, p) => s + p.totalPositions, 0)} activos en total
		</p>
	</article>

	<article class="panel summary-card">
		<p class="eyebrow">Ganancia / Pérdida total</p>
		<h2 class="hero-value {totalGainLoss >= 0 ? 'positive' : 'negative'}">
			{fmt(totalGainLoss)}
		</h2>
		<p class="hero-delta {totalGainLoss >= 0 ? 'positive' : 'negative'}">
			{formatPct(totalGainLossPct)} sobre costo
		</p>
	</article>
</section>

<section class="portfolios-section">
	<h2 class="section-title">Tus Portafolios</h2>

	<div class="portfolios-grid">
		{#each pagedPortfolios as portfolio (portfolio.id)}
			<PortfolioCard
				{portfolio}
				allocation={allocation(portfolio.totalMarketValue)}
				formatCurrency={fmt}
				onOpen={openPortfolio}
			/>
		{/each}
	</div>

	<Pagination bind:page total={portfolios.length} perPage={PER_PAGE} label="portafolios" />
</section>

<style>
	.btn-create-portfolio {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.85rem 1.5rem;
		border: none;
		border-radius: 10px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
		white-space: nowrap;
	}

	.btn-create-portfolio:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.summary-cards {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
		margin-bottom: 3rem;
	}

	.summary-card {
		padding: 1.35rem;
	}

	.eyebrow {
		margin: 0 0 0.55rem;
		font-size: 0.72rem;
		letter-spacing: 0.7px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.46);
	}

	.hero-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		margin: 0;
		font-size: 1.6rem;
		color: var(--text);
	}

	.hero-value.positive {
		color: var(--green);
	}

	.hero-value.negative {
		color: var(--red);
	}

	.hero-delta {
		margin: 0.4rem 0 0;
		font-size: 0.82rem;
		color: rgba(236, 234, 229, 0.55);
	}

	.hero-delta.positive {
		color: var(--green);
	}

	.hero-delta.negative {
		color: var(--red);
	}

	.panel {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
	}

	.portfolios-section {
		margin-top: 2rem;
	}

	.section-title {
		font-size: 1.3rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
		margin: 0 0 1.5rem;
	}

	.portfolios-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 1.5rem;
	}

	@media (max-width: 1024px) {
		.summary-cards {
			grid-template-columns: 1fr;
		}

		.portfolios-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.summary-cards {
			grid-template-columns: 1fr;
		}

		.portfolios-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
