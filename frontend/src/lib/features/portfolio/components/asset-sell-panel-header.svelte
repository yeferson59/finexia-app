<script lang="ts">
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { Transaction } from '$lib/api/types';

	/**
	 * Qué lote se está vendiendo. El título y el botón de cerrar los pone ahora
	 * el `Modal` que aloja el panel: aquí duplicaban los suyos.
	 */
	let {
		transaction,
		formatCurrency
	}: {
		transaction: Transaction;
		formatCurrency: (value: number, decimals?: number) => string;
	} = $props();
</script>

<p class="sell-panel-header">
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
</p>

<style>
	.sell-panel-header {
		margin: 0;
	}

	.sell-panel-lot {
		font-size: 0.82rem;
		color: rgba(236, 234, 229, 0.6);
		font-family: var(--font-mono);
	}
</style>
