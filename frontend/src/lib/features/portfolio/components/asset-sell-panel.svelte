<script lang="ts">
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { todayLocalDateString } from '$lib/shared/format/date';
	import type { Holding, Transaction } from '$lib/api/types';
	import AssetSellPanelHeader from './asset-sell-panel-header.svelte';
	import AssetSellCurrencyFields from './asset-sell-currency-fields.svelte';
	import AssetCreditCashField from './asset-credit-cash-field.svelte';

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
	// Lo recibido va al efectivo de la plataforma, como hace el bróker.
	let sellCreditCash = $state(true);
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
			sellCreditCash = true;
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

	/**
	 * Lo que llega al efectivo: lo cobrado menos la comisión en la moneda de la
	 * cuenta. Una comisión cobrada en la moneda del mercado pasa por la tasa.
	 */
	const sellCredited = $derived(
		sellProceeds -
			(parseFloat(sellFees) || 0) *
				(sellIsCrossCurrency && sellFeesCurrency !== sellCostCurrency
					? parseFloat(sellRate) || 0
					: 1)
	);
	const sellCanCreditCash = $derived(sellEntry?.assetType !== 'cash');

	function formatCredit(value: number): string {
		return new Intl.NumberFormat('es-CO', {
			style: 'currency',
			currency: sellCostCurrency,
			minimumFractionDigits: 2
		}).format(value);
	}
</script>

<div class="sell-panel">
	<AssetSellPanelHeader {transaction} {formatCurrency} />

	<div class="segmented" role="group" aria-label="Cuánto vender">
		<button type="button" aria-pressed={sellMode === 'full'} onclick={() => (sellMode = 'full')}>
			Venta completa
		</button>
		<button
			type="button"
			aria-pressed={sellMode === 'partial'}
			onclick={() => (sellMode = 'partial')}
		>
			Venta parcial
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
			<div class="segmented" role="group" aria-label="Cómo indicar la venta parcial">
				<button
					type="button"
					aria-pressed={sellBasis === 'quantity'}
					onclick={() => (sellBasis = 'quantity')}
				>
					Por número de acciones
				</button>
				<button
					type="button"
					aria-pressed={sellBasis === 'value'}
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
		</div>

		{#if sellExceedsLot}
			<p class="feedback error" role="alert">
				La cantidad supera el lote disponible ({sellLotMaxQty.toLocaleString('es-CO', {
					maximumFractionDigits: 8
				})} unidades).
			</p>
		{/if}

		<div class="form-row meta">
			<div class="form-group">
				<label class="form-label" for="sell-date">Fecha <span class="required">*</span></label>
				<DatePicker name="transactionDate" bind:value={sellDate} required />
			</div>
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
		</div>

		{#if sellCanCreditCash}
			<AssetCreditCashField
				bind:checked={sellCreditCash}
				currency={sellCostCurrency}
				amount={sellCredited > 0 ? formatCredit(sellCredited) : ''}
			/>
		{/if}

		{#if formError}
			<p class="feedback error" role="alert">No se pudo registrar la venta. Verifica los datos.</p>
		{/if}

		<div class="modal-actions">
			<Button type="button" variant="ghost" onclick={onClose} disabled={isSellSubmitting}
				>Cancelar</Button
			>
			<Button
				type="submit"
				loading={isSellSubmitting}
				disabled={sellExceedsLot || sellEffectiveQty <= 0}
			>
				{sellMode === 'full' ? 'Confirmar venta total' : 'Registrar venta parcial'}
			</Button>
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

	/*
	 * Las dos elecciones del panel —cuánto se vende y cómo se indica— son el
	 * mismo gesto, así que se ven igual: un interruptor de dos posiciones. Antes
	 * una iba en rojo y la otra en ámbar, y el rojo decía «borrar» donde solo se
	 * elegía un modo.
	 */
	.segmented {
		display: inline-flex;
		align-self: flex-start;
		flex-wrap: wrap;
		max-width: 100%;
		padding: 3px;
		border: 1px solid var(--border-strong);
		border-radius: 9px;
		background: rgba(255, 255, 255, 0.02);
	}

	.segmented button {
		padding: 0.45rem 0.95rem;
		border: none;
		border-radius: 6px;
		background: transparent;
		color: var(--text-muted);
		font-family: var(--font-body);
		font-size: 0.85rem;
		font-weight: 500;
		cursor: pointer;
		transition:
			color 0.2s ease,
			background-color 0.2s ease;
	}

	.segmented button:hover {
		color: var(--text);
	}

	.segmented button[aria-pressed='true'] {
		background: rgba(212, 145, 42, 0.12);
		color: var(--text);
		box-shadow: inset 0 0 0 1px rgba(212, 145, 42, 0.5);
	}

	@media (prefers-reduced-motion: reduce) {
		.segmented button {
			transition: none;
		}
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
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
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

	/* La fecha va en su propia fila: en la de las cifras no cabía y el año se
	   salía por la derecha del diálogo. */
	.form-row.meta {
		grid-template-columns: auto minmax(0, 1fr);
	}

	@media (max-width: 640px) {
		.form-row.meta {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
