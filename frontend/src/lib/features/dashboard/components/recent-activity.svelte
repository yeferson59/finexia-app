<script lang="ts">
	/*
	 * Los últimos movimientos, como un extracto.
	 *
	 * Cada fila llevaba un cuadro de color con un icono —ámbar para compras,
	 * verde para dividendos, azul para transferencias—: tres paletas más en una
	 * página que ya usaba el verde para las ganancias, y ninguna decía nada que
	 * no dijera el propio título de la fila. El signo del importe basta para
	 * saber si el dinero entró o salió.
	 *
	 * Después cada fila fue una cajita propia, y cinco cajitas en columna se
	 * leían como cinco avisos. Un extracto es una lista corrida bajo la fecha:
	 * los movimientos del mismo día van juntos y el día se escribe una vez.
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
		cash_interest: 'Intereses del efectivo',
		transfer_in: 'Transferencia recibida',
		transfer_out: 'Transferencia enviada',
		split: 'División de acciones',
		fee: 'Comisión'
	};

	/** Movimientos que meten dinero en la cartera; el resto lo sacan. */
	const INCOMING = ['dividend', 'interest', 'cash_interest', 'transfer_in', 'split'];

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

	/** Los movimientos en el orden en que llegan, bajo el día en que ocurrieron. */
	const days = $derived(
		transactions.reduce<{ label: string; items: Transaction[] }[]>((acc, tx) => {
			const label = when(tx.transactionDate);
			const last = acc.at(-1);
			if (last?.label === label) last.items.push(tx);
			else acc.push({ label, items: [tx] });
			return acc;
		}, [])
	);
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
		<ol class="days">
			{#each days as day (day.label)}
				<li>
					<h3 class="day">
						<time datetime={day.items[0].transactionDate.split('T')[0]}>{day.label}</time>
					</h3>
					<ul class="list">
						{#each day.items as tx (tx.id)}
							{@const incoming = INCOMING.includes(tx.type)}
							<li class="row">
								<div class="what">
									<p class="kind">{LABELS[tx.type] ?? tx.type}</p>
									<p class="asset">{tx.assetName} ({tx.assetTicker})</p>
								</div>
								<p class="amount" class:incoming>
									{incoming ? '+' : '−'}{privacy.money(
										formatCurrency(Math.abs(total(tx)), tx.currency || 'USD')
									)}
								</p>
							</li>
						{/each}
					</ul>
				</li>
			{/each}
		</ol>

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
		margin-bottom: 1rem;
	}

	h2 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.3rem;
		font-weight: 400;
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

	.days,
	.list {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/* El día abre su grupo con un filete: es el único corte que tiene el
	   extracto, y lo que separa un grupo del siguiente. */
	.day {
		margin: 0;
		padding: 0.9rem 0 0.35rem;
		border-top: 1px solid var(--border);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
	}

	.row {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.45rem 0;
	}

	.what {
		min-width: 0;
	}

	.kind {
		margin: 0;
		font-size: 0.88rem;
		color: var(--text);
	}

	.asset {
		margin: 0.1rem 0 0;
		font-size: 0.75rem;
		color: var(--text-dim);
		overflow-wrap: anywhere;
	}

	.amount {
		flex-shrink: 0;
		margin: 0;
		font-family: var(--font-figures);
		font-size: 0.92rem;
		font-stretch: 88%;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.amount.incoming {
		color: var(--green);
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
