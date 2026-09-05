<script lang="ts">
	import { enhance } from '$app/forms';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { todayLocalDateString } from '$lib/shared/format/date';
	import type { Holding, Transaction } from '$lib/api/types';
	import AssetSellPanelHeader from './asset-sell-panel-header.svelte';
	import AssetSellCurrencyFields from './asset-sell-currency-fields.svelte';

	let {
		transaction,
		entries,
		marketPrice,
		fallbackCurrency,
		formError = false,
		formatCurrency,
		onClose
	}: {
		transaction: Transaction;
		entries: Holding[];
		marketPrice: number | undefined;
		fallbackCurrency: string;
		formError?: boolean;
		formatCurrency: (value: number, decimals?: number) => string;
		onClose: () => void;
	} = $props();

	let sellMode = $state<'full' | 'partial'>('full');
	// Dentro de una venta parcial: por número de acciones o por valor total.
	let sellBasis = $state<'quantity' | 'value'>('quantity');
	let sellQty = $state('');
	let sellValue = $state('');
	let sellPrice = $state('');
	let sellFees = $state('');
	let sellFeesCurrency = $state('');
	let sellRate = $state('');
	let sellDate = $state(todayLocalDateString());
	let sellNotes = $state('');
	let isSellSubmitting = $state(false);

	$effect(() => {
		if (transaction) {
			sellMode = 'full';
			sellBasis = 'quantity';
			sellQty = transaction.quantity;
			sellValue = '';
			sellPrice = marketPrice ? marketPrice.toFixed(2) : transaction.price;
			sellFees = '';
			sellFeesCurrency = '';
			sellRate = '';
			sellNotes = '';
			sellDate = todayLocalDateString();
		}
	});

	$effect(() => {
		if (sellMode === 'full' && transaction) {
			sellBasis = 'quantity';
			sellQty = transaction.quantity;
		}
	});

	const sellLotMaxQty = $derived(parseFloat(transaction.quantity) || 0);

	// Cantidad que se envía: el lote completo, la tecleada o la derivada del valor.
	const sellEffectiveQty = $derived.by(() => {
		if (sellMode === 'full') return sellLotMaxQty;
		if (sellBasis === 'value') {
			const val = parseFloat(sellValue) || 0;
			const price = parseFloat(sellPrice) || 0;
			return price > 0 ? val / price : 0;
		}
		return parseFloat(sellQty) || 0;
	});

	const sellExceedsLot = $derived(
		sellMode === 'partial' && sellEffectiveQty > sellLotMaxQty + 1e-8
	);

	/**
	 * Las dos monedas de la posición que se está vendiendo.
	 *
	 * `sellPrice` se siembra del precio de mercado, que está en la moneda de
	 * cotización del activo; mandarlo etiquetado con la de la cuenta —que es lo
	 * que hacía este panel— registra una venta de 429,45 USD donde hubo uno de
	 * 429,45 EUR, y el resultado de la operación sale mal por la tasa entera.
	 */
	const sellEntry = $derived(entries.find((e) => e.id === transaction.entryId));
	const sellCostCurrency = $derived(
		sellEntry?.costCurrency?.trim().toUpperCase() || fallbackCurrency
	);
	const sellTradeCurrency = $derived(sellEntry?.currency?.trim().toUpperCase() || sellCostCurrency);
	const sellIsCrossCurrency = $derived(sellTradeCurrency !== sellCostCurrency);

	// Igual que en el alta de transacciones: la comisión arranca en la moneda de
	// la cuenta porque es de donde el bróker la cobra, y el formulario siempre la
	// manda explícita, así que el valor por defecto de la API no se le aplica.
	$effect(() => {
		if (!sellIsCrossCurrency) {
			sellFeesCurrency = '';
		} else if (sellFeesCurrency === '') {
			sellFeesCurrency = sellCostCurrency;
		}
	});

	const sellProceeds = $derived(
		sellEffectiveQty *
			(parseFloat(sellPrice) || 0) *
			(parseFloat(sellRate) || (sellIsCrossCurrency ? 0 : 1))
	);
</script>

<div class="sell-panel">
	<AssetSellPanelHeader {transaction} {formatCurrency} />

	<div class="sell-mode-toggle">
		<button
			type="button"
			class="sell-mode-btn"
			class:active={sellMode === 'full'}
			onclick={() => (sellMode = 'full')}
		>
			Venta Completa
		</button>
		<button
			type="button"
			class="sell-mode-btn"
			class:active={sellMode === 'partial'}
			onclick={() => (sellMode = 'partial')}
		>
			Venta Parcial
		</button>
	</div>

	<form
		method="POST"
		class="sell-form"
		action="?/createTransaction"
		use:enhance={() => {
			isSellSubmitting = true;
			return async ({ update }) => {
				await update({ reset: false });
				isSellSubmitting = false;
			};
		}}
	>
		<input type="hidden" name="entryId" value={transaction.entryId} />
		<input type="hidden" name="type" value="sell" />
		<input type="hidden" name="currency" value={sellTradeCurrency} />
		{#if !sellIsCrossCurrency}
			<input type="hidden" name="fxRate" value="1" />
		{/if}
		<input type="hidden" name="quantity" value={sellEffectiveQty} />

		{#if sellMode === 'partial'}
			<div class="sell-basis-toggle">
				<button
					type="button"
					class="sell-basis-btn"
					class:active={sellBasis === 'quantity'}
					onclick={() => (sellBasis = 'quantity')}
				>
					Por número de acciones
				</button>
				<button
					type="button"
					class="sell-basis-btn"
					class:active={sellBasis === 'value'}
					onclick={() => (sellBasis = 'value')}
				>
					Por valor de la venta
				</button>
			</div>
		{/if}

		<div class="form-row">
			{#if sellMode === 'full' || sellBasis === 'quantity'}
				<div class="form-group">
					<label class="form-label" for="sell-qty">
						Cantidad <span class="required">*</span>
						{#if sellMode === 'full'}
							<span class="sell-label-hint">(lote completo)</span>
						{/if}
					</label>
					<input
						id="sell-qty"
						type="number"
						class="form-input"
						bind:value={sellQty}
						disabled={sellMode === 'full'}
						min="0.00000001"
						max={sellLotMaxQty}
						step="0.00000001"
						required
					/>
				</div>
			{:else}
				<div class="form-group">
					<label class="form-label" for="sell-value"
						>Valor total de la venta <span class="required">*</span></label
					>
					<input
						id="sell-value"
						type="number"
						class="form-input"
						bind:value={sellValue}
						placeholder="0.00"
						min="0"
						step="0.01"
						required
					/>
					<span class="sell-computed-hint">
						≈ {sellEffectiveQty.toLocaleString('es-CO', { maximumFractionDigits: 8 })} unidades
					</span>
				</div>
			{/if}
			<div class="form-group">
				<label class="form-label" for="sell-price"
					>Precio unitario <span class="required">*</span></label
				>
				<input
					id="sell-price"
					type="number"
					class="form-input"
					name="price"
					bind:value={sellPrice}
					min="0"
					step="any"
					required
				/>
				{#if sellIsCrossCurrency}
					<span class="sell-computed-hint">en {sellTradeCurrency}</span>
				{/if}
			</div>
			<AssetSellCurrencyFields
				crossCurrency={sellIsCrossCurrency}
				tradeCurrency={sellTradeCurrency}
				costCurrency={sellCostCurrency}
				proceeds={sellProceeds}
				bind:rate={sellRate}
				bind:fees={sellFees}
				bind:feesCurrency={sellFeesCurrency}
			/>
			<div class="form-group">
				<label class="form-label" for="sell-date">Fecha <span class="required">*</span></label>
				<DatePicker name="transactionDate" bind:value={sellDate} required />
			</div>
		</div>

		{#if sellExceedsLot}
			<p class="form-error-msg">
				La cantidad supera el lote disponible ({sellLotMaxQty.toLocaleString('es-CO', {
					maximumFractionDigits: 8
				})} unidades).
			</p>
		{/if}

		<div class="form-group">
			<label class="form-label" for="sell-notes">Notas</label>
			<input
				id="sell-notes"
				type="text"
				class="form-input"
				name="notes"
				bind:value={sellNotes}
				placeholder="Observaciones opcionales..."
			/>
		</div>

		{#if formError}
			<p class="form-error-msg">No se pudo registrar la venta. Verifica los datos.</p>
		{/if}

		<div class="form-actions">
			<button type="button" class="btn-cancel" onclick={onClose}> Cancelar </button>
			<button
				type="submit"
				class="btn-sell-submit"
				disabled={isSellSubmitting || sellExceedsLot || sellEffectiveQty <= 0}
			>
				{isSellSubmitting
					? 'Guardando…'
					: sellMode === 'full'
						? 'Confirmar Venta Total'
						: 'Registrar Venta Parcial'}
			</button>
		</div>
	</form>
</div>

<style>
	/* El marco lo pone el modal; aquí sólo queda el ritmo vertical. */
	.sell-panel {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.sell-mode-toggle {
		display: flex;
		gap: 0.5rem;
	}

	.sell-mode-btn {
		padding: 0.45rem 1rem;
		border: 1.5px solid rgba(224, 90, 90, 0.3);
		border-radius: 6px;
		background: transparent;
		color: rgba(236, 234, 229, 0.6);
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.sell-mode-btn:hover {
		border-color: rgba(224, 90, 90, 0.5);
		color: var(--text);
	}

	.sell-mode-btn.active {
		background: rgba(224, 90, 90, 0.15);
		border-color: var(--red);
		color: var(--red);
	}

	.sell-basis-toggle {
		display: flex;
		gap: 0.5rem;
	}

	.sell-basis-btn {
		padding: 0.35rem 0.85rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 6px;
		background: transparent;
		color: rgba(236, 234, 229, 0.55);
		font-size: 0.8rem;
		font-weight: 600;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.sell-basis-btn:hover {
		border-color: rgba(212, 145, 42, 0.5);
		color: var(--text);
	}

	.sell-basis-btn.active {
		background: rgba(212, 145, 42, 0.15);
		border-color: var(--amber);
		color: var(--amber);
	}

	.sell-computed-hint {
		font-size: 0.78rem;
		color: rgba(236, 234, 229, 0.5);
		font-family: var(--font-mono);
	}

	.sell-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.sell-label-hint {
		font-weight: 400;
		color: var(--text-dim);
		text-transform: none;
		letter-spacing: 0;
		font-size: 0.75rem;
	}

	.btn-sell-submit {
		padding: 0.55rem 1.25rem;
		border: none;
		border-radius: 7px;
		background: var(--red);
		color: #fff;
		font-size: 0.88rem;
		font-weight: 700;
		cursor: pointer;
		font-family: var(--font-body);
		transition: all 0.2s ease;
	}

	.btn-sell-submit:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 6px 16px rgba(224, 90, 90, 0.25);
	}

	.btn-sell-submit:disabled {
		opacity: 0.6;
		cursor: not-allowed;
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
</style>
