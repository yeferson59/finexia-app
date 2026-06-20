<script lang="ts">
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';

	const txs = [
		{ id: 'TRX-9218', type: 'Compra', asset: 'AAPL', amount: '$5,000', status: 'Completada' },
		{
			id: 'TRX-9212',
			type: 'Depósito',
			asset: 'Cuenta principal',
			amount: '$10,000',
			status: 'Completada'
		},
		{ id: 'TRX-9184', type: 'Venta', asset: 'ETF Global', amount: '$2,700', status: 'Pendiente' },
		{
			id: 'TRX-9157',
			type: 'Transferencia',
			asset: 'Cuenta ahorro',
			amount: '$2,000',
			status: 'Completada'
		}
	];
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
	<div class="table">
		<div class="row heading">
			<span>ID</span>
			<span>Tipo</span>
			<span>Activo</span>
			<span>Monto</span>
			<span>Estado</span>
		</div>
		{#each txs as tx (tx.id)}
			<div class="row">
				<span>{tx.id}</span>
				<span>{tx.type}</span>
				<span>{tx.asset}</span>
				<span>{tx.amount}</span>
				<span class={`status ${tx.status === 'Completada' ? 'ok' : 'pending'}`}>{tx.status}</span>
			</div>
		{/each}
	</div>
</Card>

<style>
	.table {
		display: grid;
		gap: 0.55rem;
	}

	.row {
		display: grid;
		grid-template-columns: 0.8fr 0.75fr 1.1fr 0.7fr 0.7fr;
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

	.status {
		font-weight: 700;
	}

	.status.ok {
		color: var(--green);
	}

	.status.pending {
		color: var(--amber-light);
	}

	@media (max-width: 768px) {
		.row {
			grid-template-columns: 1fr;
		}
	}
</style>
