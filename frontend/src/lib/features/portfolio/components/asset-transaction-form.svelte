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
		fxRate: '',
		fees: '',
		feesCurrency: '',
		transactionDate: todayLocalDateString(),
		notes: ''
	});

	$effect(() => {
		txnForm.entryId = entries[0]?.id ?? '';
	});

	let isSubmitting = $state(false);

	const txnMode = $derived(txnModeFor(txnForm.type));
	const priceLabel = $derived(priceLabelFor(txnForm.type));

	/**
	 * La posición sobre la que se registra, y sus dos monedas.
	 *
	 * `costCurrency` es en la que se lleva el coste —la de la cuenta— y
	 * `currency` la de cotización del activo. Cuando difieren, cada operación
	 * nueva sobre esta posición volvió a pasar por una conversión, y sin su tasa
	 * el coste medio se calcularía sumando euros a dólares.
	 */
	const entry = $derived(entries.find((e) => e.id === txnForm.entryId) ?? entries[0]);
	const costCurrency = $derived(entry?.costCurrency?.trim().toUpperCase() || 'USD');
	const assetCurrency = $derived(entry?.currency?.trim().toUpperCase() || costCurrency);

	// El split no mueve dinero —el precio va a 0— así que pedir una tasa para él
	// sería un campo obligatorio que no cambia ningún número.
	const crossCurrency = $derived(assetCurrency !== costCurrency && txnMode !== 'split');

	/**
	 * La comisión arranca en la moneda de la cuenta, no en la de la operación.
	 *
	 * Es lo contrario del valor por defecto de la API, y a propósito: la API
	 * empareja `fees` con `price` porque es la lectura menos sorprendente de dos
	 * campos contiguos, pero un bróker cobra la comisión de la cuenta más a
	 * menudo que de la ejecución —la confirmación que motivó todo esto cotiza en
	 * EUR y cobra 0,00 USD—. Este formulario siempre manda el campo, así que el
	 * valor por defecto de la API nunca se le aplica y las dos elecciones no se
	 * pisan.
	 */
	$effect(() => {
		txnForm.currency = crossCurrency ? assetCurrency : costCurrency;
		if (!crossCurrency) {
			txnForm.fxRate = '';
			txnForm.feesCurrency = '';
		} else if (txnForm.feesCurrency === '') {
			txnForm.feesCurrency = costCurrency;
		}
	});

	const rate = $derived(parseFloat(txnForm.fxRate) || (crossCurrency ? 0 : 1));
	const settledTotal = $derived(
		(parseFloat(txnForm.quantity) || 0) * (parseFloat(txnForm.price) || 0) * rate
	);

	function formatIn(value: number, code: string): string {
		return new Intl.NumberFormat('es-CO', {
			style: 'currency',
			currency: code,
			minimumFractionDigits: 2
		}).format(value);
	}
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
	{#if !crossCurrency}
		<input type="hidden" name="fxRate" value="1" />
	{/if}

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
					step="any"
					required
				/>
			</div>
			<div class="form-group">
				<label class="form-label" for="txn-fees">Comisión</label>
				{#if crossCurrency}
					<div class="fees-field">
						<select
							class="form-select"
							name="feesCurrency"
							bind:value={txnForm.feesCurrency}
							aria-label="Moneda de la comisión"
						>
							<option value={costCurrency}>{costCurrency}</option>
							<option value={assetCurrency}>{assetCurrency}</option>
						</select>
						<input
							id="txn-fees"
							type="number"
							class="form-input"
							name="fees"
							bind:value={txnForm.fees}
							placeholder="0"
							min="0"
							step="any"
						/>
					</div>
					<p class="hint">La que el bróker cobró, no siempre la de la ejecución.</p>
				{:else}
					<input
						id="txn-fees"
						type="number"
						class="form-input"
						name="fees"
						bind:value={txnForm.fees}
						placeholder="0"
						min="0"
						step="any"
					/>
				{/if}
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

	{#if crossCurrency}
		<div class="form-row fx-row">
			<div class="form-group">
				<span class="form-label">Moneda de la operación</span>
				<p class="fx-static">{txnForm.currency}</p>
				<p class="hint">
					{assetCurrency} es la moneda en la que cotiza el activo; el precio y la comisión de arriba van
					en ella.
				</p>
			</div>
			<div class="form-group">
				<label class="form-label" for="txn-fx"
					>Tasa a {costCurrency} <span class="required">*</span></label
				>
				<input
					id="txn-fx"
					type="number"
					class="form-input"
					name="fxRate"
					bind:value={txnForm.fxRate}
					placeholder="1.0638"
					min="0"
					step="any"
					required
				/>
				<p class="hint">
					Cuántos {costCurrency} costaba 1 {assetCurrency} ese día, según la confirmación del bróker.
				</p>
			</div>
			<div class="form-group">
				<span class="form-label">Coste en {costCurrency}</span>
				<p class="fx-static">{formatIn(settledTotal, costCurrency)}</p>
				<p class="hint">Contrástalo con el importe que te debitaron.</p>
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

	.fees-field {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 0.4rem;
		align-items: center;
	}

	.fx-row {
		padding: 0.9rem;
		border: 1px dashed rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: rgba(212, 145, 42, 0.04);
	}

	.fx-static {
		margin: 0;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--amber);
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
