<script lang="ts">
	/*
	 * Los últimos movimientos, como un extracto.
	 *
	 * Cada fila llevaba un cuadro de color con un icono —ámbar para compras,
	 * verde para dividendos, azul para transferencias—: tres paletas más en una
	 * página que ya usaba el verde para las ganancias, y ninguna decía nada que
	 * no dijera el propio título de la fila. El signo del importe basta para
	 * saber si el dinero entró o salió.
	 */
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { formatCurrency } from '$lib/shared/format/money';

	interface Transaction {
		id: string;
		type: string;
		quantity: string;
		price: string;
		currency: string;
		transactionDate: string;
		assetTicker: string;
		assetName: string;
	}

	const { transactions = [] }: { transactions: Transaction[] } = $props();

	const LABELS: Record<string, string> = {
		buy: 'Compra',
		sell: 'Venta',
		dividend: 'Dividendo',
		interest: 'Interés',
		transfer_in: 'Transferencia recibida',
		transfer_out: 'Transferencia enviada',
		split: 'División de acciones',
		fee: 'Comisión'
	};

	/** Movimientos que meten dinero en la cartera; el resto lo sacan. */
	const INCOMING = ['dividend', 'interest', 'transfer_in', 'split'];

	function total(tx: Transaction): number {
		return (parseFloat(tx.quantity) || 0) * (parseFloat(tx.price) || 0);
	}

	/** «Hoy» y «Ayer» se leen antes que una fecha, y son los dos casos frecuentes. */
	function when(iso: string): string {
		const [year, month, day] = iso.split('T')[0].split('-').map(Number);
		const now = new Date();
		const days = Math.round(
			(Date.UTC(year, month - 1, day) -
				Date.UTC(now.getFullYear(), now.getMonth(), now.getDate())) /
				86400000
		);

		if (days === 0) return 'Hoy';
		if (days === -1) return 'Ayer';
		return formatCalendarDate(iso, { day: 'numeric', month: 'short' });
	}
</script>

<section class="activity" aria-labelledby="activity-title">
	<header class="head">
		<h2 id="activity-title">Movimientos</h2>
		<a class="all" href={resolve('/dashboard/transactions')}>Ver todos</a>
	</header>

	{#if transactions.length === 0}
		<p class="empty">
			Tus compras, ventas y dividendos aparecerán aquí en cuanto los registres.
			<a href={resolve('/dashboard/transactions/import')}>Importar un CSV</a>
		</p>
	{:else}
		<ul class="list">
			{#each transactions as tx (tx.id)}
				{@const incoming = INCOMING.includes(tx.type)}
				<li class="row">
					<div class="what">
						<p class="kind">{LABELS[tx.type] ?? tx.type}</p>
						<p class="asset">{tx.assetName} ({tx.assetTicker})</p>
					</div>
					<div class="figures">
						<p class="amount" class:incoming>
							{incoming ? '+' : '−'}{privacy.money(
								formatCurrency(Math.abs(total(tx)), tx.currency || 'USD')
							)}
						</p>
						<time class="date" datetime={tx.transactionDate}>{when(tx.transactionDate)}</time>
					</div>
				</li>
			{/each}
		</ul>

		<a class="statement" href={resolve('/dashboard/reports')}>Descargar el extracto</a>
	{/if}
</section>

<style>
	.activity {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.25rem;
	}

	h2 {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	.all,
	.statement {
		font-size: 0.8rem;
		color: var(--text-muted);
		text-decoration: none;
		white-space: nowrap;
	}

	.all:hover,
	.statement:hover {
		color: var(--text);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.statement {
		align-self: flex-start;
		margin-top: 1.25rem;
	}

	.list {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/*
	 * La única superficie levantada de la página, y por eso significa algo: cada
	 * fila es un objeto suelto del extracto, no una sección.
	 */
	.row {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.7rem 0.85rem;
		border-radius: 8px;
		background: var(--panel);
	}

	.what {
		min-width: 0;
	}

	.kind {
		margin: 0;
		font-size: 0.85rem;
		color: var(--text);
	}

	.asset {
		margin: 0.15rem 0 0;
		font-size: 0.75rem;
		color: var(--text-dim);
		overflow-wrap: anywhere;
	}

	.figures {
		flex-shrink: 0;
		text-align: right;
	}

	.amount {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.85rem;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.amount.incoming {
		color: var(--green);
	}

	.date {
		display: block;
		margin-top: 0.15rem;
		font-family: var(--font-mono);
		font-size: 0.7rem;
		color: var(--text-dim);
	}

	.empty {
		margin: 0;
		font-size: 0.85rem;
		line-height: 1.6;
		color: var(--text-dim);
	}

	.empty a {
		color: var(--text-muted);
		text-decoration: underline;
		text-underline-offset: 3px;
	}
</style>
