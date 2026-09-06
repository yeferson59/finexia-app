<script lang="ts">
	/*
	 * Los dos ajustes sueltos del alta: la meta y si es el portafolio por
	 * defecto.
	 *
	 * Era una tarjeta con leyenda propia, «OBJETIVO FINANCIERO», para dos
	 * controles. Ahora son dos campos dentro del bloque que los contiene: el
	 * marco lo pone `portfolio-form-section`.
	 */
	let {
		currency,
		targetAmount = $bindable(''),
		isDefault = $bindable(false),
		disabled = false
	}: {
		currency: string;
		targetAmount?: string;
		isDefault?: boolean;
		disabled?: boolean;
	} = $props();
</script>

<div class="field">
	<label for="targetAmount">Meta <span class="optional">(opcional)</span></label>
	<div class="amount">
		<!-- La moneda va dentro del campo, en gris y sin filete ni fondo propios:
		     antes era una pastilla ámbar pegada al borde izquierdo que pesaba más
		     que la cifra que la acompaña. -->
		<span class="currency" aria-hidden="true">{currency}</span>
		<input
			type="number"
			id="targetAmount"
			name="priceValue"
			bind:value={targetAmount}
			placeholder="0,00"
			step="0.01"
			min="0"
			{disabled}
		/>
	</div>
	<p class="hint">Lo que te gustaría llegar a tener en este portafolio, en {currency}.</p>
</div>

<label class="default">
	<input type="checkbox" id="isDefault" name="isDefault" bind:checked={isDefault} {disabled} />
	<span class="text">
		<span class="name">Usar este portafolio por defecto</span>
		<span class="description">Vendrá elegido de antemano cuando registres un movimiento.</span>
	</span>
</label>

<style>
	.amount {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding-left: 0.95rem;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.03);
		transition: border-color 0.2s ease;
	}

	.amount:hover {
		border-color: rgba(212, 145, 42, 0.35);
	}

	.amount:focus-within {
		border-color: var(--amber);
	}

	.currency {
		font-family: var(--font-mono);
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	/*
	 * El borde y el fondo los pone el contenedor, para que la moneda y la cifra
	 * se lean como un solo campo. Con el selector de atributo, no con `input` a
	 * secas: el bloque que envuelve el formulario pinta los campos con
	 * `:global(input[type='number'])`, que gana a un selector de elemento y
	 * dejaba aquí una caja dentro de otra.
	 */
	.amount input[type='number'] {
		border: none;
		border-radius: 0;
		background: none;
		padding-left: 0;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.amount input[type='number']:focus,
	.amount input[type='number']:hover:not(:disabled) {
		border: none;
	}

	.default {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: start;
		gap: 0.9rem;
		cursor: pointer;
	}

	.default input[type='checkbox'] {
		width: 18px;
		height: 18px;
		margin: 0.1rem 0 0;
		cursor: pointer;
	}

	.text {
		min-width: 0;
	}

	.name {
		display: block;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
	}

	.description {
		display: block;
		max-width: 46ch;
		margin-top: 0.2rem;
		font-size: 0.8rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	@media (prefers-reduced-motion: reduce) {
		.amount {
			transition: none;
		}
	}
</style>
