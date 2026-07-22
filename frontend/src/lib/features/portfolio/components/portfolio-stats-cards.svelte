<script lang="ts">
	import Card from '$lib/ui/card.svelte';
	import { formatCalendarDate } from '$lib/utils';
	import { formatPct, type HoldingView, type TopTransactionData } from '../portfolio';

	let {
		totalValue,
		totalCost,
		capitalPct,
		gainPct,
		bestHolding,
		worstHolding,
		topConcentration,
		topTransaction,
		formatCurrency
	}: {
		totalValue: number;
		totalCost: number;
		capitalPct: number;
		gainPct: number;
		bestHolding: HoldingView | null;
		worstHolding: HoldingView | null;
		topConcentration: HoldingView | null;
		topTransaction: TopTransactionData | null;
		formatCurrency: (value: number) => string;
	} = $props();
</script>

<section class="stats-grid">
	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Capital invertido</p>
		<h2 class="hero-value">{formatCurrency(totalCost)}</h2>
		<div class="progress-track">
			<div
				class="progress-fill"
				style="width: {totalValue > 0 ? Math.min((totalCost / totalValue) * 100, 100) : 0}%"
			></div>
		</div>
		<p class="hero-delta">Valor actual: {formatCurrency(totalValue)}</p>
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Composición del portafolio</p>
		{#if totalValue > 0}
			<div class="composition-bar">
				<div
					class="comp-segment comp-capital"
					style="width: {Math.max(capitalPct, 0)}%"
					title="Capital: {capitalPct.toFixed(1)}%"
				></div>
				<div
					class="comp-segment {gainPct >= 0 ? 'comp-gain' : 'comp-loss'}"
					style="width: {Math.abs(gainPct)}%"
					title="Ganancia: {gainPct.toFixed(1)}%"
				></div>
			</div>
			<p class="comp-labels">
				<span class="comp-label-capital">{capitalPct.toFixed(1)}% capital</span>
				<span class="comp-sep">·</span>
				<span class={gainPct >= 0 ? 'comp-label-gain' : 'comp-label-loss'}
					>{gainPct >= 0 ? '+' : ''}{gainPct.toFixed(1)}% {gainPct >= 0
						? 'ganancia'
						: 'pérdida'}</span
				>
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin datos suficientes</p>
		{/if}
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Mejor activo</p>
		{#if bestHolding}
			<h2 class="hero-value">{bestHolding.symbol}</h2>
			<p class="hero-delta {bestHolding.gainLossPct >= 0 ? 'positive' : 'negative'}">
				{formatPct(bestHolding.gainLossPct)} · {bestHolding.name}
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin activos</p>
		{/if}
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Peor activo</p>
		{#if worstHolding}
			<h2 class="hero-value">{worstHolding.symbol}</h2>
			<p class="hero-delta {worstHolding.gainLossPct >= 0 ? 'positive' : 'negative'}">
				{formatPct(worstHolding.gainLossPct)} · {worstHolding.name}
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin activos</p>
		{/if}
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Concentración</p>
		{#if topConcentration}
			<h2 class="hero-value">{topConcentration.allocation.toFixed(1)}%</h2>
			<p class="hero-delta">
				{topConcentration.symbol} · {topConcentration.name}
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin activos</p>
		{/if}
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Transacción más alta</p>
		{#if topTransaction}
			<h2 class="hero-value">{formatCurrency(parseFloat(topTransaction.value))}</h2>
			<p class="hero-delta">
				{topTransaction.assetTicker} · {topTransaction.type} ·
				{formatCalendarDate(topTransaction.transactionDate, {
					year: 'numeric',
					month: 'short',
					day: 'numeric'
				})}
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin transacciones</p>
		{/if}
	</Card>
</section>

<style>
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
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

	.progress-track {
		height: 6px;
		border-radius: 999px;
		background: rgba(236, 234, 229, 0.1);
		overflow: hidden;
		margin: 0.6rem 0 0.5rem;
	}

	.progress-fill {
		height: 100%;
		border-radius: inherit;
		background: var(--amber);
		transition: width 0.4s ease;
	}

	.composition-bar {
		display: flex;
		height: 10px;
		border-radius: 999px;
		overflow: hidden;
		background: rgba(236, 234, 229, 0.1);
		gap: 2px;
		margin: 0.6rem 0 0.5rem;
	}

	.comp-segment {
		height: 100%;
		border-radius: 999px;
		min-width: 2px;
		transition: width 0.4s ease;
	}

	.comp-capital {
		background: var(--amber);
	}

	.comp-gain {
		background: var(--green);
	}

	.comp-loss {
		background: var(--red);
	}

	.comp-labels {
		margin: 0;
		font-size: 0.78rem;
		display: flex;
		align-items: center;
		gap: 0.4rem;
		flex-wrap: wrap;
	}

	.comp-sep {
		color: rgba(236, 234, 229, 0.3);
	}

	.comp-label-capital {
		color: var(--amber);
		font-weight: 600;
	}

	.comp-label-gain {
		color: var(--green);
		font-weight: 600;
	}

	.comp-label-loss {
		color: var(--red);
		font-weight: 600;
	}

	@media (max-width: 1024px) {
		.stats-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
