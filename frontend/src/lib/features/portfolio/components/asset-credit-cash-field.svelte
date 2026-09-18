<script lang="ts">
	/**
	 * La casilla que abona un dividendo al efectivo de la plataforma, en el alta y
	 * en la edición de una transacción.
	 *
	 * En el alta viene marcada porque es lo que hace el bróker: el dividendo llega
	 * a la cuenta. Se desmarca si ese dinero ya se anotó como un depósito o llegó
	 * a otra cuenta. En la edición arranca como está el dividendo, y desmarcarla
	 * lo saca del efectivo.
	 */
	let {
		checked = $bindable(true),
		currency,
		amount = '',
		editing = false
	}: {
		checked?: boolean;
		/** La moneda de la cuenta, que es en la que se abona. */
		currency: string;
		/** Lo que se abonará, ya con formato; vacío mientras no se sepa. */
		amount?: string;
		editing?: boolean;
	} = $props();
</script>

<label class="credit-cash">
	<input type="checkbox" name="creditCash" bind:checked />
	<span class="text">
		<span class="name">Abonar al efectivo de la plataforma</span>
		<span class="hint">
			{#if editing}
				El dividendo suma a tu efectivo en {currency} en esta plataforma. Si lo desmarcas, sale de ahí.
			{:else}
				Suma {amount || 'el dividendo'} a tu efectivo en {currency} en esta plataforma. Desmárcalo si
				ya lo anotaste como un depósito o si te lo pagaron en otra cuenta.
			{/if}
		</span>
	</span>
</label>

<style>
	.credit-cash {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: start;
		gap: 0.75rem;
		cursor: pointer;
	}

	.credit-cash input[type='checkbox'] {
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
</style>
