<script lang="ts">
	import Badge from '$lib/ui/badge.svelte';
	import ProgressBar from '$lib/ui/progress-bar.svelte';
	import type { PortfolioSummary } from '$lib/api/types';
	import { formatPortfolioType } from '$lib/shared/format/portfolio-type';
	import { formatPct, riskTone } from '../portfolio';

	let {
		portfolio,
		allocation,
		formatCurrency,
		onOpen
	}: {
		portfolio: PortfolioSummary;
		/** Porcentaje que representa este portafolio sobre el valor total. */
		allocation: number;
		formatCurrency: (value: number, currency?: string) => string;
		onOpen: (id: string) => void;
	} = $props();

	const marketValue = $derived(parseFloat(portfolio.totalMarketValue) || 0);
	const gainLoss = $derived(parseFloat(portfolio.totalGainLoss) || 0);
	const gainLossPct = $derived(parseFloat(portfolio.totalGainLossPct) || 0);
</script>

<button
	class="portfolio-card"
	onclick={() => onOpen(portfolio.id)}
	aria-label={`Abrir ${portfolio.name}`}
>
	<div class="card-header">
		<div class="portfolio-info">
			<h3 class="portfolio-name">{portfolio.name}</h3>
			<p class="portfolio-type">{formatPortfolioType(portfolio.type)}</p>
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
			<p class="value">{formatCurrency(marketValue, portfolio.baseCurrency)}</p>
		</div>

		<div class="metric">
			<p class="label">Activos</p>
			<p class="value">{portfolio.totalPositions}</p>
		</div>

		<div class="metric">
			<p class="label">ROI</p>
			<p class="value {gainLoss >= 0 ? 'positive' : 'negative'}">{formatPct(gainLossPct)}</p>
		</div>
	</div>

	<ProgressBar
		value={allocation}
		label={`${allocation.toFixed(1)}% del total`}
		ariaLabel={`Asignación de ${portfolio.name}`}
	/>
</button>

<style>
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

	@media (max-width: 768px) {
		.card-metrics {
			grid-template-columns: repeat(3, 1fr);
		}
	}
</style>
