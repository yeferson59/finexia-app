<script lang="ts">
	/**
	 * Cuerpo del formulario de edición de una transacción. El diálogo lo pone el
	 * `Modal` del historial, que es quien tiene el estado que lo abre.
	 */
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import { untrack } from 'svelte';
	import Button from '$lib/ui/button.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import type { Transaction } from '$lib/api/types';
	import { TRANSACTION_TYPES, formatSettled, priceLabelFor, txnModeFor } from '../asset';
	import AssetCreditCashField from './asset-credit-cash-field.svelte';
	import AssetPayFromCashField from './asset-pay-from-cash-field.svelte';

	let {
		transaction,
		onCash = false,
		cashAvailable = 0,
		onClose
	}: {
		transaction: Transaction;
		/** La posición es un saldo de efectivo, que no se abona dividendos ni ventas a sí mismo. */
		onCash?: boolean;
		/**
		 * Lo que queda en el efectivo de la plataforma, en la moneda de la cuenta.
		 * Si la compra ya se paga de ahí, este saldo ya lo tiene descontado.
		 */
		cashAvailable?: number;
		onClose: () => void;
	} = $props();

	// El modal monta este formulario al abrirse y lo desmonta al cerrarlo, así
	// que los campos se llenan una sola vez, en el montaje: es un borrador, y
	// reaccionar a la prop pisaría lo que el usuario lleve escrito. Antes se
	// recargaban desde un `$effect` porque el contenedor reutilizaba la instancia
	// al pasar de una fila a otra.
	let editForm = $state(
		untrack(() => ({
			type: transaction.type,
			quantity: transaction.quantity,
			price: transaction.price,
			currency: transaction.currency,
			fxRate: transaction.fxRate ?? '1',
			fees: transaction.fees,
			// La que trae la transacción, no un valor por defecto: editar una nota
			// no puede reinterpretar de qué lado se cobró la comisión.
			feesCurrency: transaction.feesCurrency ?? transaction.currency,
			transactionDate: transaction.transactionDate.split('T')[0],
			notes: transaction.notes,
			// Lo que el dividendo tiene hoy. El PUT reemplaza la fila entera, y
			// sin la casilla un cambio de nota sacaría el dividendo del efectivo.
			creditCash: transaction.cashCredited ?? false,
			// Y lo que la compra tiene hoy, por la misma razón. Marcarla aquí es
			// lo que registra, después del hecho, que se pagó con el efectivo de
			// la cuenta; desmarcarla devuelve el dinero al saldo.
			payFromCash: transaction.cashPaid ?? false
		}))
	);

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

	const canCreditCash = $derived(
		(editForm.type === 'dividend' || editForm.type === 'sell') && !onCash
	);
	const creditCurrency = $derived(editCostCurrency || editForm.currency);

	/**
	 * Pagar la compra con el efectivo de la plataforma. Aquí la casilla se ofrece
	 * aunque el saldo sea cero: puede que la compra ya salga de él —y entonces
	 * desmarcarla es cómo se deshace— y el saldo que se ve ya está descontado.
	 */
	const alreadyPaid = $derived(transaction.cashPaid ?? false);
	const canPayFromCash = $derived(editForm.type === 'buy' && !onCash);
	// La misma cuenta que hace el backend: la comisión va por la tasa solo si se
	// cobró en la moneda de la operación; cobrada en la de la cuenta ya está en
	// ella.
	const payRate = $derived(parseFloat(editForm.fxRate) || 1);
	const payFees = $derived(
		(parseFloat(editForm.fees) || 0) * (editForm.feesCurrency === editForm.currency ? payRate : 1)
	);
	const payAmount = $derived(
		(parseFloat(editForm.quantity) || 0) * (parseFloat(editForm.price) || 0) * payRate + payFees
	);
</script>

<form
	method="POST"
	action="?/editTransaction"
	class="txn-form"
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
		<div class="form-group date-field">
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

	{#if canCreditCash}
		<AssetCreditCashField bind:checked={editForm.creditCash} currency={creditCurrency} editing />
	{/if}

	{#if canPayFromCash}
		<AssetPayFromCashField
			bind:checked={editForm.payFromCash}
			currency={creditCurrency}
			amount={payAmount > 0 ? formatSettled(payAmount, creditCurrency) : ''}
			balance={formatSettled(cashAvailable, creditCurrency)}
			enough={alreadyPaid || payAmount <= cashAvailable}
			editing
		/>
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
		<p class="feedback error" role="alert">
			{editErrorMessage || 'No se pudo actualizar la transacción. Verifica los datos.'}
		</p>
	{/if}

	<div class="modal-actions">
		<Button type="button" variant="ghost" onclick={onClose} disabled={isEditSubmitting}>
			Cancelar
		</Button>
		<Button type="submit" loading={isEditSubmitting}>Guardar cambios</Button>
	</div>
</form>

<style>
	.txn-form {
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
		font-size: 0.87rem;
		font-weight: 500;
		color: var(--text);
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
	/* En el móvil los tres huecos de la fecha no caben al lado de otro campo:
	   el año se salía por la derecha del diálogo. */
	@media (max-width: 640px) {
		.date-field {
			grid-column: 1 / -1;
		}
	}
</style>
