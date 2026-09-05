<script lang="ts">
	import DatePicker from '$lib/ui/date-picker.svelte';
	import { SUPPORTED_CURRENCIES } from '$lib/shared/currency';
	import type { Asset } from '$lib/api/types';

	let {
		asset,
		quantity = $bindable(),
		purchasePrice = $bindable(),
		purchaseDate = $bindable(),
		currency = $bindable(),
		costCurrency = $bindable(),
		fxRate = $bindable(),
		formatCurrency
	}: {
		asset: Asset | null;
		quantity: string;
		purchasePrice: string;
		purchaseDate: string;
		/** Moneda en la que cotizó la operación, que el padre siembra del activo. */
		currency: string;
		/** Moneda en la que la cuenta liquidó, y en la que queda el coste. */
		costCurrency: string;
		/** Cuánto valía una unidad de `currency` en `costCurrency` ese día. */
		fxRate: string;
		formatCurrency: (value: number, code: string) => string;
	} = $props();

	// La moneda del activo puede no estar en la lista de conversión (DKK, por
	// ejemplo): se ofrece igual, porque es la que hace correcto el coste.
	const currencyOptions = $derived.by(() => {
		const options: string[] = [...SUPPORTED_CURRENCIES];
		for (const code of [currency, costCurrency]) {
			if (code && !options.includes(code)) options.unshift(code);
		}

		return options;
	});

	const assetCurrency = $derived(asset?.currency?.trim().toUpperCase() ?? '');

	/**
	 * Si la cuenta liquidó en otra moneda.
	 *
	 * Es un interruptor y no dos selectores siempre visibles porque el caso
	 * normal —cuenta y activo en la misma moneda— no tiene nada que decidir, y
	 * un campo de tasa a la vista invita a rellenarlo con la tasa de hoy, que
	 * para una compra vieja es justo el número equivocado.
	 */
	let converted = $state(false);

	// Apagarlo tiene que deshacer lo que encendió: si quedara una moneda de
	// liquidación distinta con tasa 1, el backend guardaría el precio cotizado
	// etiquetado con la moneda de la cuenta, que es el error original.
	$effect(() => {
		if (!converted) {
			costCurrency = currency;
			fxRate = '';
		}
	});

	/**
	 * El interruptor está encendido pero las dos monedas son la misma.
	 *
	 * Es un estado alcanzable —basta no tocar el selector de la izquierda— y no
	 * es una conversión: una moneda no se convierte en sí misma a 1,0638. El
	 * backend lo rechaza con un 400, así que el formulario lo dice antes, junto
	 * al campo que hay que cambiar, y manda tasa 1 en vez de una que no puede
	 * ser cierta.
	 */
	const sameCurrency = $derived(converted && costCurrency === currency);

	const rate = $derived(parseFloat(fxRate) || (converted ? 0 : 1));
	const units = $derived(parseFloat(quantity) || 0);
	const unitPrice = $derived(parseFloat(purchasePrice) || 0);

	const tradedTotal = $derived(units * unitPrice);
	const settledTotal = $derived(tradedTotal * rate);
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
					id="currency"
					name="currency"
					bind:value={currency}
					class="currency-select"
					aria-label="Moneda de la operación"
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
					step="any"
					required
				/>
			</div>
			{#if asset && assetCurrency && currency !== assetCurrency}
				<p class="field-warning">
					{asset.ticker} cotiza en {assetCurrency}. Copia el precio tal como lo muestra tu bróker,
					en la moneda en la que se ejecutó.
				</p>
			{:else}
				<p class="field-hint">Precio por unidad en {currency}</p>
			{/if}
		</div>
	</div>

	<div class="form-group">
		<label class="settlement-toggle">
			<input type="checkbox" bind:checked={converted} />
			<span>Mi cuenta liquidó en otra moneda</span>
		</label>
		<p class="field-hint">
			Actívalo si el activo cotiza en una moneda y tu cuenta está en otra: el bróker convirtió a una
			tasa que forma parte de lo que te costó la posición.
		</p>
	</div>

	{#if converted}
		{#if sameCurrency}
			<!-- La tasa igual va oculta en 1: mientras las dos monedas coincidan no
			     hubo conversión, y mandar la que esté escrita solo produce el 400. -->
			<input type="hidden" name="fxRate" value="1" />
		{/if}
		<div class="form-row">
			<div class="form-group">
				<label for="costCurrency" class="form-label"
					>Moneda de la cuenta <span class="required">*</span></label
				>
				<select
					id="costCurrency"
					name="costCurrency"
					bind:value={costCurrency}
					class="currency-select"
				>
					{#each currencyOptions as code (code)}
						<option value={code}>{code}</option>
					{/each}
				</select>
				<p class="field-hint">En la que el bróker debitó el importe</p>
			</div>

			<div class="form-group">
				<label for="fxRate" class="form-label"
					>Tasa de la operación <span class="required">*</span></label
				>
				<input
					id="fxRate"
					type="number"
					name={sameCurrency ? 'fxRateIgnored' : 'fxRate'}
					bind:value={fxRate}
					placeholder="1.0638"
					class="form-input"
					min="0"
					step="any"
					disabled={sameCurrency}
					required={converted && !sameCurrency}
				/>
				{#if sameCurrency}
					<p class="field-warning">
						La operación y la cuenta están las dos en {currency}, así que no hubo conversión y no
						hay tasa que aplicar. Si el activo cotizó en otra moneda, cámbiala arriba, en el
						selector que está junto al precio.
					</p>
				{:else}
					<p class="field-hint">
						Cuántos {costCurrency} costaba 1 {currency} ese día. Cópiala de la confirmación del bróker,
						no la de hoy: la de hoy convierte la compra a un precio que nunca pagaste.
					</p>
				{/if}
			</div>
		</div>
	{:else}
		<!-- Sin conversión las dos monedas son la misma y la tasa es 1. Van en
		     campos ocultos para que el cuerpo enviado sea idéntico en los dos
		     casos y el servidor no tenga que adivinar cuál falta. -->
		<input type="hidden" name="costCurrency" value={currency} />
		<input type="hidden" name="fxRate" value="1" />
	{/if}

	<div class="form-row">
		<div class="form-group">
			<span class="form-label">Fecha de Compra</span>
			<DatePicker name="purchaseDate" bind:value={purchaseDate} required />
		</div>

		<div class="form-group">
			<span class="form-label">Valor Total Invertido</span>
			<div class="value-display">
				<p class="total-value">{formatCurrency(settledTotal, costCurrency || currency)}</p>
			</div>
			{#if converted && !sameCurrency}
				<!-- El número que el usuario puede contrastar contra la pantalla de
				     su bróker: si no coincide con el «open value» de allí, la tasa
				     o el precio están mal, y se ve antes de guardar. -->
				<p class="field-hint">
					{formatCurrency(tradedTotal, currency)} × {fxRate || '—'} = {formatCurrency(
						settledTotal,
						costCurrency
					)}
				</p>
			{/if}
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
		color: var(--text-dim);
		font-style: italic;
	}

	.field-warning {
		margin: 0.4rem 0 0;
		font-size: 0.8rem;
		color: rgba(212, 145, 42, 0.85);
	}

	.settlement-toggle {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text);
		cursor: pointer;
	}

	.settlement-toggle input {
		width: 1rem;
		height: 1rem;
		accent-color: var(--amber);
		cursor: pointer;
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
