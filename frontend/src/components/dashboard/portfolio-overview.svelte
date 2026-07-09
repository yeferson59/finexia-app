<script lang="ts">
	import Card from '$components/ui/card.svelte';
	import CardHeader from '$components/ui/card-header.svelte';
	import Stat from '$components/ui/stat.svelte';
	import { privacy } from '$lib/stores/privacy.svelte';
	import { formatCurrency } from '$lib/utils';

	interface PortfolioSummary {
		id: string;
		name: string;
		type: string;
		baseCurrency: string;
		displayCurrency?: string;
		totalPositions: number;
		totalCostBase: string;
		totalMarketValue: string;
		totalGainLoss: string;
		totalGainLossPct: string;
	}

	const { summaries = [], currency = 'USD' }: { summaries: PortfolioSummary[]; currency?: string } =
		$props();

	const totalInvested = $derived(
		summaries.reduce((acc, s) => acc + parseFloat(s.totalCostBase || '0'), 0)
	);
	const totalValue = $derived(
		summaries.reduce((acc, s) => acc + parseFloat(s.totalMarketValue || '0'), 0)
	);
	const totalGainLoss = $derived(
		summaries.reduce((acc, s) => acc + parseFloat(s.totalGainLoss || '0'), 0)
	);
	const totalGainLossPct = $derived(totalInvested > 0 ? (totalGainLoss / totalInvested) * 100 : 0);

	function fmtMoney(value: number, currencyCode = currency): string {
		return privacy.money(formatCurrency(value, currencyCode));
	}

	function fmtPct(value: number): string {
		return new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		}).format(value);
	}
</script>

<Card>
	<CardHeader eyebrow="Resumen" title="Portafolios" />

	{#if summaries.length === 0}
		<div class="empty-state">
			<svg
				width="48"
				height="48"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="1.5"
			>
				<rect x="2" y="3" width="20" height="14" rx="2" />
				<path d="M8 21h8M12 17v4" />
			</svg>
			<p>Sin portafolios registrados</p>
		</div>
	{:else}
		<div class="portfolio-list">
			<div class="list-header">
				<span>Nombre</span>
				<span>Tipo</span>
				<span class="align-right">Valor actual</span>
				<span class="align-right">Invertido</span>
				<span class="align-right">Ganancia/Pérdida</span>
			</div>

			{#each summaries as s (s.id)}
				{@const gainLoss = parseFloat(s.totalGainLoss || '0')}
				{@const marketValue = parseFloat(s.totalMarketValue || '0')}
				{@const costBase = parseFloat(s.totalCostBase || '0')}
				{@const pct = parseFloat(s.totalGainLossPct || '0')}
				{@const isUp = gainLoss >= 0}
				{@const rowCurrency = s.displayCurrency || s.baseCurrency}
				<div class="list-row">
					<div class="portfolio-name">
						<span class="name">{s.name}</span>
						<span class="currency">{rowCurrency}</span>
					</div>
					<span class="type-badge">{s.type}</span>
					<span class="align-right mono">{fmtMoney(marketValue, rowCurrency)}</span>
					<span class="align-right mono dim">{fmtMoney(costBase, rowCurrency)}</span>
					<div class="align-right gain-cell" class:positive={isUp} class:negative={!isUp}>
						<span class="mono">{isUp ? '+' : '−'}{fmtMoney(Math.abs(gainLoss), rowCurrency)}</span>
						<span class="pct">{isUp ? '+' : ''}{fmtPct(pct)}%</span>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<div class="chart-stats">
		<Stat label="Total Invertido" value={fmtMoney(totalInvested)} />
		<Stat label="Valor Actual" tone="highlight" value={fmtMoney(totalValue)} />
		<Stat
			label="Ganancia Total"
			tone={totalGainLoss >= 0 ? 'positive' : 'negative'}
			value="{totalGainLoss >= 0 ? '+' : ''}{fmtPct(totalGainLossPct)}%"
		/>
	</div>
</Card>

<style>
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		padding: 3rem 2rem;
		color: rgba(236, 234, 229, 0.4);
		text-align: center;
	}

	.empty-state p {
		margin: 0;
		font-size: 0.9rem;
	}

	.portfolio-list {
		margin-bottom: 1.5rem;
		overflow-x: auto;
	}

	.list-header {
		display: grid;
		grid-template-columns: minmax(0, 2fr) minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1fr) minmax(
				0,
				1.2fr
			);
		gap: 1rem;
		padding: 0.5rem 1rem;
		font-size: 0.72rem;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: rgba(236, 234, 229, 0.45);
		font-weight: 600;
		border-bottom: 1px solid var(--border);
		margin-bottom: 0.25rem;
	}

	.list-row {
		display: grid;
		grid-template-columns: minmax(0, 2fr) minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1fr) minmax(
				0,
				1.2fr
			);
		gap: 1rem;
		padding: 0.9rem 1rem;
		border-radius: 8px;
		transition: background 0.2s ease;
		align-items: center;
	}

	.list-row:hover {
		background: rgba(255, 255, 255, 0.03);
	}

	.portfolio-name {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		min-width: 0;
	}

	.name {
		font-weight: 600;
		color: var(--text);
		font-size: 0.9rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.currency {
		font-size: 0.72rem;
		color: rgba(236, 234, 229, 0.45);
		text-transform: uppercase;
		font-family: var(--font-mono);
	}

	.type-badge {
		font-size: 0.72rem;
		color: rgba(212, 145, 42, 0.75);
		text-transform: uppercase;
		letter-spacing: 0.3px;
		font-weight: 600;
	}

	.align-right {
		text-align: right;
	}

	.mono {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		font-size: 0.85rem;
	}

	.dim {
		color: rgba(236, 234, 229, 0.55);
	}

	.gain-cell {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		min-width: 0;
	}

	.gain-cell .mono {
		overflow-wrap: anywhere;
	}

	.gain-cell.positive .mono,
	.gain-cell.positive .pct {
		color: var(--green);
	}

	.gain-cell.negative .mono,
	.gain-cell.negative .pct {
		color: var(--red);
	}

	.pct {
		font-family: var(--font-mono);
		font-size: 0.72rem;
		font-variant-numeric: tabular-nums;
	}

	.chart-stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1.5rem;
		padding-top: 1.5rem;
		border-top: 1px solid var(--border);
	}

	@media (max-width: 1024px) {
		.list-header,
		.list-row {
			grid-template-columns: minmax(0, 2fr) minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1.2fr);
		}

		.list-header :nth-child(4),
		.list-row :nth-child(4) {
			display: none;
		}

		.chart-stats {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 768px) {
		.list-header {
			display: none;
		}

		.list-row {
			grid-template-columns: minmax(0, 1fr) minmax(0, auto);
			grid-template-rows: auto auto;
		}

		.list-row :nth-child(2),
		.list-row :nth-child(3),
		.list-row :nth-child(4) {
			display: none;
		}

		.chart-stats {
			grid-template-columns: 1fr;
		}
	}
</style>
