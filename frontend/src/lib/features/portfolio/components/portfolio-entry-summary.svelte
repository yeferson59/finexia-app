<script lang="ts">
	import type { Asset } from '$lib/api/types';

	let {
		asset,
		quantity,
		purchasePrice,
		totalValue,
		formatCurrency
	}: {
		asset: Asset;
		quantity: string;
		purchasePrice: string;
		totalValue: number;
		formatCurrency: (value: number) => string;
	} = $props();
</script>

<section class="summary-card">
	<h3 class="summary-title">Resumen de Inversión</h3>
	<div class="summary-items">
		<div class="summary-item">
			<span class="summary-label">Activo</span>
			<span class="summary-value">{asset.ticker} - {asset.name}</span>
		</div>
		<div class="summary-item">
			<span class="summary-label">Cantidad</span>
			<span class="summary-value">{parseFloat(quantity).toLocaleString()}</span>
		</div>
		<div class="summary-item">
			<span class="summary-label">Precio Unitario</span>
			<span class="summary-value">{formatCurrency(parseFloat(purchasePrice))}</span>
		</div>
		<div class="summary-item border-top">
			<span class="summary-label">Inversión Total</span>
			<span class="summary-value highlight">{formatCurrency(totalValue)}</span>
		</div>
	</div>
</section>

<style>
	.summary-card {
		border: 2px solid rgba(212, 145, 42, 0.3);
		border-radius: 16px;
		background: rgba(212, 145, 42, 0.08);
		padding: 1.5rem;
		animation: slide-in 0.4s ease-out;
	}

	.summary-title {
		margin: 0 0 1.25rem;
		font-size: 1rem;
		font-weight: 700;
		color: var(--amber);
		font-family: var(--font-body);
	}

	.summary-items {
		display: grid;
		gap: 0.9rem;
	}

	.summary-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 0;
	}

	.summary-item.border-top {
		border-top: 1px solid rgba(212, 145, 42, 0.2);
		padding-top: 1rem;
		margin-top: 0.5rem;
	}

	.summary-label {
		font-size: 0.9rem;
		color: rgba(236, 234, 229, 0.65);
		font-weight: 500;
	}

	.summary-value {
		font-size: 0.95rem;
		color: var(--text);
		font-weight: 600;
	}

	.summary-value.highlight {
		color: var(--amber);
		font-size: 1.1rem;
	}

	@keyframes slide-in {
		from {
			opacity: 0;
			transform: translateX(-10px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}
</style>
