<script lang="ts">
	import { formatCalendarDate } from '$lib/utils';
	import type { Transaction } from '$lib/api/types';

	let {
		transaction,
		formatCurrency,
		onClose
	}: {
		transaction: Transaction;
		formatCurrency: (value: number, decimals?: number) => string;
		onClose: () => void;
	} = $props();
</script>

<div class="sell-panel-header">
	<div class="sell-panel-info">
		<span class="sell-panel-title">Vender desde compra</span>
		<span class="sell-panel-lot">
			Lote: {parseFloat(transaction.quantity).toLocaleString('es-CO', {
				maximumFractionDigits: 8
			})}
			unidades @ {formatCurrency(parseFloat(transaction.price))} ·
			{formatCalendarDate(transaction.transactionDate, {
				year: 'numeric',
				month: 'short',
				day: 'numeric'
			})}
		</span>
	</div>
	<button class="sell-panel-close" type="button" onclick={onClose}>✕</button>
</div>

<style>
	.sell-panel-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.sell-panel-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.sell-panel-title {
		font-size: 0.9rem;
		font-weight: 700;
		color: var(--red);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.sell-panel-lot {
		font-size: 0.82rem;
		color: rgba(236, 234, 229, 0.6);
		font-family: var(--font-mono);
	}

	.sell-panel-close {
		padding: 0.2rem 0.5rem;
		border: none;
		background: transparent;
		color: rgba(236, 234, 229, 0.4);
		font-size: 1rem;
		cursor: pointer;
		border-radius: 4px;
		transition: color 0.2s ease;
		flex-shrink: 0;
	}

	.sell-panel-close:hover {
		color: var(--text);
	}
</style>
