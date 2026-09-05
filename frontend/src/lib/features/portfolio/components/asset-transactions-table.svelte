<script lang="ts">
	/*
	 * Lo que has hecho con este activo, del movimiento más reciente al más
	 * antiguo.
	 *
	 * Era una rejilla de `<div>` con párrafos dentro: quien usa lector de
	 * pantalla oía una cadena de cifras sin saber qué columna era cada una.
	 * Ahora es una tabla de verdad, con cabeceras que nombran las columnas y el
	 * tipo de movimiento como cabecera de fila.
	 *
	 * Las insignias de colores por tipo —ocho, una por cada clase de
	 * movimiento— se han quedado en la palabra: «Compra» ya dice lo que es, y
	 * ocho cápsulas de color en la misma tabla no dejaban destacar a ninguna.
	 * La nota, que se cortaba con puntos suspensivos en su propia columna, va
	 * bajo el tipo y se lee entera.
	 */
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { Transaction } from '$lib/api/types';
	import { TYPE_LABEL, type TxnMeta } from '../asset';
	import AssetTransactionActions from './asset-transaction-actions.svelte';

	let {
		transactions,
		txnMeta,
		sellingTxnId = null,
		formatAmount,
		onEdit,
		onToggleSell,
		onDelete
	}: {
		transactions: Transaction[];
		txnMeta: TxnMeta;
		sellingTxnId?: string | null;
		/**
		 * Cada fila se formatea con la moneda de su propia transacción: una
		 * compra liquidada en EUR no es la misma cifra con el símbolo de la
		 * moneda base delante.
		 */
		formatAmount: (value: number, currency: string) => string;
		onEdit: (txn: Transaction) => void;
		onToggleSell: (txn: Transaction) => void;
		onDelete: (txn: Transaction) => void;
	} = $props();

	const currentPage = $derived(txnMeta.page);
	const totalPages = $derived(txnMeta.totalPages);

	function fmtDate(iso: string): string {
		return formatCalendarDate(iso, { year: 'numeric', month: 'short', day: 'numeric' });
	}
</script>

{#if transactions.length === 0}
	<p class="empty">
		Aún no has registrado ningún movimiento de este activo. Con «Registrar movimiento» anotas una
		compra, una venta o un dividendo, y la posición se recalcula sola.
	</p>
{:else}
	<div class="scroll">
		<table>
			<caption class="sr-only">
				Los movimientos registrados de este activo, del más reciente al más antiguo, con su fecha,
				cantidad, precio unitario, comisión y total
			</caption>
			<thead>
				<tr>
					<th scope="col" class="col-kind">Movimiento</th>
					<th scope="col" class="col-date">Fecha</th>
					<th scope="col" class="col-qty num">Cantidad</th>
					<th scope="col" class="col-price num">Precio</th>
					<th scope="col" class="col-fees num">Comisión</th>
					<th scope="col" class="col-total num">Total</th>
					<th scope="col" class="col-actions"><span class="sr-only">Acciones</span></th>
				</tr>
			</thead>
			<tbody>
				{#each transactions as transaction (transaction.id)}
					{@const qty = parseFloat(transaction.quantity) || 0}
					{@const price = parseFloat(transaction.price) || 0}
					{@const fees = parseFloat(transaction.fees) || 0}
					{@const rate = parseFloat(transaction.fxRate ?? '1') || 1}
					{@const costCurrency = transaction.costCurrency || transaction.currency}
					{@const converted = costCurrency !== transaction.currency}
					{@const feesCurrency = transaction.feesCurrency || transaction.currency}
					<!-- El total es lo que la cuenta pagó o recibió, no lo que la
					     operación cotizó: precio y comisión se quedan en la moneda del
					     mercado, y la tasa los lleva a la de la cuenta. -->
					{@const total = qty * price * rate}
					{@const isActiveSell = sellingTxnId === transaction.id}
					<tr class:selling={isActiveSell}>
						<th scope="row" class="col-kind kind">
							{TYPE_LABEL[transaction.type] ?? transaction.type}
							{#if transaction.notes}
								<span class="note">{transaction.notes}</span>
							{/if}
						</th>

						<td class="col-date date">{fmtDate(transaction.transactionDate)}</td>
						<td class="col-qty num qty">
							{qty.toLocaleString('es-CO', { maximumFractionDigits: 8 })}
						</td>
						<td class="col-price num figure">{formatAmount(price, transaction.currency)}</td>
						<!-- La comisión en la moneda en la que se cobró, que no siempre es la
						     de la ejecución: un bróker que llenó en euros pudo cargarla en
						     dólares, y etiquetarla con la del precio la mueve por la tasa. -->
						<td class="col-fees num figure muted" class:no-fee={fees <= 0}>
							{fees > 0 ? formatAmount(fees, feesCurrency) : '—'}
						</td>
						<td class="col-total num figure total">
							{formatAmount(total, costCurrency)}
							{#if converted}
								<span class="fx-note">{transaction.currency} × {rate}</span>
							{/if}
						</td>

						<td class="col-actions">
							<AssetTransactionActions
								{transaction}
								selling={isActiveSell}
								{onEdit}
								{onToggleSell}
								{onDelete}
							/>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	{#if totalPages > 1}
		<form class="pagination" method="GET">
			<input type="hidden" name="limit" value={txnMeta.limit} />
			<span class="pg-status">Página {currentPage} de {totalPages}</span>
			<button
				type="submit"
				name="page"
				value={currentPage - 1}
				class="pg"
				disabled={currentPage === 1}
			>
				Anteriores
			</button>
			<button
				type="submit"
				name="page"
				value={currentPage + 1}
				class="pg"
				disabled={currentPage === totalPages}
			>
				Siguientes
			</button>
		</form>
	{/if}
{/if}

<style>
	.empty {
		max-width: 60ch;
		margin: 1.5rem 0 0;
		font-size: 0.88rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	.scroll {
		overflow-x: auto;
	}

	table {
		width: 100%;
		min-width: 46rem;
		margin-top: 1.35rem;
		border-collapse: collapse;
	}

	thead th {
		padding: 0 0.75rem 0.6rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
		text-align: left;
		white-space: nowrap;
	}

	thead th.num {
		text-align: right;
	}

	thead th:first-child {
		padding-left: 0;
	}

	thead th:last-child {
		padding-right: 0;
	}

	tbody th,
	tbody td {
		padding: 0.85rem 0.75rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.85rem;
		font-weight: 400;
		color: var(--text);
		text-align: left;
		vertical-align: top;
	}

	tbody th:first-child {
		padding-left: 0;
	}

	tbody td:last-child {
		padding-right: 0;
	}

	tbody tr:last-child th,
	tbody tr:last-child td {
		border-bottom: none;
	}

	@media (hover: hover) {
		tbody tr:hover th,
		tbody tr:hover td {
			background: var(--panel);
		}
	}

	/* El lote que se está vendiendo, señalado con el filete rojo de la venta:
	   el diálogo tapa parte de la tabla y al cerrarlo hay que saber cuál era. */
	.selling th:first-child {
		box-shadow: inset 2px 0 0 var(--red);
	}

	.kind {
		font-weight: 500;
		white-space: nowrap;
	}

	/* La nota completa, debajo del tipo. Tenía columna propia y se cortaba en
	   «Dividendo trim…», que es justo la parte que no se puede adivinar. */
	.note {
		display: block;
		max-width: 26ch;
		margin-top: 0.2rem;
		font-size: 0.78rem;
		font-weight: 400;
		line-height: 1.35;
		color: var(--text-muted);
		white-space: normal;
	}

	.date {
		color: var(--text-muted);
		white-space: nowrap;
	}

	.num {
		text-align: right;
	}

	.qty,
	.figure {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.muted {
		color: var(--text-dim);
	}

	.total {
		font-weight: 600;
	}

	.fx-note {
		display: block;
		margin-top: 0.2rem;
		font-size: 0.72rem;
		font-weight: 400;
		color: var(--text-dim);
	}

	.pagination {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 0.6rem;
		padding-top: 1.25rem;
	}

	.pg-status {
		margin-right: auto;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-dim);
	}

	.pg {
		padding: 0.4rem 0.9rem;
		border: 1px solid var(--border-strong);
		border-radius: 8px;
		background: none;
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.82rem;
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			background 0.2s ease;
	}

	.pg:hover:not(:disabled) {
		border-color: var(--text-dim);
		background: var(--panel);
	}

	.pg:disabled {
		color: var(--text-dim);
		cursor: default;
	}

	@media (prefers-reduced-motion: reduce) {
		.pg {
			transition: none;
		}
	}

	/* La comisión es la columna que menos se consulta y la primera que sobra
	   cuando la tabla empieza a apretar; su importe vuelve más abajo, cuando la
	   fila se pliega y hay sitio de sobra. */
	@media (max-width: 900px) {
		table {
			min-width: 38rem;
		}

		.col-fees {
			display: none;
		}
	}

	/*
	 * Debajo de esto la fila se pliega, como en el listado de portafolios y en
	 * el de activos: qué pasó y por cuánto arriba, el detalle debajo. Sin esto
	 * la tabla se iba en scroll horizontal y el total —la cifra por la que se
	 * mira una fila— quedaba fuera de la pantalla.
	 *
	 * Las cabeceras siguen en el árbol para el lector de pantalla; las
	 * etiquetas de aquí son decorativas y solo reponen a la vista el nombre de
	 * la columna que se acaba de esconder.
	 */
	@media (max-width: 760px) {
		.scroll {
			overflow-x: visible;
		}

		table {
			min-width: 0;
		}

		thead {
			display: none;
		}

		tbody tr {
			display: grid;
			grid-template-columns: minmax(0, 1fr) auto;
			column-gap: 1rem;
			padding: 0.9rem 0;
			border-bottom: 1px solid var(--border);
		}

		tbody tr:last-child {
			border-bottom: none;
		}

		tbody th,
		tbody td {
			width: auto;
			padding: 0;
			border: none;
		}

		.col-kind {
			grid-column: 1;
			grid-row: 1;
		}

		.col-total {
			grid-column: 2;
			grid-row: 1;
		}

		.col-date {
			grid-column: 1;
			grid-row: 2;
			align-self: end;
			margin-top: 0.5rem;
		}

		.col-qty {
			grid-column: 2;
			grid-row: 2;
			margin-top: 0.5rem;
		}

		.col-price {
			grid-column: 2;
			grid-row: 3;
			margin-top: 0.2rem;
		}

		.col-fees {
			display: table-cell;
			grid-column: 2;
			grid-row: 4;
			margin-top: 0.2rem;
		}

		.col-actions {
			grid-column: 1 / -1;
			grid-row: 5;
			margin-top: 0.75rem;
		}

		.col-qty::before,
		.col-price::before,
		.col-fees::before {
			margin-right: 0.5rem;
			font-family: var(--font-body);
			font-size: 0.78rem;
			color: var(--text-dim);
		}

		.col-qty::before {
			content: 'Cantidad';
		}

		.col-price::before {
			content: 'Precio';
		}

		.col-fees::before {
			content: 'Comisión';
		}

		/* Plegada, una comisión de cero es una línea que solo dice «—»: en la
		   tabla hace falta para que cuadre la columna, aquí no. */
		.col-fees.no-fee {
			display: none;
		}

		.selling th:first-child {
			box-shadow: none;
		}

		.selling {
			box-shadow: inset 2px 0 0 var(--red);
			padding-left: 0.75rem;
		}
	}
</style>
