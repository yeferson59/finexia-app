<script lang="ts">
	import Card from '$lib/ui/card.svelte';
	import { formatPct, type HoldingView } from '../portfolio';

	let {
		holdings,
		formatCurrency,
		onViewAsset,
		onAddAsset
	}: {
		holdings: HoldingView[];
		formatCurrency: (value: number) => string;
		onViewAsset: (symbol: string) => void;
		onAddAsset: () => void;
	} = $props();
</script>

<Card variant="elevated" padding="none">
	<div class="holdings">
		<header class="panel-header">
			<h2>Posiciones</h2>
			<span>{holdings.length} {holdings.length === 1 ? 'activo' : 'activos'}</span>
		</header>

		{#if holdings.length > 0}
			<div class="holdings-list">
				{#each holdings as holding (holding.symbol)}
					<button
						class="holding-row"
						onclick={() => onViewAsset(holding.symbol)}
						aria-label={`Ver detalles de ${holding.symbol}`}
					>
						<div class="holding-main">
							<p class="symbol">{holding.symbol}</p>
							<p class="name">{holding.name}</p>
						</div>
						<div class="bar-wrap" aria-label={`Asignación ${holding.allocation.toFixed(1)}%`}>
							<div class="bar-fill" style={`width: ${holding.allocation}%`}></div>
						</div>
						<p class="metric">{formatCurrency(holding.value)}</p>
						<p class="metric delta {holding.gainLoss >= 0 ? 'positive' : 'negative'}">
							{formatPct(holding.gainLossPct)}
						</p>
						<svg
							class="arrow-icon"
							width="20"
							height="20"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<path d="M9 18l6-6-6-6" />
						</svg>
					</button>
				{/each}
			</div>
		{:else}
			<div class="empty-holdings">
				<p class="empty-text">Este portafolio aún no tiene activos.</p>
				<button onclick={onAddAsset} class="btn-add-asset">
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
					Agregar tu primer activo
				</button>
			</div>
		{/if}
	</div>
</Card>

<style>
	.holdings {
		padding: 1.5rem;
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}

	.panel-header h2 {
		margin: 0;
		font-size: 1.2rem;
		color: var(--text);
	}

	.panel-header span {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.52);
	}

	.holdings-list {
		display: grid;
		gap: 0.75rem;
	}

	.empty-holdings {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		padding: 2.5rem 1.5rem;
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.022);
		border: 1px dashed rgba(212, 145, 42, 0.25);
		text-align: center;
	}

	.empty-text {
		margin: 0;
		font-size: 0.95rem;
		color: rgba(236, 234, 229, 0.6);
	}

	.btn-add-asset {
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

	.btn-add-asset:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.holding-row {
		display: grid;
		grid-template-columns: 1.25fr 1.4fr 0.8fr 0.45fr auto;
		align-items: center;
		gap: 0.9rem;
		padding: 0.85rem;
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.022);
		cursor: pointer;
		transition: all 0.3s ease;
		border: 1px solid rgba(212, 145, 42, 0);
	}

	.holding-row:hover {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(212, 145, 42, 0.2);
		transform: translateX(4px);
	}

	.holding-main .symbol {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 700;
		color: var(--amber-light);
	}

	.holding-main .name {
		margin: 0.15rem 0 0;
		font-size: 0.78rem;
		color: rgba(236, 234, 229, 0.6);
	}

	.bar-wrap {
		height: 8px;
		border-radius: 999px;
		background: rgba(236, 234, 229, 0.12);
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		border-radius: inherit;
		background: linear-gradient(90deg, var(--amber), var(--green));
	}

	.metric {
		margin: 0;
		font-size: 0.84rem;
		font-weight: 600;
		text-align: right;
		color: var(--text);
	}

	.metric.delta {
		color: var(--amber-light);
	}

	.metric.delta.positive {
		color: var(--green);
	}

	.metric.delta.negative {
		color: var(--red);
	}

	.arrow-icon {
		color: rgba(212, 145, 42, 0.4);
		transition: all 0.3s ease;
	}

	.holding-row:hover .arrow-icon {
		color: var(--amber-light);
		transform: translateX(2px);
	}

	@media (max-width: 768px) {
		.holding-row {
			grid-template-columns: 1fr;
			gap: 0.55rem;
		}

		.metric {
			text-align: left;
		}
	}
</style>
