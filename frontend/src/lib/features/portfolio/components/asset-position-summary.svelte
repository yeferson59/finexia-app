<script lang="ts">
	import { formatPct } from '../portfolio';
	import type { AssetPosition } from '../asset';

	let {
		position,
		formatCurrency
	}: { position: AssetPosition; formatCurrency: (value: number, decimals?: number) => string } =
		$props();
</script>

<section class="panel">
	<header class="panel-header">
		<h2>Resumen de Posición</h2>
	</header>

	<div class="metrics-grid">
		<article class="metric-card">
			<p class="metric-label">Cantidad total</p>
			<p class="metric-value">
				{position.totalQty.toLocaleString('es-CO', { maximumFractionDigits: 8 })}
				<span class="metric-unit">{position.ticker}</span>
			</p>
		</article>

		<article class="metric-card">
			<p class="metric-label">Precio promedio</p>
			<p class="metric-value">{formatCurrency(position.averageCost)}</p>
		</article>

		<article class="metric-card">
			<p class="metric-label">Precio actual</p>
			<p class="metric-value">{formatCurrency(position.marketPrice)}</p>
		</article>

		<article class="metric-card">
			<p class="metric-label">Costo total</p>
			<p class="metric-value">{formatCurrency(position.totalCost, 0)}</p>
		</article>

		<article class="metric-card">
			<p class="metric-label">Valor de mercado</p>
			<p class="metric-value">{formatCurrency(position.totalValue, 0)}</p>
		</article>

		<article class="metric-card gain">
			<p class="metric-label">Ganancia / Pérdida</p>
			<p class="metric-value {position.gainLoss >= 0 ? 'positive' : 'negative'}">
				{formatCurrency(position.gainLoss, 0)}
			</p>
			<p class="metric-pct {position.gainLoss >= 0 ? 'positive' : 'negative'}">
				{formatPct(position.gainLossPercent)}
			</p>
		</article>
	</div>
</section>

<style>
	.panel {
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		padding: 1.75rem;
		margin-bottom: 1.5rem;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--border);
	}

	.panel-header h2 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
	}

	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		gap: 1rem;
	}

	.metric-card {
		background: rgba(212, 145, 42, 0.06);
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 10px;
		padding: 1.25rem;
		text-align: center;
	}

	.metric-label {
		margin: 0 0 0.6rem;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.metric-value {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--amber);
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.metric-value.positive {
		color: var(--green);
	}

	.metric-value.negative {
		color: var(--red);
	}

	.metric-unit {
		font-size: 0.7rem;
		color: rgba(236, 234, 229, 0.4);
		margin-left: 0.25rem;
	}

	.metric-pct {
		margin: 0.25rem 0 0;
		font-size: 0.85rem;
		font-weight: 600;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.metric-pct.positive {
		color: var(--green);
	}

	.metric-pct.negative {
		color: var(--red);
	}

	@media (max-width: 768px) {
		.metrics-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 480px) {
		.metrics-grid {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
