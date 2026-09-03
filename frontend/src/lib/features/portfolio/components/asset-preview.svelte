<script lang="ts">
	import type { Asset } from '$lib/api/types';

	let {
		asset,
		formatCurrency
	}: { asset: Asset; formatCurrency: (value: number, code?: string) => string } = $props();

	// El precio de mercado está en la moneda del activo, que no tiene por qué
	// ser aquella en la que se pagará la compra.
	const assetCurrency = $derived(asset.currency?.trim().toUpperCase() || undefined);
</script>

<div class="asset-preview">
	<div class="preview-item">
		<span class="preview-label">Ticker</span>
		<span class="preview-value">{asset.ticker}</span>
	</div>
	<div class="preview-item">
		<span class="preview-label">Nombre</span>
		<span class="preview-value">{asset.name}</span>
	</div>
	<div class="preview-item">
		<span class="preview-label">Tipo</span>
		<span class="preview-value">{asset.assetType}</span>
	</div>
	<div class="preview-item">
		<span class="preview-label">Exchange</span>
		<span class="preview-value">{asset.exchange}</span>
	</div>
	{#if asset.currentPrice}
		<div class="preview-item">
			<span class="preview-label">Precio de mercado</span>
			<span class="preview-value"
				>{formatCurrency(parseFloat(asset.currentPrice.value), assetCurrency)}</span
			>
		</div>
	{/if}
</div>

<style>
	.asset-preview {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		margin-top: 1rem;
		padding: 1rem;
		border-radius: 12px;
		background: var(--surface);
		border: 1px solid var(--border-strong);
	}

	.preview-item {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.preview-label {
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: rgba(236, 234, 229, 0.5);
		font-weight: 600;
	}

	.preview-value {
		font-size: 0.95rem;
		color: var(--amber);
		font-weight: 600;
	}

	@media (max-width: 1024px) {
		.asset-preview {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.asset-preview {
			grid-template-columns: 1fr;
		}
	}
</style>
