<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import { untrack } from 'svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import type { Transaction } from '$lib/api/types';
	import { TRANSACTION_TYPES, priceLabelFor, txnModeFor } from '../asset';

	let { transaction, onClose }: { transaction: Transaction; onClose: () => void } = $props();

	let editForm = $state({
		type: '',
		quantity: '',
		price: '',
		currency: 'USD',
		fxRate: '1',
		fees: '',
		feesCurrency: '',
		transactionDate: '',
		notes: ''
	});

	// Recarga el formulario cada vez que cambia la transacción editada (el
	// contenedor puede reutilizar la instancia al pasar de una fila a otra).
	$effect(() => {
		const txn = transaction;
		untrack(() => {
			editForm = {
				type: txn.type,
				quantity: txn.quantity,
				price: txn.price,
				currency: txn.currency,
				fxRate: txn.fxRate ?? '1',
				fees: txn.fees,
				// La que trae la transacción, no un valor por defecto: editar una
				// nota no puede reinterpretar de qué lado se cobró la comisión.
				feesCurrency: txn.feesCurrency ?? txn.currency,
				transactionDate: txn.transactionDate.split('T')[0],
				notes: txn.notes
			};
			editError = false;
			editErrorMessage = '';
		});
	});

	let isEditSubmitting = $state(false);
	let editError = $state(false);
	let editErrorMessage = $state('');

	const editTxnMode = $derived(txnModeFor(editForm.type));
	const editPriceLabel = $derived(priceLabelFor(editForm.type));

	/**
	 * La transacción se convirtió al liquidarse.
	 *
	 * `PUT /portfolios/transactions/:id` reemplaza la fila entera, así que el
	 * formulario tiene que devolver la tasa igual que devuelve el precio. Si no
	 * la mandara, editar una nota de una compra en euros liquidada en dólares la
	 * regrabaría con tasa 1 y el coste de la posición se reescribiría a la baja
	 * sin que nada lo señalara.
	 */
	const editCostCurrency = $derived(transaction.costCurrency?.trim().toUpperCase() || '');
	const editIsCrossCurrency = $derived(
		!!editCostCurrency && editCostCurrency !== editForm.currency
	);
</script>

<div
	class="modal-backdrop"
	role="button"
	tabindex="0"
	aria-label="Cerrar modal"
	onclick={onClose}
	onkeydown={(e) => e.key === 'Enter' && onClose()}
></div>
<div class="modal" role="dialog" aria-modal="true" aria-label="Editar transacción">
	<header class="modal-header">
		<span>Editar transacción</span>
		<button class="modal-close" type="button" onclick={onClose}>✕</button>
	</header>

	<form
		method="POST"
		action="?/editTransaction"
		class="modal-form"
		use:enhance={() => {
			isEditSubmitting = true;
			editError = false;
			editErrorMessage = '';
			return async ({ result, update }) => {
				await update({ reset: false });
				isEditSubmitting = false;
				const data =
					result.type === 'success'
						? (result.data as { success?: boolean; error?: string } | undefined)
						: undefined;
				if (data?.success) {
					onClose();
					await invalidateAll();
				} else {
					editError = true;
					editErrorMessage = data?.error ?? '';
				}
			};
		}}
	>
		<input type="hidden" name="txnId" value={transaction.id} />
		<input type="hidden" name="currency" value={editForm.currency} />
		{#if !editIsCrossCurrency}
			<input type="hidden" name="fxRate" value={editForm.fxRate || '1'} />
			<input type="hidden" name="feesCurrency" value={editForm.feesCurrency || editForm.currency} />
		{/if}

		<div class="form-row">
			<div class="form-group">
				<label class="form-label" for="edit-type">Tipo <span class="required">*</span></label>
				<select id="edit-type" class="form-select" name="type" bind:value={editForm.type} required>
					{#each TRANSACTION_TYPES as t (t.value)}
						<option value={t.value}>{t.label}</option>
					{/each}
				</select>
			</div>
			<div class="form-group">
				<span class="form-label">Fecha <span class="required">*</span></span>
				<DatePicker name="transactionDate" bind:value={editForm.transactionDate} required />
			</div>
		</div>

		{#if editTxnMode === 'trade'}
			<div class="form-row">
				<div class="form-group">
					<label class="form-label" for="edit-qty">Cantidad <span class="required">*</span></label>
					<input
						id="edit-qty"
						type="number"
						class="form-input"
						name="quantity"
						bind:value={editForm.quantity}
						placeholder="100"
						min="0"
						step="any"
						required
					/>
				</div>
				<div class="form-group">
					<label class="form-label" for="edit-price"
						>{editPriceLabel} <span class="required">*</span></label
					>
					<input
						id="edit-price"
						type="number"
						class="form-input"
						name="price"
						bind:value={editForm.price}
						placeholder="150.50"
						min="0"
						step="any"
						required
					/>
				</div>
				<div class="form-group">
					<label class="form-label" for="edit-fees">Comisión</label>
					<input
						id="edit-fees"
						type="number"
						class="form-input"
						name="fees"
						bind:value={editForm.fees}
						placeholder="0"
						min="0"
						step="0.01"
					/>
				</div>
			</div>
		{:else if editTxnMode === 'amount'}
			<input type="hidden" name="quantity" value="1" />
			<input type="hidden" name="fees" value="0" />
			<div class="form-row">
				<div class="form-group">
					<label class="form-label" for="edit-amount"
						>{editPriceLabel} <span class="required">*</span></label
					>
					<input
						id="edit-amount"
						type="number"
						class="form-input"
						name="price"
						bind:value={editForm.price}
						placeholder="0.00"
						min="0"
						step="0.01"
						required
					/>
				</div>
			</div>
		{:else}
			<input type="hidden" name="price" value="0" />
			<input type="hidden" name="fees" value="0" />
			<div class="form-row">
				<div class="form-group">
					<label class="form-label" for="edit-split-qty"
						>Nuevas acciones recibidas <span class="required">*</span></label
					>
					<input
						id="edit-split-qty"
						type="number"
						class="form-input"
						name="quantity"
						bind:value={editForm.quantity}
						placeholder="100"
						min="0"
						step="0.00000001"
						required
					/>
				</div>
			</div>
		{/if}

		{#if editIsCrossCurrency}
			<div class="form-row">
				<div class="form-group">
					<label class="form-label" for="edit-fx"
						>Tasa {editForm.currency} → {editCostCurrency} <span class="required">*</span></label
					>
					<input
						id="edit-fx"
						type="number"
						class="form-input"
						name="fxRate"
						bind:value={editForm.fxRate}
						min="0"
						step="any"
						required
					/>
				</div>
				<div class="form-group">
					<label class="form-label" for="edit-fees-currency">Moneda de la comisión</label>
					<select
						id="edit-fees-currency"
						class="form-select"
						name="feesCurrency"
						bind:value={editForm.feesCurrency}
					>
						<option value={editCostCurrency}>{editCostCurrency}</option>
						<option value={editForm.currency}>{editForm.currency}</option>
					</select>
				</div>
			</div>
		{/if}

		<div class="form-group">
			<label class="form-label" for="edit-notes">Notas</label>
			<input
				id="edit-notes"
				type="text"
				class="form-input"
				name="notes"
				bind:value={editForm.notes}
				placeholder="Observaciones opcionales..."
			/>
		</div>

		{#if editError}
			<p class="form-error-msg">
				{editErrorMessage || 'No se pudo actualizar la transacción. Verifica los datos.'}
			</p>
		{/if}

		<div class="form-actions">
			<button type="button" class="btn-cancel" onclick={onClose}>Cancelar</button>
			<button type="submit" class="btn-submit" disabled={isEditSubmitting}>
				{isEditSubmitting ? 'Guardando…' : 'Guardar cambios'}
			</button>
		</div>
	</form>
</div>

<style>
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.65);
		z-index: 100;
	}

	.modal {
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		z-index: 101;
		width: min(540px, 92vw);
		background: var(--surface);
		border: 1.5px solid rgba(212, 145, 42, 0.35);
		border-radius: 16px;
		box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
		backdrop-filter: blur(16px);
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 1.25rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--amber);
	}

	.modal-close {
		background: transparent;
		border: none;
		color: rgba(236, 234, 229, 0.4);
		font-size: 1rem;
		cursor: pointer;
		padding: 0.2rem 0.4rem;
		border-radius: 4px;
		transition: color 0.2s ease;
		line-height: 1;
	}

	.modal-close:hover {
		color: var(--text);
	}

	.modal-form {
		padding: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		max-height: 80vh;
		overflow-y: auto;
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
