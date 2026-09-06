<script lang="ts">
	/*
	 * El libro de movimientos de la cuenta, del más reciente al más antiguo.
	 *
	 * Era una rejilla de `<div>` con una fila por movimiento, cada una con su
	 * fondo y su radio: doce cápsulas idénticas apiladas, y una cabecera ámbar en
	 * versalitas encima. Ahora es una tabla de verdad —la misma que la de
	 * movimientos de un activo—, así que un lector de pantalla nombra cada
	 * columna en vez de leer una tira de cifras.
	 *
	 * La primera columna era el identificador («TRX-66666666»): un UUID cortado
	 * que no se puede buscar, ni copiar a soporte, ni cuadrar con el extracto del
	 * bróker. En su sitio va lo que pasó, y el hueco que ocupaba lo aprovechan la
	 * cantidad y el precio, que son con lo que se comprueba el total.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { formatCurrency as formatMoney } from '$lib/shared/format/money';
	import type { UserTransaction } from '$lib/api/types';
	import { TXN_TYPE_LABELS, transactionTotal } from '../transactions';

	let { transactions }: { transactions: UserTransaction[] } = $props();

	/*
	 * El mismo formateador que el resto del panel: esta pantalla componía el
	 * importe a mano y escribía «USD 2.436,00» donde las demás dicen
	 * «$2,436.00» del mismo número.
	 *
	 * Y con el mismo escape para el precio por unidad: el interés de una cuenta
	 * a 0,0021 por dólar salía «$0.00» en la misma fila que su total de $19.95,
	 * así que la fila no cuadraba.
	 */
	function money(value: number, currency: string): string {
		const code = currency || 'USD';
		const maxDigits = value !== 0 && Math.abs(value) < 0.01 ? 6 : undefined;

		return privacy.money(formatMoney(value, code, maxDigits));
	}

	function fmtDate(iso: string): string {
		return formatCalendarDate(iso, { year: 'numeric', month: 'short', day: 'numeric' });
	}
</script>

<div class="scroll">
	<table>
		<caption class="sr-only">
			Los movimientos registrados en tu cuenta, del más reciente al más antiguo, con su activo,
			fecha, cantidad, precio unitario y total
		</caption>
		<thead>
			<tr>
				<th scope="col" class="col-kind">Movimiento</th>
				<th scope="col" class="col-asset">Activo</th>
				<th scope="col" class="col-date">Fecha</th>
				<th scope="col" class="col-qty num">Cantidad</th>
				<th scope="col" class="col-price num">Precio</th>
				<th scope="col" class="col-total num">Total</th>
			</tr>
		</thead>
		<tbody>
			{#each transactions as txn (txn.id)}
				{@const total = transactionTotal(txn)}
				<tr>
					<th scope="row" class="col-kind kind">
						{TXN_TYPE_LABELS[txn.type] ?? txn.type}
						<!-- La nota, debajo del tipo y entera. No se veía en ninguna parte
						     de esta pantalla, y es lo único que escribió el usuario. -->
						{#if txn.notes}
							<span class="note">{txn.notes}</span>
						{/if}
					</th>

					<td class="col-asset asset">
						{txn.assetName}
						<span class="ticker">{txn.assetTicker}</span>
					</td>

					<td class="col-date date">{fmtDate(txn.transactionDate)}</td>

					<td class="col-qty num figure">
						{(parseFloat(txn.quantity) || 0).toLocaleString('es-CO', { maximumFractionDigits: 8 })}
					</td>

					<td class="col-price num figure muted">
						{money(parseFloat(txn.price) || 0, txn.currency)}
					</td>

					<td class="col-total num figure total">
						{money(total.amount, total.currency)}
						{#if total.converted}
							<span class="fx">{total.quoteCurrency} × {total.rate}</span>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<style>
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
		min-width: 48rem;
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

	.kind {
		font-weight: 500;
		white-space: nowrap;
	}

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

	/* El nombre manda y el ticker va debajo: «Vanguard FTSE All-World UCITS ETF
	   (VWCE)» en una línea empujaba las cifras fuera de la pantalla. */
	.asset {
		min-width: 0;
	}

	.ticker {
		display: block;
		margin-top: 0.15rem;
		font-family: var(--font-mono);
		font-size: 0.72rem;
		letter-spacing: 0.04em;
		color: var(--text-dim);
	}

	.date {
		color: var(--text-muted);
		white-space: nowrap;
	}

	.num {
		text-align: right;
	}

	.figure {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.muted {
		color: var(--text-muted);
	}

	.total {
		font-weight: 600;
	}

	.fx {
		display: block;
		margin-top: 0.2rem;
		font-size: 0.72rem;
		font-weight: 400;
		color: var(--text-dim);
	}

	/* El precio unitario es lo primero que sobra cuando aprieta: el total lo
	   resume y la cantidad sigue estando. Vuelve al plegarse la fila. */
	@media (max-width: 1000px) {
		table {
			min-width: 38rem;
		}

		.col-price {
			display: none;
		}
	}

	/*
	 * Debajo de esto la fila se pliega, como en el listado de portafolios y en la
	 * tabla de movimientos de un activo: qué pasó y por cuánto arriba, el detalle
	 * debajo. Antes la rejilla pasaba a una sola columna sin etiquetas, así que
	 * quedaban cinco valores apilados sin decir cuál era cuál.
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

		.col-asset {
			grid-column: 1;
			grid-row: 2;
			margin-top: 0.45rem;
		}

		.col-date {
			grid-column: 1;
			grid-row: 3;
			margin-top: 0.35rem;
			font-size: 0.8rem;
		}

		.col-qty {
			grid-column: 2;
			grid-row: 2;
			margin-top: 0.45rem;
		}

		.col-price {
			display: table-cell;
			grid-column: 2;
			grid-row: 3;
			margin-top: 0.35rem;
		}

		.col-qty::before,
		.col-price::before {
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
	}
</style>
