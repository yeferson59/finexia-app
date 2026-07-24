<script lang="ts">
	import { formatPct } from '../portfolio';
	import type { AssetPosition } from '../asset';

	let { position, transactionsCount }: { position: AssetPosition; transactionsCount: number } =
		$props();
</script>

<section class="panel">
	<header class="panel-header">
		<h2>Información del Activo</h2>
	</header>

	<div class="perf-grid">
		<article class="perf-card">
			<h3>Tipo</h3>
			<p class="perf-value">{position.assetType}</p>
		</article>

		<article class="perf-card">
			<h3>Exchange</h3>
			<p class="perf-value">{position.exchange || '—'}</p>
		</article>

		<article class="perf-card">
			<h3>Moneda</h3>
			<p class="perf-value">{position.currency}</p>
		</article>

		<article class="perf-card">
			<h3>Asignación</h3>
			<p class="perf-value">{position.allocation.toFixed(1)}%</p>
			<div class="bar-wrap">
				<div class="bar-fill" style="width: {Math.min(position.allocation, 100)}%"></div>
			</div>
		</article>

		<article class="perf-card">
			<h3>Transacciones</h3>
			<p class="perf-value">{transactionsCount}</p>
		</article>

		<article class="perf-card">
			<h3>ROI</h3>
			<p class="perf-value {position.gainLossPercent >= 0 ? 'positive' : 'negative'}">
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

	.perf-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 1rem;
	}

	.perf-card {
		background: rgba(255, 255, 255, 0.022);
		border: 1px solid var(--border);
		border-radius: 10px;
		padding: 1.25rem;
		text-align: center;
	}

	.perf-card h3 {
		margin: 0 0 0.6rem;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.perf-value {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 700;
		color: var(--amber);
	}

	.perf-value.positive {
		color: var(--green);
	}

	.perf-value.negative {
		color: var(--red);
	}

	.bar-wrap {
		width: 100%;
		height: 4px;
		background: var(--border);
		border-radius: 2px;
		margin-top: 0.6rem;
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		background: linear-gradient(90deg, var(--amber) 0%, var(--amber-light, var(--amber)) 100%);
		border-radius: 2px;
	}

	@media (max-width: 768px) {
		.perf-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 480px) {
		.perf-grid {
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
