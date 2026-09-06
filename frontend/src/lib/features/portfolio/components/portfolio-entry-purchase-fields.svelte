<script lang="ts">
	/*
	 * Lo que pagaste: cuántas unidades, a qué precio, en qué moneda y qué día.
	 *
	 * Era una tarjeta con sombra propia y el total metido dentro, en un recuadro
	 * que imitaba un campo de formulario pero no se podía escribir. El total ha
	 * salido de aquí: lo pinta `portfolio-entry-total`, solo, al cierre del
	 * bloque, porque es lo que hay que mirar antes de guardar.
	 */
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
		fxRate = $bindable()
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
</script>

<div class="pair">
	<div class="field">
		<label for="quantity">Cantidad</label>
		<input
			id="quantity"
			type="number"
			name="quantity"
			bind:value={quantity}
			placeholder="10"
			min="0"
			step="any"
			required
		/>
		<p class="hint">Cuántas unidades compraste.</p>
	</div>

	<div class="field">
		<label for="purchasePrice">Precio por unidad</label>
		<!-- El selector ocupa el sitio del símbolo de dólar fijo que había aquí:
		     es el mismo hueco, pero ahora dice la verdad sobre el precio. -->
		<div class="price">
			<select
				id="currency"
				name="currency"
				bind:value={currency}
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
				min="0"
				step="any"
				required
			/>
		</div>
		{#if asset && assetCurrency && currency !== assetCurrency}
			<p class="warning">
				{asset.ticker} cotiza en {assetCurrency}. Copia el precio tal como lo muestra tu bróker, en
				la moneda en la que se ejecutó.
			</p>
		{:else}
			<p class="hint">Lo que costó cada una, en {currency}.</p>
		{/if}
	</div>
</div>

<div class="field">
	<span class="field-label">Fecha de compra</span>
	<DatePicker name="purchaseDate" bind:value={purchaseDate} required />
</div>

<div class="field">
	<label class="settlement">
		<input type="checkbox" bind:checked={converted} />
		<span class="text">
			<span class="name">Mi cuenta liquidó en otra moneda</span>
			<span class="description">
				Márcalo si el activo cotiza en una moneda y tu cuenta está en otra: el bróker convirtió a
				una tasa que forma parte de lo que te costó la posición.
			</span>
		</span>
	</label>
</div>

{#if converted}
	{#if sameCurrency}
		<!-- La tasa igual va oculta en 1: mientras las dos monedas coincidan no
		     hubo conversión, y mandar la que esté escrita solo produce el 400. -->
		<input type="hidden" name="fxRate" value="1" />
	{/if}
	<div class="pair">
		<div class="field">
			<label for="costCurrency">Moneda de la cuenta</label>
			<select id="costCurrency" name="costCurrency" bind:value={costCurrency}>
				{#each currencyOptions as code (code)}
					<option value={code}>{code}</option>
				{/each}
			</select>
			<p class="hint">En la que el bróker debitó el importe.</p>
		</div>

		<div class="field">
			<label for="fxRate">Tasa de la operación</label>
			<input
				id="fxRate"
				type="number"
				name={sameCurrency ? 'fxRateIgnored' : 'fxRate'}
				bind:value={fxRate}
				placeholder="1.0638"
				min="0"
				step="any"
				disabled={sameCurrency}
				required={converted && !sameCurrency}
			/>
			{#if sameCurrency}
				<p class="warning">
					La operación y la cuenta están las dos en {currency}, así que no hubo conversión y no hay
					tasa que aplicar. Si el activo cotizó en otra moneda, cámbiala arriba, en el selector que
					está junto al precio.
				</p>
			{:else}
				<p class="hint">
					Cuántos {costCurrency} costaba 1 {currency} ese día. Cópiala de la confirmación del bróker,
					no la de hoy: la de hoy convierte la compra a un precio que nunca pagaste.
				</p>
			{/if}
		</div>
	</div>
{:else}
	<!-- Sin conversión las dos monedas son la misma y la tasa es 1. Van en campos
	     ocultos para que el cuerpo enviado sea idéntico en los dos casos y el
	     servidor no tenga que adivinar cuál falta. -->
	<input type="hidden" name="costCurrency" value={currency} />
	<input type="hidden" name="fxRate" value="1" />
{/if}

<style>
	/* La moneda pegada al precio: son un solo dato partido en dos controles. */
	.price {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		gap: 0.5rem;
	}

	.price select {
		font-family: var(--font-mono);
		font-size: 0.85rem;
	}

	.price input {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	/* Aviso en el tono de lo que dice, sin caja: el mismo idioma que el resto
	   del panel. */
	.warning {
		margin: 0;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--amber);
	}

	/* Como las opciones de correo de notificaciones: la casilla delante y la
	   fila entera de etiqueta. */
	.settlement {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: start;
		gap: 0.9rem;
		cursor: pointer;
	}

	.settlement input[type='checkbox'] {
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
		max-width: 52ch;
		margin-top: 0.2rem;
		font-size: 0.8rem;
		line-height: 1.55;
		color: var(--text-muted);
	}
</style>
