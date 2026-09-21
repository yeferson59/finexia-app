<script lang="ts">
	/**
	 * La casilla que paga una compra con el efectivo que la plataforma ya tiene,
	 * en el alta de una posición, en el alta de una transacción y en la edición.
	 *
	 * Es el espejo de `asset-credit-cash-field`: allí el dinero entra, aquí sale.
	 * Y al revés que allí, viene desmarcada. Abonar una venta es lo que hace el
	 * bróker siempre; pagar con el saldo de la cuenta solo pasa cuando el dinero
	 * ya estaba ahí, y marcarlo por defecto descuadraría el saldo de quien
	 * transfiere el dinero para cada compra.
	 *
	 * Sin saldo en esa moneda la casilla no se dibuja: no hay nada con qué pagar,
	 * y el backend rechazaría la compra entera. Quien la monta decide eso.
	 */
	let {
		checked = $bindable(false),
		currency,
		amount = '',
		balance = '',
		enough = true,
		editing = false
	}: {
		checked?: boolean;
		/** La moneda de la cuenta, que es en la que se paga. */
		currency: string;
		/** Lo que saldrá del saldo, ya con formato; vacío mientras no se sepa. */
		amount?: string;
		/** Lo que hay en esa cuenta, ya con formato. */
		balance?: string;
		/** Si lo que hay alcanza para lo que va a salir. */
		enough?: boolean;
		editing?: boolean;
	} = $props();
</script>

<label class="pay-cash">
	<input type="checkbox" name="payFromCash" bind:checked />
	<span class="text">
		<span class="name">Pagarlo con mi efectivo en esta plataforma</span>
		<span class="hint">
			{#if editing}
				Resta {amount || 'lo que costó'} de tu efectivo en {currency} en esta plataforma. Si lo desmarcas,
				el dinero vuelve al saldo.
			{:else}
				Resta {amount || 'lo que cuesta'} de tu efectivo en {currency} en esta plataforma. Déjalo sin
				marcar si transferiste el dinero para esta compra.
			{/if}
		</span>
		{#if balance}
			<span class="balance" class:short={checked && !enough}>
				Tienes {balance} en esta cuenta.
				{#if checked && !enough}
					No alcanza para esta compra.
				{/if}
			</span>
		{/if}
	</span>
</label>

<style>
	.pay-cash {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: start;
		gap: 0.75rem;
		cursor: pointer;
	}

	.pay-cash input[type='checkbox'] {
		width: 18px;
		height: 18px;
		margin: 0.1rem 0 0;
		accent-color: var(--amber);
		cursor: pointer;
	}

	.text {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		min-width: 0;
	}

	.name {
		font-size: 0.87rem;
		font-weight: 500;
		color: var(--text);
	}

	.balance {
		font-size: 0.8rem;
		color: var(--text-muted);
		font-variant-numeric: tabular-nums;
	}

	/* El aviso no bloquea el envío: el saldo puede haber cambiado en otra
	   pestaña, y quien decide es el backend. Solo dice lo que va a pasar. */
	.balance.short {
		color: var(--red);
	}
</style>
