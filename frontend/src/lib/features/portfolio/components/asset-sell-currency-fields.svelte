<script lang="ts">
	/**
	 * La tasa y la comisión de una venta, con sus monedas.
	 *
	 * Vive fuera del panel porque las tres cosas solo tienen sentido juntas: el
	 * precio se cotizó en la moneda del mercado, el importe llegó a la cuenta en
	 * otra, y la comisión pudo cobrarse de cualquiera de las dos. Cuando la
	 * posición es de una sola moneda no hay nada que decidir y el bloque se
	 * reduce al campo de comisión de siempre.
	 */
	let {
		crossCurrency,
		tradeCurrency,
		costCurrency,
		proceeds,
		rate = $bindable(),
		fees = $bindable(),
		feesCurrency = $bindable()
	}: {
		crossCurrency: boolean;
		/** Moneda en la que cotiza el activo, y en la que va el precio unitario. */
		tradeCurrency: string;
		/** Moneda de la posición: en la que se recibe el importe. */
		costCurrency: string;
		/** Lo que la cuenta recibiría con lo tecleado, para contrastarlo. */
		proceeds: number;
		rate: string;
		fees: string;
		feesCurrency: string;
	} = $props();
</script>

{#if crossCurrency}
	<div class="form-group">
		<label class="form-label" for="sell-rate"
			>Tasa a {costCurrency} <span class="required">*</span></label
		>
		<input
			id="sell-rate"
			type="number"
			class="form-input"
			name="fxRate"
			bind:value={rate}
			placeholder="1.1565"
			min="0"
			step="any"
			required
		/>
		<span class="sell-computed-hint">
			≈ {proceeds.toLocaleString('es-CO', {
				style: 'currency',
				currency: costCurrency,
				maximumFractionDigits: 2
			})} recibidos
		</span>
	</div>
{/if}

<div class="form-group">
	<label class="form-label" for="sell-fees">Comisión</label>
	<input
		id="sell-fees"
		type="number"
		class="form-input"
		name="fees"
		bind:value={fees}
		placeholder="0"
		min="0"
		step="any"
	/>
	{#if crossCurrency}
		<select
			class="form-input"
			name="feesCurrency"
			bind:value={feesCurrency}
			aria-label="Moneda de la comisión"
		>
			<option value={costCurrency}>{costCurrency}</option>
			<option value={tradeCurrency}>{tradeCurrency}</option>
		</select>
	{/if}
</div>

<style>
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

	.sell-computed-hint {
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.5);
		font-variant-numeric: tabular-nums;
	}
</style>
