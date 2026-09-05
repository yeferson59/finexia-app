<script lang="ts">
	import type { AssetPosition } from '../asset';

	let {
		position,
		formatAmount
	}: {
		position: AssetPosition;
		formatAmount: (value: number, currency: string, decimals?: number) => string;
	} = $props();

	// El precio de mercado se cotiza en la moneda del activo, no en la del
	// portafolio: MC.FR vale 461,65 € y pintarlo como dólares es otro precio.
	const marketPrice = $derived(formatAmount(position.marketPrice, position.currency));
</script>

<div class="header-section">
	<div class="header-content">
		<div class="symbol-info">
			<h1>{position.ticker}</h1>
			<p class="asset-name">{position.name}</p>
			<span class="asset-badge">{position.assetType} · {position.exchange}</span>
		</div>

		<div class="price-display">
			<p class="current-price">{marketPrice}</p>
			<p class="price-label">Precio de mercado · {position.currency}</p>
		</div>
	</div>
</div>

<style>
	.header-section {
		margin-bottom: 1.5rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border-strong);
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.symbol-info h1 {
		margin: 0 0 0.25rem;
		font-size: 2.25rem;
		font-weight: 700;
		color: var(--amber);
		letter-spacing: -0.5px;
		font-family: var(--font-display);
	}

	.asset-name {
		margin: 0 0 0.5rem;
		color: rgba(236, 234, 229, 0.7);
		font-size: 1rem;
	}

	.asset-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.25);
		color: var(--amber);
		font-size: 0.8rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.price-display {
		text-align: right;
	}

	.current-price {
		margin: 0 0 0.25rem;
		font-size: 2rem;
		font-weight: 700;
		color: var(--text);
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.price-label {
		margin: 0;
		font-size: 0.8rem;
		color: var(--text-dim);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	@media (max-width: 768px) {
		.header-content {
			flex-direction: column;
			align-items: flex-start;
		}

		.price-display {
			text-align: left;
		}
	}
</style>
