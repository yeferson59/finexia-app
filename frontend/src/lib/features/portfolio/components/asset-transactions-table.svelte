<script lang="ts">
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { Transaction } from '$lib/api/types';
	import { TYPE_LABEL, TYPE_STYLE, type TxnMeta } from '../asset';

	let {
		transactions,
		txnMeta,
		sellingTxnId = null,
		formatCurrency,
		onEdit,
		onToggleSell
	}: {
		transactions: Transaction[];
		txnMeta: TxnMeta;
		sellingTxnId?: string | null;
		formatCurrency: (value: number, decimals?: number) => string;
		onEdit: (txn: Transaction) => void;
		onToggleSell: (txn: Transaction) => void;
	} = $props();

	const currentPage = $derived(txnMeta.page);
	const totalPages = $derived(txnMeta.totalPages);

	function fmtDate(iso: string): string {
		return formatCalendarDate(iso, { year: 'numeric', month: 'short', day: 'numeric' });
	}
</script>

{#if transactions.length === 0}
	<p class="empty-txn">No hay transacciones registradas aún.</p>
{:else}
	<div class="transactions-table">
		<div class="table-header">
			<p>Tipo</p>
			<p>Fecha</p>
			<p>Cantidad</p>
			<p>Precio</p>
			<p>Comisión</p>
			<p>Total</p>
			<p>Notas</p>
			<p>Acciones</p>
		</div>

		{#each transactions as txn (txn.id)}
			{@const qty = parseFloat(txn.quantity) || 0}
			{@const price = parseFloat(txn.price) || 0}
			{@const fees = parseFloat(txn.fees) || 0}
			{@const total = qty * price}
			{@const isBuyLot = txn.type === 'buy' || txn.type === 'transfer_in'}
			{@const isActiveSell = sellingTxnId === txn.id}
			<div class="table-row" class:row-selling={isActiveSell}>
				<p>
					<span class="type-badge {TYPE_STYLE[txn.type] ?? ''}">
						{TYPE_LABEL[txn.type] ?? txn.type}
					</span>
				</p>
				<p class="date">{fmtDate(txn.transactionDate)}</p>
				<p class="qty">{qty.toLocaleString('es-CO', { maximumFractionDigits: 8 })}</p>
				<p class="price">{formatCurrency(price)}</p>
				<p class="fees">{fees > 0 ? formatCurrency(fees) : '—'}</p>
				<p class="total">{formatCurrency(total, 0)}</p>
				<p class="notes">{txn.notes || '—'}</p>
				<p class="cell-action">
					<button
						type="button"
						class="btn-edit-row"
						onclick={() => onEdit(txn)}
						aria-label="Editar transacción"
					>
						<svg
							width="13"
							height="13"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2.5"
						>
							<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
							<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
						</svg>
					</button>
					{#if isBuyLot}
						<button
							type="button"
							class="btn-sell-row"
							class:active={isActiveSell}
							onclick={() => onToggleSell(txn)}
						>
							{isActiveSell ? 'Cancelar' : 'Vender'}
						</button>
					{/if}
				</p>
			</div>
		{/each}
	</div>

	{#if totalPages > 1}
		<form class="pagination" method="GET">
			<input type="hidden" name="limit" value={txnMeta.limit} />
			<button
				type="submit"
				name="page"
				value={currentPage - 1}
				class="pg-btn"
				disabled={currentPage === 1}
			>
				‹ Anterior
			</button>
			{#each Array.from({ length: totalPages }, (_, i) => i + 1) as p (p)}
				<button
					type="submit"
					name="page"
					value={p}
					class="pg-btn pg-num"
					class:pg-active={p === currentPage}>{p}</button
				>
			{/each}
			<button
				type="submit"
				name="page"
				value={currentPage + 1}
				class="pg-btn"
				disabled={currentPage === totalPages}
			>
				Siguiente ›
			</button>
		</form>
	{/if}
{/if}

<style>
	.empty-txn {
		margin: 0;
		padding: 1.5rem;
		text-align: center;
		color: rgba(236, 234, 229, 0.4);
		font-size: 0.9rem;
	}

	.transactions-table {
		overflow-x: auto;
	}

	.table-header {
		display: grid;
		grid-template-columns: 110px 100px 1fr 1fr 1fr 1fr 1fr 120px;
		gap: 1rem;
		padding: 0.75rem 1rem;
		background: rgba(0, 0, 0, 0.2);
		border-radius: 8px 8px 0 0;
		border-bottom: 1px solid var(--border-strong);
		font-weight: 600;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.table-row {
		display: grid;
		grid-template-columns: 110px 100px 1fr 1fr 1fr 1fr 1fr 120px;
		gap: 1rem;
		padding: 1rem;
		border-bottom: 1px solid var(--border);
		align-items: center;
		transition: background 0.2s ease;
	}

	.row-selling {
		background: rgba(224, 90, 90, 0.05) !important;
		border-left: 2px solid rgba(224, 90, 90, 0.4);
	}

	.cell-action {
		display: flex;
		justify-content: flex-end;
	}

	.btn-sell-row {
		padding: 0.3rem 0.65rem;
		border: 1.5px solid rgba(224, 90, 90, 0.4);
		border-radius: 6px;
		background: transparent;
		color: var(--red);
		font-size: 0.78rem;
		font-weight: 600;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
		white-space: nowrap;
	}

	.btn-sell-row:hover,
	.btn-sell-row.active {
		background: rgba(224, 90, 90, 0.12);
		border-color: var(--red);
	}

	.table-row:hover {
		background: rgba(212, 145, 42, 0.03);
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.table-row p {
		margin: 0;
		font-size: 0.9rem;
	}

	.type-badge {
		display: inline-block;
		padding: 0.2rem 0.6rem;
		border-radius: 20px;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.2px;
	}

	.type-buy {
		background: rgba(80, 200, 120, 0.15);
		color: var(--green);
		border: 1px solid rgba(80, 200, 120, 0.3);
	}

	.type-sell {
		background: rgba(224, 90, 90, 0.15);
		color: var(--red);
		border: 1px solid rgba(224, 90, 90, 0.3);
	}

	.type-dividend,
	.type-interest {
		background: rgba(212, 145, 42, 0.15);
		color: var(--amber);
		border: 1px solid rgba(212, 145, 42, 0.3);
	}

	.type-fee {
		background: rgba(150, 150, 150, 0.15);
		color: rgba(236, 234, 229, 0.5);
		border: 1px solid rgba(150, 150, 150, 0.2);
	}

	.type-transfer,
	.type-split {
		background: rgba(100, 160, 230, 0.15);
		color: #7ab4f0;
		border: 1px solid rgba(100, 160, 230, 0.3);
	}

	.date {
		color: rgba(236, 234, 229, 0.5);
	}

	.qty {
		color: var(--text);
		font-weight: 500;
	}

	.price {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--amber);
		font-weight: 500;
	}

	.fees {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: rgba(236, 234, 229, 0.4);
		font-size: 0.85rem;
	}

	.total {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--text);
		font-weight: 600;
	}

	.notes {
		color: rgba(236, 234, 229, 0.4);
		font-size: 0.85rem;
		font-style: italic;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.btn-edit-row {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.28rem 0.5rem;
		border: 1.5px solid rgba(212, 145, 42, 0.3);
		border-radius: 6px;
		background: transparent;
		color: rgba(212, 145, 42, 0.6);
		cursor: pointer;
		transition: all 0.2s ease;
		flex-shrink: 0;
	}

	.btn-edit-row:hover {
		background: rgba(212, 145, 42, 0.1);
		border-color: var(--amber);
		color: var(--amber);
	}

	.pagination {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.35rem;
		padding: 1rem 0 0.25rem;
		flex-wrap: wrap;
	}

	.pg-btn {
		padding: 0.35rem 0.75rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 6px;
		background: transparent;
		color: rgba(236, 234, 229, 0.6);
		font-size: 0.82rem;
		font-weight: 600;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.pg-btn:hover:not(:disabled) {
		border-color: var(--amber);
		color: var(--amber);
		background: rgba(212, 145, 42, 0.08);
	}

	.pg-btn:disabled {
		opacity: 0.3;
		cursor: default;
	}

	.pg-num {
		min-width: 2rem;
	}

	.pg-active {
		background: rgba(212, 145, 42, 0.15);
		border-color: var(--amber);
		color: var(--amber);
	}

	@media (max-width: 768px) {
		.table-header,
		.table-row {
			grid-template-columns: 90px 90px 1fr 1fr 1fr 70px;
		}

		.table-header p:nth-child(5),
		.table-header p:nth-child(6),
		.table-row .fees,
		.table-row .notes {
			display: none;
		}
	}

	@media (max-width: 480px) {
		.table-header {
			display: none;
		}

		.table-row {
			grid-template-columns: 1fr 1fr;
			gap: 0.5rem;
			background: rgba(255, 255, 255, 0.022);
			border: 1px solid var(--border-strong);
			border-radius: 8px;
			margin-bottom: 0.75rem;
		}

		.table-row:last-child {
			border-bottom: 1px solid var(--border-strong);
		}

		.fees,
		.notes {
			display: none;
		}
	}
</style>
