<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import type { Holding, Transaction } from '$lib/api/types';
	import type { AssetActionResult, TxnMeta } from '../asset';
	import AssetTransactionForm from './asset-transaction-form.svelte';
	import AssetSellPanel from './asset-sell-panel.svelte';
	import AssetTransactionsTable from './asset-transactions-table.svelte';
	import AssetTransactionEditModal from './asset-transaction-edit-modal.svelte';
	import AssetTransactionDeleteDialog from './asset-transaction-delete-dialog.svelte';

	let {
		portfolioId,
		symbol,
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

	let showAddForm = $state(false);
	let sellFromTxn = $state<Transaction | null>(null);
	let editingTxn = $state<Transaction | null>(null);
	let deletingTxn = $state<Transaction | null>(null);

	const formError = $derived(form?.success === false);
	const currentPage = $derived(txnMeta.page);
	const totalPages = $derived(txnMeta.totalPages);

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

<section class="panel">
	<header class="panel-header">
		<h2>Historial de Transacciones</h2>
		<div class="header-actions">
			<span>
				{txnMeta.total}
				{txnMeta.total === 1 ? 'transacción' : 'transacciones'}
				{#if totalPages > 1}
					· página {currentPage} de {totalPages}
				{/if}
			</span>
			<button class="btn-add" onclick={() => (showAddForm = !showAddForm)}>
				{#if showAddForm}
					Cancelar
				{:else}
					+ Agregar
				{/if}
			</button>
		</div>
	</header>

	{#if showAddForm}
		<AssetTransactionForm {entries} {formError} onCancel={() => (showAddForm = false)} />
	{/if}

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

{#if editingTxn}
	<AssetTransactionEditModal transaction={editingTxn} onClose={() => (editingTxn = null)} />
{/if}

{#if deletingTxn}
	<AssetTransactionDeleteDialog
		transaction={deletingTxn}
		{formatAmount}
		onClose={() => (deletingTxn = null)}
	/>
{/if}

<style>
	.panel {
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		padding: 1.75rem;
		margin-bottom: 1.5rem;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--border);
	}

	.panel-header h2 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
	}

	.panel-header span {
		color: rgba(236, 234, 229, 0.4);
		font-size: 0.85rem;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.btn-add {
		padding: 0.4rem 0.9rem;
		border: 1.5px solid rgba(212, 145, 42, 0.4);
		border-radius: 6px;
		background: transparent;
		color: var(--amber);
		font-size: 0.82rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		font-family: var(--font-body);
	}

	.btn-add:hover {
		background: rgba(212, 145, 42, 0.12);
		border-color: var(--amber);
	}
</style>
