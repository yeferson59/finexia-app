<script lang="ts">
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const typeLabels: Record<string, string> = {
		buy: 'Compra',
		sell: 'Venta',
		dividend: 'Dividendo',
		interest: 'Interés',
		transfer_in: 'Transferencia Entrada',
		transfer_out: 'Transferencia Salida',
		split: 'División',
		fee: 'Cargo'
	};

	function formatType(type: string): string {
		return typeLabels[type] ?? type;
	}

	function formatAmount(quantity: string, price: string, currency: string): string {
		const total = (parseFloat(quantity) || 0) * (parseFloat(price) || 0);
		return `${currency} ${new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		}).format(total)}`;
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('es-CO', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function shortId(id: string): string {
		return `TRX-${id.slice(0, 8).toUpperCase()}`;
	}
</script>

<svelte:head>
	<title>Transacciones - FINEXIA</title>
	<meta name="description" content="Historial de movimientos y estados de transacciones" />
</svelte:head>

<PageHeader
	title="Transacciones"
	subtitle="Monitorea en tiempo real todos los movimientos de tu cuenta."
/>

<Card variant="elevated" padding="sm">
	{#if data.transactions.length === 0}
		<p class="empty-state">No hay transacciones registradas.</p>
	{:else}
		<div class="table">
			<div class="row heading">
				<span>ID</span>
				<span>Tipo</span>
				<span>Activo</span>
				<span>Monto</span>
				<span>Fecha</span>
			</div>
			{#each data.transactions as tx (tx.id)}
				<div class="row">
					<span class="mono">{shortId(tx.id)}</span>
					<span>{formatType(tx.type)}</span>
					<span>{tx.assetName} ({tx.assetTicker})</span>
					<span class="mono">{formatAmount(tx.quantity, tx.price, tx.currency)}</span>
					<span>{formatDate(tx.transactionDate)}</span>
				</div>
			{/each}
		</div>
	{/if}
</Card>

<style>
	.empty-state {
		font-size: 0.875rem;
		color: var(--text-muted);
		text-align: center;
		padding: 2rem;
	}

	.table {
		display: grid;
		gap: 0.55rem;
	}

	.row {
		display: grid;
		grid-template-columns: 0.8fr 0.9fr 1.4fr 0.9fr 0.7fr;
		gap: 0.7rem;
		padding: 0.85rem;
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
	}

	.row.heading {
		background: rgba(212, 145, 42, 0.12);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.75);
	}

	.row span {
		font-size: 0.84rem;
		color: var(--text);
	}

	.mono {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	@media (max-width: 768px) {
		.row {
			grid-template-columns: 1fr;
		}
	}
</style>
