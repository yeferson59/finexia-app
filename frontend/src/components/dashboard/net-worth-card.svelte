<script lang="ts">
	import CardHeader from '$components/ui/card-header.svelte';
	import Badge from '$components/ui/badge.svelte';
	import Stat from '$components/ui/stat.svelte';

	interface PortfolioSummary {
		id: string;
		name: string;
		totalMarketValue: string;
		totalCostBase: string;
		totalGainLoss: string;
		totalGainLossPct: string;
		totalPositions: number;
	}

	const { summaries = [] }: { summaries: PortfolioSummary[] } = $props();

	const netWorth = $derived(
		summaries.reduce((acc, s) => acc + parseFloat(s.totalMarketValue || '0'), 0)
	);
	const totalGainLoss = $derived(
		summaries.reduce((acc, s) => acc + parseFloat(s.totalGainLoss || '0'), 0)
	);
	const totalCostBase = $derived(
		summaries.reduce((acc, s) => acc + parseFloat(s.totalCostBase || '0'), 0)
	);
	const totalPositions = $derived(summaries.reduce((acc, s) => acc + (s.totalPositions ?? 0), 0));
	const gainLossPct = $derived(totalCostBase > 0 ? (totalGainLoss / totalCostBase) * 100 : 0);
	const isIncreasing = $derived(totalGainLoss >= 0);
</script>

<div class="net-worth-card">
	<CardHeader eyebrow="Patrimonio total" title="Patrimonio Neto">
		{#snippet action()}
			<Badge tone="neutral" pill={false}>Acumulado</Badge>
		{/snippet}
	</CardHeader>

	<div class="net-worth-content">
		<div class="main-metric">
			<h1 class="amount">
				${new Intl.NumberFormat('es-CO', {
					minimumFractionDigits: 2,
					maximumFractionDigits: 2
				}).format(netWorth)}
			</h1>
			{#if summaries.length > 0}
				<p class="amount-delta" class:positive={isIncreasing} class:negative={!isIncreasing}>
					{isIncreasing ? '+' : '−'}${new Intl.NumberFormat('es-CO', {
						minimumFractionDigits: 2,
						maximumFractionDigits: 2
					}).format(Math.abs(totalGainLoss))}
					· {new Intl.NumberFormat('es-CO', {
						minimumFractionDigits: 2,
						maximumFractionDigits: 2
					}).format(Math.abs(gainLossPct))}% total
				</p>
			{:else}
				<p class="amount-delta neutral">Sin portafolios registrados</p>
			{/if}
		</div>

		<div class="metric-stats">
			<Stat label="Portafolios" value={String(summaries.length)} />
			<Stat label="Posiciones" tone="highlight" value={String(totalPositions)} />
			<Stat
				label="Ganancia"
				tone={isIncreasing ? 'positive' : 'negative'}
				value="{isIncreasing ? '+' : ''}{new Intl.NumberFormat('es-CO', {
					minimumFractionDigits: 2,
					maximumFractionDigits: 2
				}).format(gainLossPct)}%"
			/>
		</div>
	</div>

	<div class="card-footer">
		<button class="action-button primary">Agregar fondos</button>
		<button class="action-button secondary">Ver detalles</button>
	</div>
</div>

<style>
	.net-worth-card {
		position: relative;
		overflow: hidden;
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		padding: 2rem;
		backdrop-filter: blur(10px);
	}

	/* Warm amber wash anchoring the hero figure */
	.net-worth-card::before {
		content: '';
		position: absolute;
		inset: 0;
		background: radial-gradient(
			ellipse 60% 90% at 0% 0%,
			rgba(212, 145, 42, 0.07),
			transparent 55%
		);
		pointer-events: none;
	}

	.net-worth-card > * {
		position: relative;
	}

	.net-worth-content {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 3rem;
		margin-bottom: 2rem;
		align-items: end;
	}

	.main-metric {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		min-width: 0;
	}

	.amount {
		font-family: var(--font-mono);
		font-size: clamp(1.65rem, 7vw, 3rem);
		overflow-wrap: anywhere;
		font-weight: 600;
		color: var(--text);
		margin: 0;
		line-height: 1;
		letter-spacing: -0.03em;
		font-variant-numeric: tabular-nums;
	}

	.amount-delta {
		font-size: 0.85rem;
		font-weight: 400;
		margin: 0;
	}

	.amount-delta.positive {
		color: var(--green);
	}

	.amount-delta.negative {
		color: var(--red);
	}

	.amount-delta.neutral {
		color: rgba(236, 234, 229, 0.5);
	}

	.metric-stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1.5rem;
	}

	.card-footer {
		display: flex;
		gap: 0.75rem;
	}

	.action-button {
		flex: 1;
		padding: 0.75rem 1.5rem;
		border-radius: 6px;
		font-weight: 600;
		font-size: 0.85rem;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease,
			color 0.2s ease,
			transform 0.15s ease;
		font-family: var(--font-body);
	}

	.action-button.primary {
		border: none;
		background: var(--amber);
		color: #0d0800;
	}

	.action-button.primary:hover {
		background: var(--amber-light);
		transform: translateY(-1px);
	}

	.action-button.primary:active {
		transform: none;
	}

	.action-button.secondary {
		background: transparent;
		border: 1px solid var(--border-strong);
		color: var(--text);
	}

	.action-button.secondary:hover {
		background: rgba(212, 145, 42, 0.06);
		border-color: rgba(212, 145, 42, 0.4);
		color: var(--amber-light);
	}

	@media (prefers-reduced-motion: reduce) {
		.action-button.primary:hover {
			transform: none;
		}
	}

	@media (max-width: 1024px) {
		.net-worth-content {
			grid-template-columns: 1fr;
			gap: 2rem;
		}

		.metric-stats {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}
	}

	@media (max-width: 768px) {
		.net-worth-card {
			padding: 1.5rem;
		}

		.net-worth-content {
			gap: 1.5rem;
		}

		.action-button {
			padding: 0.75rem 1rem;
			font-size: 0.85rem;
		}
	}

	@media (max-width: 480px) {
		.metric-stats {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
