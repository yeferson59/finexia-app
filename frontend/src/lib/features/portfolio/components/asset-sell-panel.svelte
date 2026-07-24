<script lang="ts">
	import { enhance } from '$app/forms';
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { formatCalendarDate, todayLocalDateString } from '$lib/utils';
	import type { Holding, Transaction } from '$lib/api/types';

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
	let sellDate = $state(todayLocalDateString());
	let sellNotes = $state('');
	let isSellSubmitting = $state(false);
	let sellPanelEl = $state<HTMLElement | null>(null);

	$effect(() => {
		if (transaction) {
			sellMode = 'full';
			sellBasis = 'quantity';
			sellQty = transaction.quantity;
			sellValue = '';
			sellPrice = marketPrice ? marketPrice.toFixed(2) : transaction.price;
			sellFees = '';
			sellNotes = '';
			sellDate = todayLocalDateString();
			setTimeout(() => sellPanelEl?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 50);
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
</script>

<div class="sell-panel" bind:this={sellPanelEl}>
	<div class="sell-panel-header">
		<div class="sell-panel-info">
			<span class="sell-panel-title">Vender desde compra</span>
			<span class="sell-panel-lot">
				Lote: {parseFloat(transaction.quantity).toLocaleString('es-CO', {
					maximumFractionDigits: 8
				})}
				unidades @ {formatCurrency(parseFloat(transaction.price))} ·
				{formatCalendarDate(transaction.transactionDate, {
					year: 'numeric',
					month: 'short',
					day: 'numeric'
				})}
			</span>
		</div>
		<button class="sell-panel-close" type="button" onclick={onClose}>✕</button>
	</div>

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
		<input
			type="hidden"
			name="currency"
			value={entries.find((e) => e.id === transaction.entryId)?.costCurrency ?? fallbackCurrency}
		/>
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
					step="0.01"
					required
				/>
			</div>
			<div class="form-group">
				<label class="form-label" for="sell-fees">Comisión</label>
				<input
					id="sell-fees"
					type="number"
					class="form-input"
					name="fees"
					bind:value={sellFees}
					placeholder="0"
					min="0"
					step="0.01"
				/>
			</div>
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
	.sell-panel {
		margin-bottom: 1.5rem;
		padding: 1.25rem;
		border: 1px solid rgba(224, 90, 90, 0.3);
		border-radius: 10px;
		background: rgba(224, 90, 90, 0.05);
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.sell-panel-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.sell-panel-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.sell-panel-title {
		font-size: 0.9rem;
		font-weight: 700;
		color: var(--red);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.sell-panel-lot {
		font-size: 0.82rem;
		color: rgba(236, 234, 229, 0.6);
		font-family: var(--font-mono);
	}

	.sell-panel-close {
		padding: 0.2rem 0.5rem;
		border: none;
		background: transparent;
		color: rgba(236, 234, 229, 0.4);
		font-size: 1rem;
		cursor: pointer;
		border-radius: 4px;
		transition: color 0.2s ease;
		flex-shrink: 0;
	}

	.sell-panel-close:hover {
		color: var(--text);
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
		color: rgba(236, 234, 229, 0.4);
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
