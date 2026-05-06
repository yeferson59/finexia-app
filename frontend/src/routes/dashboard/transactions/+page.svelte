<script lang="ts">
	const txs = [
		{ id: 'TRX-9218', type: 'Compra', asset: 'AAPL', amount: '$5,000', status: 'Completada' },
		{ id: 'TRX-9212', type: 'Depósito', asset: 'Cuenta principal', amount: '$10,000', status: 'Completada' },
		{ id: 'TRX-9184', type: 'Venta', asset: 'ETF Global', amount: '$2,700', status: 'Pendiente' },
		{ id: 'TRX-9157', type: 'Transferencia', asset: 'Cuenta ahorro', amount: '$2,000', status: 'Completada' }
	];
</script>

<svelte:head>
	<title>Transacciones - FINEXIA</title>
	<meta name="description" content="Historial de movimientos y estados de transacciones" />
</svelte:head>

<header class="page-header">
	<h1 class="page-title">Transacciones</h1>
	<p class="page-subtitle">Monitorea en tiempo real todos los movimientos de tu cuenta.</p>
</header>

<section class="panel">
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
</section>

<style>
	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
		font-weight: 700;
		color: #d4af37;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.page-subtitle {
		margin: 0;
		color: rgba(224, 224, 224, 0.62);
	}

	.panel {
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		padding: 1.35rem;
	}

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
		background: rgba(15, 20, 25, 0.42);
	}

	.row.heading {
		background: rgba(212, 175, 55, 0.12);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		color: rgba(224, 224, 224, 0.75);
	}

	.row span {
		font-size: 0.84rem;
		color: #e0e0e0;
	}

	.status {
		font-weight: 700;
	}

	.status.ok {
		color: #2ecc71;
	}

	.status.pending {
		color: #e8c547;
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}

		.row {
			grid-template-columns: 1fr;
		}
	}
</style>
