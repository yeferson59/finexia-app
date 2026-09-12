<script lang="ts">
	/**
	 * En qué moneda liquidó cada posición, y cómo corregirla sin borrarla.
	 *
	 * La moneda de la cuenta se elige al abrir la posición —la casilla «Mi cuenta
	 * liquidó en otra moneda»— y la edición de una transacción no la puede
	 * cambiar. Dejarla sin marcar para un ETF en euros comprado desde una cuenta
	 * en dólares tenía una sola salida: borrar la posición y volver a cargarla.
	 * Aquí se cambia la moneda y se pone la tasa de cada operación, y el historial
	 * se queda como estaba.
	 *
	 * La historia entera se pide al abrir el diálogo y no con la página: la ficha
	 * solo carga una hoja de transacciones, y ninguna puede quedarse sin su tasa.
	 */
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import Modal from '$lib/ui/modal.svelte';
	import Button from '$lib/ui/button.svelte';
	import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { Holding, Transaction } from '$lib/api/types';
	import { TYPE_LABEL, transactionsNeedingRate } from '../asset';

	let {
		entries,
		formatAmount
	}: {
		entries: Holding[];
		formatAmount: (value: number, currency: string) => string;
	} = $props();

	let editing = $state<Holding | null>(null);
	let history = $state<Transaction[]>([]);
	let costCurrency = $state('');
	let loadingEntry = $state<string | null>(null);
	let loadError = $state('');
	let saving = $state(false);
	let saveError = $state('');

	const currencyOptions = $derived.by(() => {
		const options: string[] = [...SUPPORTED_CURRENCIES];
		for (const code of [editing?.currency, editing?.costCurrency]) {
			const normalized = code?.trim().toUpperCase();
			if (normalized && !options.includes(normalized)) options.unshift(normalized);
		}

		return options;
	});

	const needingRate = $derived(transactionsNeedingRate(history, costCurrency));

	function open(entry: Holding, transactions: Transaction[]) {
		editing = entry;
		history = transactions;
		costCurrency = entry.costCurrency.trim().toUpperCase();
		saveError = '';
	}

	function close() {
		editing = null;
		history = [];
		saveError = '';
	}

	/** La tasa guardada, si es una de verdad; una tasa 1 es la ausencia de una. */
	function storedRate(txn: Transaction): string {
		const rate = parseFloat(txn.fxRate ?? '');
		return Number.isFinite(rate) && rate !== 1 ? String(rate) : '';
	}

	function txnDate(value: string): string {
		return formatCalendarDate(value, { year: 'numeric', month: 'short', day: 'numeric' });
	}
</script>

<section class="settlement" aria-labelledby="settlement-title">
	<h2 class="title" id="settlement-title">Moneda de liquidación</h2>
	<p class="hint">
		La moneda en la que tu bróker cobró la compra. Si la registraste mal —un activo en euros
		comprado desde una cuenta en dólares—, cámbiala aquí sin borrar la posición.
	</p>

	{#each entries as entry (entry.id)}
		<div class="entry">
			<p class="detail">
				Cotiza en {entry.currency.trim().toUpperCase()}; tu cuenta la pagó en
				<strong>{entry.costCurrency.trim().toUpperCase()}</strong>.
			</p>
			<form
				method="POST"
				action="?/loadSettlement"
				use:enhance={() => {
					loadingEntry = entry.id;
					loadError = '';

					return async ({ result }) => {
						loadingEntry = null;

						if (result.type === 'success' && result.data?.success) {
							open(entry, (result.data.settlementTransactions as Transaction[]) ?? []);
							return;
						}

						loadError =
							(result.type === 'success' && (result.data?.error as string)) ||
							'No se pudo cargar el historial de la posición.';
					};
				}}
			>
				<input type="hidden" name="entryId" value={entry.id} />
				<button type="submit" class="change" disabled={loadingEntry === entry.id}>
					{loadingEntry === entry.id ? 'Cargando…' : 'Cambiar'}
				</button>
			</form>
		</div>
	{/each}

	{#if loadError}
		<p class="error" role="alert">{loadError}</p>
	{/if}
</section>

<Modal open={!!editing} title="Cambiar moneda de liquidación" onClose={close}>
	{#if editing}
		<form
			method="POST"
			action="?/changeSettlement"
			use:enhance={() => {
				saving = true;
				saveError = '';

				return async ({ result }) => {
					saving = false;

					if (result.type === 'success' && result.data?.success) {
						close();
						await invalidateAll();
						return;
					}

					saveError =
						(result.type === 'success' && (result.data?.error as string)) ||
						'No se pudo cambiar la moneda de liquidación.';
				};
			}}
		>
			<input type="hidden" name="entryId" value={editing.id} />

			<label class="field">
				<span class="label">Moneda de la cuenta</span>
				<select name="costCurrency" bind:value={costCurrency}>
					{#each currencyOptions as code (code)}
						<option value={code}>{code}</option>
					{/each}
				</select>
			</label>

			{#if needingRate.length === 0}
				<p class="note">
					Todas sus transacciones cotizan en {costCurrency}, así que no hace falta ninguna tasa.
				</p>
			{:else}
				<p class="note">
					La tasa de cada operación, tal como aparece en la confirmación del bróker: cuántos
					{costCurrency} costaba una unidad de la moneda de la operación ese día. No la de hoy.
				</p>

				<ul class="rates">
					{#each needingRate as txn (txn.id)}
						<li class="rate">
							<span class="txn">
								{TYPE_LABEL[txn.type] ?? txn.type} del {txnDate(txn.transactionDate)} ·
								{parseFloat(txn.quantity)} × {formatAmount(
									parseFloat(txn.price) || 0,
									txn.currency
								)}
							</span>
							<label class="rate-field">
								<span class="label">Tasa {txn.currency.trim().toUpperCase()} → {costCurrency}</span>
								<input
									type="number"
									name="rate:{txn.id}"
									value={storedRate(txn)}
									placeholder="1.0000"
									min="0"
									step="any"
									required
								/>
							</label>
						</li>
					{/each}
				</ul>
			{/if}

			{#if saveError}
				<p class="error" role="alert">{saveError}</p>
			{/if}

			<div class="modal-actions">
				<Button type="button" variant="ghost" onclick={close} disabled={saving}>Cancelar</Button>
				<Button type="submit" loading={saving}>Guardar</Button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	/* El mismo idioma que «Quitar esta posición»: un apartado de la ficha, sin caja. */
	.settlement {
		padding: 2rem 0 0;
	}

	.title {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	.hint {
		max-width: 64ch;
		margin: 0.5rem 0 0;
	}

	.entry {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem 1rem;
		margin-top: 1.1rem;
		padding-top: 1.1rem;
		border-top: 1px solid var(--border);
	}

	.detail {
		margin: 0;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.change {
		flex-shrink: 0;
		padding: 0.5rem 1.1rem;
		border: 1px solid var(--border-strong);
		border-radius: 9px;
		background: none;
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		transition: border-color 0.2s ease;
	}

	.change:hover:not(:disabled) {
		border-color: var(--amber);
	}

	.change:disabled {
		opacity: 0.6;
		cursor: wait;
	}

	@media (prefers-reduced-motion: reduce) {
		.change {
			transition: none;
		}
	}

	.field,
	.rate-field {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.label {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	select,
	input {
		padding: 0.55rem 0.75rem;
		border: 1px solid var(--border-strong);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.04);
		color: var(--text);
		font-family: var(--font-mono);
		font-size: 0.9rem;
	}

	.note {
		margin: 1rem 0 0;
		font-size: 0.85rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.rates {
		display: flex;
		flex-direction: column;
		gap: 0.9rem;
		margin: 1rem 0 0;
		padding: 0;
		list-style: none;
	}

	.rate {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(9rem, 12rem);
		align-items: end;
		gap: 0.75rem;
		padding-top: 0.9rem;
		border-top: 1px solid var(--border);
	}

	.txn {
		font-size: 0.85rem;
		color: var(--text);
		font-variant-numeric: tabular-nums;
	}

	.error {
		margin: 1rem 0 0;
		padding-left: 0.75rem;
		border-left: 2px solid var(--red);
		color: var(--red);
		font-size: 0.85rem;
		line-height: 1.5;
	}

	.modal-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
		margin-top: 1.25rem;
	}

	@media (max-width: 520px) {
		.rate {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
