<script lang="ts">
	import { enhance } from '$app/forms';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { formatCalendarDate, todayLocalDateString } from '$lib/shared/format/date';
	import type { Holding } from '$lib/api/types';
	import { TRANSACTION_TYPES, priceLabelFor, txnModeFor } from '../asset';

	let {
		entries,
		formError = false,
		onCancel
	}: { entries: Holding[]; formError?: boolean; onCancel: () => void } = $props();

	// entryId/currency se rellenan de forma reactiva desde la primera entrada.
	let txnForm = $state({
		entryId: '',
		type: 'buy',
		quantity: '',
		price: '',
		currency: 'USD',
		fees: '',
		transactionDate: todayLocalDateString(),
		notes: ''
	});

	$effect(() => {
		txnForm.entryId = entries[0]?.id ?? '';
		txnForm.currency = entries[0]?.costCurrency ?? 'USD';
	});

	let isSubmitting = $state(false);

	const txnMode = $derived(txnModeFor(txnForm.type));
	const priceLabel = $derived(priceLabelFor(txnForm.type));
</script>

<form
	method="POST"
	class="add-txn-form"
	action="?/createTransaction"
	use:enhance={() => {
		isSubmitting = true;
		return async ({ update }) => {
			await update({ reset: false });
			isSubmitting = false;
		};
	}}
>
	<input type="hidden" name="entryId" value={txnForm.entryId} />
	<input type="hidden" name="currency" value={txnForm.currency} />

	{#if entries.length > 1}
		<div class="form-row">
			<div class="form-group">
				<label class="form-label" for="txn-platform">Plataforma</label>
				<select id="txn-platform" class="form-select" name="entryId" bind:value={txnForm.entryId}>
					{#each entries as entry (entry.id)}
						<option value={entry.id}>
							{entry.costCurrency} · {formatCalendarDate(entry.entryDate, {
								year: 'numeric',
								month: 'short',
								day: 'numeric'
							})}
						</option>
					{/each}
				</select>
			</div>
		</div>
	{/if}

	<div class="form-row">
		<div class="form-group">
			<label class="form-label" for="txn-type">Tipo <span class="required">*</span></label>
			<select id="txn-type" class="form-select" name="type" bind:value={txnForm.type} required>
				{#each TRANSACTION_TYPES as t (t.value)}
					<option value={t.value}>{t.label}</option>
				{/each}
			</select>
		</div>
		<div class="form-group">
			<label class="form-label" for="txn-date">Fecha <span class="required">*</span></label>
			<DatePicker name="transactionDate" bind:value={txnForm.transactionDate} required />
		</div>
	</div>

	<!-- trade: cantidad + precio unitario + comisión -->
	{#if txnMode === 'trade'}
		<div class="form-row">
			<div class="form-group">
				<label class="form-label" for="txn-qty">Cantidad <span class="required">*</span></label>
				<input
					id="txn-qty"
					type="number"
					class="form-input"
					name="quantity"
					bind:value={txnForm.quantity}
					placeholder="100"
					min="0"
					step="any"
					required
				/>
			</div>
			<div class="form-group">
				<label class="form-label" for="txn-price"
					>Precio unitario <span class="required">*</span></label
				>
				<input
					id="txn-price"
					type="number"
					class="form-input"
					name="price"
					bind:value={txnForm.price}
					placeholder="150.50"
					min="0"
					step="0.01"
					required
				/>
			</div>
			<div class="form-group">
				<label class="form-label" for="txn-fees">Comisión</label>
				<input
					id="txn-fees"
					type="number"
					class="form-input"
					name="fees"
					bind:value={txnForm.fees}
					placeholder="0"
					min="0"
					step="0.01"
				/>
			</div>
		</div>

		<!-- amount: solo monto (dividendo / comisión / interés) -->
	{:else if txnMode === 'amount'}
		<input type="hidden" name="quantity" value="1" />
		<input type="hidden" name="fees" value="0" />
		<div class="form-row">
			<div class="form-group">
				<label class="form-label" for="txn-amount"
					>{priceLabel} <span class="required">*</span></label
				>
				<input
					id="txn-amount"
					type="number"
					class="form-input"
					name="price"
					bind:value={txnForm.price}
					placeholder="0.00"
					min="0"
					step="0.01"
					required
				/>
			</div>
		</div>

		<!-- split: nuevas acciones, sin precio -->
	{:else}
		<input type="hidden" name="price" value="0" />
		<input type="hidden" name="fees" value="0" />
		<div class="form-row">
			<div class="form-group">
				<label class="form-label" for="txn-split-qty"
					>Nuevas acciones recibidas <span class="required">*</span></label
				>
				<input
					id="txn-split-qty"
					type="number"
					class="form-input"
					name="quantity"
					bind:value={txnForm.quantity}
					placeholder="100"
					min="0"
					step="0.00000001"
					required
				/>
			</div>
		</div>
	{/if}

	<div class="form-group">
		<label class="form-label" for="txn-notes">Notas</label>
		<input
			id="txn-notes"
			type="text"
			class="form-input"
			name="notes"
			bind:value={txnForm.notes}
			placeholder="Observaciones opcionales..."
		/>
	</div>

	{#if formError}
		<p class="form-error-msg">No se pudo registrar la transacción. Verifica los datos.</p>
	{/if}

	<div class="form-actions">
		<button type="button" class="btn-cancel" onclick={onCancel}> Cancelar </button>
		<button type="submit" class="btn-submit" disabled={isSubmitting}>
			{isSubmitting ? 'Guardando…' : 'Registrar transacción'}
		</button>
	</div>
</form>

<style>
	.add-txn-form {
		margin-bottom: 1.5rem;
		padding: 1.25rem;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 10px;
		background: rgba(212, 145, 42, 0.04);
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 1rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.form-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: rgba(236, 234, 229, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.required {
		color: var(--red);
	}

	.form-input {
		padding: 0.6rem 0.85rem;
		border: 1.5px solid rgba(212, 145, 42, 0.2);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.04);
		color: var(--text);
		font-size: 0.9rem;
		font-family: var(--font-body);
		transition: border-color 0.2s ease;
	}

	.form-input:focus {
		outline: none;
		border-color: var(--amber);
	}

	.form-select {
		padding: 0.6rem 2.2rem 0.6rem 0.85rem;
		border: 1.5px solid rgba(212, 145, 42, 0.35);
		border-radius: 8px;
		background-color: rgba(212, 145, 42, 0.06);
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%23d4912a' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.65rem center;
		background-size: 1rem;
		appearance: none;
		-webkit-appearance: none;
		color: var(--text);
		font-size: 0.9rem;
		font-family: var(--font-body);
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			background-color 0.2s ease;
	}

	.form-select:focus {
		outline: none;
		border-color: var(--amber);
		background-color: rgba(212, 145, 42, 0.1);
	}

	.form-error-msg {
		margin: 0;
		padding: 0.6rem 0.9rem;
		border-radius: 6px;
		background: rgba(224, 90, 90, 0.1);
		border: 1px solid rgba(224, 90, 90, 0.3);
		color: rgba(224, 90, 90, 0.9);
		font-size: 0.85rem;
	}

	.form-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
	}

	.btn-cancel {
		padding: 0.55rem 1.1rem;
		border: 1.5px solid rgba(212, 145, 42, 0.2);
		border-radius: 7px;
		background: transparent;
		color: rgba(236, 234, 229, 0.6);
		font-size: 0.88rem;
		font-weight: 600;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.btn-cancel:hover {
		border-color: rgba(212, 145, 42, 0.4);
		color: var(--text);
	}

	.btn-submit {
		padding: 0.55rem 1.25rem;
		border: none;
		border-radius: 7px;
		background: var(--amber);
		color: #0d0800;
		font-size: 0.88rem;
		font-weight: 700;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.btn-submit:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 6px 16px rgba(212, 145, 42, 0.25);
	}

	.btn-submit:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
</style>
