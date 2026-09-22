<script lang="ts">
	import Modal from '$lib/ui/modal.svelte';
	import { OptimisticList } from '$lib/shared/optimistic.svelte';
	import type { Holding, Transaction } from '$lib/api/types';
	import type { CashSource, TxnMeta } from '../asset';
	import AssetTransactionForm from './asset-transaction-form.svelte';
	import AssetSellPanel from './asset-sell-panel.svelte';
	import AssetTransactionsTable from './asset-transactions-table.svelte';
	import AssetTransactionEditForm from './asset-transaction-edit-form.svelte';
	import AssetTransactionDeleteConfirm from './asset-transaction-delete-confirm.svelte';

	let {
		showAddForm = $bindable(false),
		entries,
		cashByEntry = {},
		transactions,
		txnMeta,
		marketPrice,
		formatCurrency,
		formatAmount
	}: {
		/**
		 * Abierto el formulario de alta. Lo controla la página porque el botón
		 * que lo abre vive en la cabecera, junto al nombre del activo: allí es
		 * la acción principal de la ficha y no una más de la tabla.
		 */
		showAddForm?: boolean;
		entries: Holding[];
		/**
		 * De dónde puede salir el dinero de una compra en cada posición: la cuenta
		 * principal de su plataforma y sus bolsillos, en la moneda de la posición.
		 * Lo resuelve la página.
		 */
		cashByEntry?: Record<string, CashSource[]>;
		transactions: Transaction[];
		txnMeta: TxnMeta;
		marketPrice: number | undefined;
		/** Importes de la posición, en la moneda de coste. */
		formatCurrency: (value: number, decimals?: number) => string;
		/** Importes de una transacción, en la moneda de esa transacción. */
		formatAmount: (value: number, currency: string) => string;
	} = $props();

	let sellFromTxn = $state<Transaction | null>(null);
	let editingTxn = $state<Transaction | null>(null);
	let deletingTxn = $state<Transaction | null>(null);

	/*
	 * Los envíos no esperan al servidor: la fila nueva, la editada o la borrada
	 * se pintan al pulsar y el diálogo se cierra, aunque se queda montado —oculto—
	 * hasta la respuesta. Si el servidor la rechaza, se deshace lo pintado y el
	 * diálogo vuelve con lo escrito y el motivo. Mientras uno está en vuelo no se
	 * abre otro igual: compartirían la instancia del formulario.
	 */
	const pending = new OptimisticList<Transaction>();
	const shownTransactions = $derived(pending.view(transactions));
	const shownMeta = $derived({ ...txnMeta, total: Math.max(0, txnMeta.total + pending.delta) });

	let addHidden = $state(false);
	let sellHidden = $state(false);
	let editHidden = $state(false);

	/** El borrado que el servidor rechazó, con el diálogo ya cerrado. */
	let deleteError = $state('');

	function closeAdd() {
		showAddForm = false;
		addHidden = false;
	}

	function closeSell() {
		sellFromTxn = null;
		sellHidden = false;
	}

	function closeEdit() {
		editingTxn = null;
		editHidden = false;
	}

	function toggleSell(txn: Transaction) {
		if (sellHidden) return;
		sellFromTxn = sellFromTxn?.id === txn.id ? null : txn;
	}
</script>

<section class="movements" aria-labelledby="movements-title">
	<header class="head">
		<h2 id="movements-title">Movimientos</h2>
		<p class="count">
			{shownMeta.total}
			{shownMeta.total === 1 ? 'movimiento' : 'movimientos'}
		</p>
	</header>

	{#if deleteError}
		<p class="feedback error" role="alert">{deleteError}</p>
	{/if}

	<Modal
		open={showAddForm}
		hidden={addHidden}
		title="Registrar transacción"
		onClose={closeAdd}
		size="lg"
	>
		<AssetTransactionForm
			{entries}
			{cashByEntry}
			{pending}
			onSent={() => (addHidden = true)}
			onRejected={() => (addHidden = false)}
			onCancel={closeAdd}
		/>
	</Modal>

	<Modal
		open={!!sellFromTxn}
		hidden={sellHidden}
		title="Vender posición"
		onClose={closeSell}
		size="lg"
	>
		{#if sellFromTxn}
			<AssetSellPanel
				transaction={sellFromTxn}
				{entries}
				{marketPrice}
				fallbackCurrency={entries[0]?.costCurrency ?? 'USD'}
				{pending}
				{formatCurrency}
				onSent={() => (sellHidden = true)}
				onRejected={() => (sellHidden = false)}
				onClose={closeSell}
			/>
		{/if}
	</Modal>

	<AssetTransactionsTable
		transactions={shownTransactions}
		txnMeta={shownMeta}
		sellingTxnId={sellFromTxn?.id ?? null}
		isPending={(id) => pending.isPending(id)}
		{formatAmount}
		onEdit={(txn) => {
			if (!editHidden) editingTxn = txn;
		}}
		onToggleSell={toggleSell}
		onDelete={(txn) => {
			deleteError = '';
			deletingTxn = txn;
		}}
	/>
</section>

<Modal
	open={!!editingTxn}
	hidden={editHidden}
	title="Editar transacción"
	onClose={closeEdit}
	size="lg"
>
	{#if editingTxn}
		<AssetTransactionEditForm
			transaction={editingTxn}
			onCash={entries.find((e) => e.id === editingTxn?.entryId)?.assetType === 'cash'}
			cashSources={cashByEntry[editingTxn.entryId] ?? []}
			{pending}
			onSent={() => (editHidden = true)}
			onRejected={() => (editHidden = false)}
			onClose={closeEdit}
		/>
	{/if}
</Modal>

<Modal
	open={!!deletingTxn}
	title="Eliminar transacción"
	onClose={() => (deletingTxn = null)}
	size="sm"
	tone="danger"
>
	{#if deletingTxn}
		<AssetTransactionDeleteConfirm
			transaction={deletingTxn}
			{formatAmount}
			{pending}
			onRejected={(message) => (deleteError = message)}
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
	.feedback {
		margin: 1rem 0 0;
	}

	.count {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-dim);
	}
</style>
