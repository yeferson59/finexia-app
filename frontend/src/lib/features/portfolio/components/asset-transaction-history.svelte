<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Modal from '$lib/ui/modal.svelte';
	import type { Holding, Transaction } from '$lib/api/types';
	import type { AssetActionResult, TxnMeta } from '../asset';
	import AssetTransactionForm from './asset-transaction-form.svelte';
	import AssetSellPanel from './asset-sell-panel.svelte';
	import AssetTransactionsTable from './asset-transactions-table.svelte';
	import AssetTransactionEditForm from './asset-transaction-edit-form.svelte';
	import AssetTransactionDeleteConfirm from './asset-transaction-delete-confirm.svelte';

	let {
		portfolioId,
		symbol,
		showAddForm = $bindable(false),
		entries,
		transactions,
		txnMeta,
		marketPrice,
		form,
		formatCurrency,
		formatAmount
	}: {
		portfolioId: string;
		symbol: string;
		/**
		 * Abierto el formulario de alta. Lo controla la página porque el botón
		 * que lo abre vive en la cabecera, junto al nombre del activo: allí es
		 * la acción principal de la ficha y no una más de la tabla.
		 */
		showAddForm?: boolean;
		entries: Holding[];
		transactions: Transaction[];
		txnMeta: TxnMeta;
		marketPrice: number | undefined;
		form: AssetActionResult | null;
		/** Importes de la posición, en la moneda de coste. */
		formatCurrency: (value: number, decimals?: number) => string;
		/** Importes de una transacción, en la moneda de esa transacción. */
		formatAmount: (value: number, currency: string) => string;
	} = $props();

	let sellFromTxn = $state<Transaction | null>(null);
	let editingTxn = $state<Transaction | null>(null);
	let deletingTxn = $state<Transaction | null>(null);

	const formError = $derived(form?.success === false);

	// Tras crear una transacción (no una edición) se cierra el formulario y se
	// recarga la página del activo.
	$effect(() => {
		if (form?.success === true && !form?.edited) {
			showAddForm = false;
			sellFromTxn = null;
			goto(
				resolve('/dashboard/portfolios/[id]/assets/[symbol]', {
					id: portfolioId,
					symbol
				})
			);
		}
	});
</script>

<section class="movements" aria-labelledby="movements-title">
	<header class="head">
		<h2 id="movements-title">Movimientos</h2>
		<p class="count">
			{txnMeta.total}
			{txnMeta.total === 1 ? 'movimiento' : 'movimientos'}
		</p>
	</header>

	<Modal
		open={showAddForm}
		title="Registrar transacción"
		onClose={() => (showAddForm = false)}
		size="lg"
	>
		<AssetTransactionForm {entries} {formError} onCancel={() => (showAddForm = false)} />
	</Modal>

	<Modal
		open={!!sellFromTxn}
		title="Vender posición"
		onClose={() => (sellFromTxn = null)}
		size="lg"
	>
		{#if sellFromTxn}
			<AssetSellPanel
				transaction={sellFromTxn}
				{entries}
				{marketPrice}
				fallbackCurrency={entries[0]?.costCurrency ?? 'USD'}
				formError={formError && !showAddForm}
				{formatCurrency}
				onClose={() => (sellFromTxn = null)}
			/>
		{/if}
	</Modal>

	<AssetTransactionsTable
		{transactions}
		{txnMeta}
		sellingTxnId={sellFromTxn?.id ?? null}
		{formatAmount}
		onEdit={(txn) => (editingTxn = txn)}
		onToggleSell={(txn) => (sellFromTxn = sellFromTxn?.id === txn.id ? null : txn)}
		onDelete={(txn) => (deletingTxn = txn)}
	/>
</section>

<Modal open={!!editingTxn} title="Editar transacción" onClose={() => (editingTxn = null)} size="lg">
	{#if editingTxn}
		<AssetTransactionEditForm transaction={editingTxn} onClose={() => (editingTxn = null)} />
	{/if}
</Modal>

<Modal
	open={!!deletingTxn}
	title="Eliminar transacción"
	onClose={() => (deletingTxn = null)}
	size="sm"
>
	{#if deletingTxn}
		<AssetTransactionDeleteConfirm
			transaction={deletingTxn}
			{formatAmount}
			onClose={() => (deletingTxn = null)}
		/>
	{/if}
</Modal>

<style>
	.movements {
		padding: 2rem 0;
		border-bottom: 1px solid var(--border);
	}

	.head {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.5rem 1rem;
	}

	h2 {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	/* El contador que era una tarjeta «TRANSACCIONES 6» arriba del todo. La
	   página que se está viendo la dice el pie de la tabla, junto a sus
	   botones, que es donde se cambia. */
	.count {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-dim);
	}
</style>
