<script lang="ts">
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';
	import type { Asset } from '$lib/api/types';

	let {
		asset,
		quantity = $bindable(),
		purchasePrice = $bindable(),
		purchaseDate = $bindable(),
		costCurrency = $bindable(),
		totalValue,
		formatCurrency
	}: {
		asset: Asset | null;
		quantity: string;
		purchasePrice: string;
		purchaseDate: string;
		/** Moneda en la que se pagó, que el padre siembra desde el activo. */
		costCurrency: string;
		totalValue: number;
		formatCurrency: (value: number) => string;
	} = $props();

	// La moneda del activo puede no estar en la lista de conversión (DKK, por
	// ejemplo): se ofrece igual, porque es la que hace correcto el coste.
	const currencyOptions = $derived.by(() => {
		const options: string[] = [...SUPPORTED_CURRENCIES];
		if (costCurrency && !options.includes(costCurrency)) options.unshift(costCurrency);

		return options;
	});

	const assetCurrency = $derived(asset?.currency?.trim().toUpperCase() ?? '');
</script>

<section class="form-section">
	<h2 class="section-title">Detalles de Compra</h2>

	<div class="form-row">
		<div class="form-group">
			<label for="quantity" class="form-label">Cantidad <span class="required">*</span></label>
			<div class="input-addon">
				<input
					id="quantity"
					type="number"
					name="quantity"
					bind:value={quantity}
					placeholder="1000"
					class="form-input"
					min="0"
					step="any"
					required
				/>
			</div>
			<p class="field-hint">Número de unidades</p>
		</div>

		<div class="form-group">
			<label for="purchasePrice" class="form-label"
				>Precio de Compra <span class="required">*</span></label
			>
			<!-- El selector ocupa el sitio del símbolo de dólar fijo que había aquí:
			     es el mismo hueco, pero ahora dice la verdad sobre el precio. -->
			<div class="price-field">
				<select
					id="costCurrency"
					name="costCurrency"
					bind:value={costCurrency}
					class="currency-select"
					aria-label="Moneda de la compra"
				>
					{#each currencyOptions as code (code)}
						<option value={code}>{code}</option>
					{/each}
				</select>
				<input
					id="purchasePrice"
					type="number"
					name="purchasePrice"
					bind:value={purchasePrice}
					placeholder="150.50"
					class="form-input"
					min="0"
					step="0.01"
					required
				/>
			</div>
			{#if asset && assetCurrency && costCurrency !== assetCurrency}
				<p class="field-warning">
					{asset.ticker} cotiza en {assetCurrency}. Elige {costCurrency} solo si tu bróker liquidó la
					compra en esa moneda.
				</p>
			{:else}
				<p class="field-hint">Precio por unidad en {costCurrency}</p>
			{/if}
		</div>
	</div>

	<div class="form-row">
		<div class="form-group">
			<span class="form-label">Fecha de Compra</span>
			<DatePicker name="purchaseDate" bind:value={purchaseDate} required />
		</div>

		<div class="form-group">
			<span class="form-label">Valor Total Invertido</span>
			<div class="value-display">
				<p class="total-value">{formatCurrency(totalValue)}</p>
			</div>
		</div>
	</div>
</section>

<style>
	.form-section {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		padding: 1.75rem;
	}

	.section-title {
		margin: 0 0 1.5rem;
		font-size: 1.15rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
		margin-bottom: 1.35rem;
	}

	.form-group:last-child {
		margin-bottom: 0;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
	}

	.form-label {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text);
		letter-spacing: 0.3px;
	}

	.required {
		color: var(--red);
	}

	.form-input {
		padding: 0.85rem 1rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		color: var(--text);
		font-size: 0.95rem;
		font-family: var(--font-body);
		transition: all 0.3s ease;
	}

	.form-input::placeholder {
		color: rgba(236, 234, 229, 0.55);
	}

	.form-input:focus {
		outline: none;
		border-color: var(--amber);
		background: rgba(255, 255, 255, 0.022);
		box-shadow: 0 0 0 3px var(--border);
	}

	.field-hint {
		margin: 0.4rem 0 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.4);
		font-style: italic;
	}

	.field-warning {
		margin: 0.4rem 0 0;
		font-size: 0.8rem;
		color: rgba(212, 145, 42, 0.85);
	}

	.input-addon {
		position: relative;
		display: flex;
		align-items: center;
	}

	.input-addon .form-input {
		padding-left: 2.5rem;
	}

	.price-field {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 0.5rem;
		align-items: center;
	}

	.currency-select {
		padding: 0.85rem 0.75rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		color: var(--text);
		font-size: 0.9rem;
		font-weight: 600;
		font-family: var(--font-body);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.currency-select:focus {
		outline: none;
		border-color: var(--amber);
		box-shadow: 0 0 0 3px var(--border);
	}

	.value-display {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		padding: 0.85rem 1rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 44px;
	}

	.total-value {
		font-variant-numeric: tabular-nums;
		margin: 0;
		font-size: 1.2rem;
		font-weight: 700;
		color: var(--amber);
		font-family: var(--font-body);
	}

	@media (max-width: 768px) {
		.form-row {
			grid-template-columns: 1fr;
		}
	}
</style>
