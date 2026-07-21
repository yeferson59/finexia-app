<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Badge from '$lib/ui/badge.svelte';
	import ProgressBar from '$lib/ui/progress-bar.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import { privacy } from '$lib/stores/privacy.svelte';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const portfolios = $derived(data.portfolios ?? []);

	const PER_PAGE = 9;
	let page = $state(1);
	const pagedPortfolios = $derived(portfolios.slice((page - 1) * PER_PAGE, page * PER_PAGE));

	// Aggregate totals across all portfolios
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

	function fmtPct(value: number): string {
		return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
	}

	function openPortfolio(id: string) {
		goto(resolve('/dashboard/portfolios/[id]', { id }));
	}

	function createPortfolio() {
		goto(resolve('/dashboard/portfolios/add'));
	}

	function riskTone(name: string): 'success' | 'warning' | 'danger' | 'neutral' {
		const n = name.toLowerCase();
		if (n.includes('bajo')) return 'success';
		if (n.includes('moderado')) return 'warning';
		if (n.includes('alto')) return 'danger';
		return 'neutral';
	}

	// Each portfolio's share of total market value (for the progress bar)
	function allocation(marketValue: string): number {
		const v = parseFloat(marketValue) || 0;
		return totalMarketValue > 0 ? (v / totalMarketValue) * 100 : 0;
	}

	const TYPE_LABELS: Record<string, string> = {
		stocks: 'Acciones',
		etfs: 'ETFs',
		cryptos: 'Criptomonedas',
		bonds: 'Bonos',
		cash: 'Efectivo',
		forex: 'Forex',
		real_estates: 'Inmobiliario',
		commodities: 'Materias primas',
		stocks_etfs: 'Acciones & ETFs',
		stocks_cryptos: 'Acciones & Cripto',
		stocks_bonds: 'Acciones & Bonos',
		stocks_cash: 'Acciones & Efectivo',
		stocks_real_estates: 'Acciones & Inmobiliario',
		stocks_commodities: 'Acciones & Materias primas',
		etfs_cryptos: 'ETFs & Cripto',
		etfs_bonds: 'ETFs & Bonos',
		etfs_cash: 'ETFs & Efectivo',
		etfs_real_estates: 'ETFs & Inmobiliario',
		etfs_commodities: 'ETFs & Materias primas',
		cryptos_bonds: 'Cripto & Bonos',
		cryptos_cash: 'Cripto & Efectivo',
		cryptos_real_estates: 'Cripto & Inmobiliario',
		cryptos_commodities: 'Cripto & Materias primas',
		bonds_cash: 'Bonos & Efectivo',
		bonds_real_estates: 'Bonos & Inmobiliario',
		bonds_commodities: 'Bonos & Materias primas',
		cash_real_estates: 'Efectivo & Inmobiliario',
		cash_commodities: 'Efectivo & Materias primas',
		real_estates_commodities: 'Inmobiliario & Materias primas',
		forex_stocks: 'Forex & Acciones',
		forex_etfs: 'Forex & ETFs',
		forex_cryptos: 'Forex & Cripto',
		forex_bonds: 'Forex & Bonos',
		forex_cash: 'Forex & Efectivo',
		forex_real_states: 'Forex & Inmobiliario',
		forex_commodities: 'Forex & Materias primas',
		diversified: 'Diversificado'
	};

	function formatType(type: string): string {
		return TYPE_LABELS[type] ?? type.replace(/_/g, ' ');
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
			{fmtPct(totalGainLossPct)} sobre costo
		</p>
	</article>
</section>

<section class="portfolios-section">
	<h2 class="section-title">Tus Portafolios</h2>

	<div class="portfolios-grid">
		{#each pagedPortfolios as portfolio (portfolio.id)}
			{@const marketValue = parseFloat(portfolio.totalMarketValue) || 0}
			{@const gainLoss = parseFloat(portfolio.totalGainLoss) || 0}
			{@const gainLossPct = parseFloat(portfolio.totalGainLossPct) || 0}
			{@const alloc = allocation(portfolio.totalMarketValue)}
			<button
				class="portfolio-card"
				onclick={() => openPortfolio(portfolio.id)}
				aria-label={`Abrir ${portfolio.name}`}
			>
				<div class="card-header">
					<div class="portfolio-info">
						<h3 class="portfolio-name">{portfolio.name}</h3>
						<p class="portfolio-type">{formatType(portfolio.type)}</p>
					</div>
					<div class="card-header-right">
						<Badge tone={riskTone(portfolio.riskName)} size="md">{portfolio.riskName}</Badge>
						<svg
							class="arrow-icon"
							width="18"
							height="18"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<path d="M9 18l6-6-6-6" />
						</svg>
					</div>
				</div>

				<div class="card-metrics">
					<div class="metric">
						<p class="label">Valor</p>
						<p class="value">{fmt(marketValue, portfolio.baseCurrency)}</p>
					</div>

					<div class="metric">
						<p class="label">Activos</p>
						<p class="value">{portfolio.totalPositions}</p>
					</div>

					<div class="metric">
						<p class="label">ROI</p>
						<p class="value {gainLoss >= 0 ? 'positive' : 'negative'}">{fmtPct(gainLossPct)}</p>
					</div>
				</div>

				<ProgressBar
					value={alloc}
					label={`${alloc.toFixed(1)}% del total`}
					ariaLabel={`Asignación de ${portfolio.name}`}
				/>
			</button>
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

	.portfolio-card {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		padding: 1.35rem;
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		cursor: pointer;
		transition: all 0.3s ease;
		text-align: left;
	}

	.portfolio-card:hover {
		background: var(--surface-2);
		border-color: rgba(212, 145, 42, 0.3);
		transform: translateY(-4px);
		box-shadow:
			0 30px 80px rgba(0, 0, 0, 0.4),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
	}

	.card-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.portfolio-info {
		flex: 1;
	}

	.portfolio-name {
		margin: 0 0 0.3rem;
		font-size: 1.15rem;
		color: var(--text);
		font-weight: 600;
	}

	.portfolio-type {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.52);
	}

	.card-metrics {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1rem;
		padding: 1rem 0;
		border-top: 1px solid var(--border);
		border-bottom: 1px solid var(--border);
	}

	.metric {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.metric .label {
		margin: 0;
		font-size: 0.7rem;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.46);
	}

	.metric .value {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 600;
		color: var(--amber-light);
	}

	.metric .value.positive {
		color: var(--green);
	}

	.metric .value.negative {
		color: var(--red);
	}

	.card-header-right {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		flex-shrink: 0;
	}

	.arrow-icon {
		color: rgba(212, 145, 42, 0.3);
		transition: all 0.3s ease;
		flex-shrink: 0;
	}

	.portfolio-card:hover .arrow-icon {
		color: var(--amber-light);
		transform: translateX(4px);
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

		.card-metrics {
			grid-template-columns: repeat(3, 1fr);
		}
	}
</style>
