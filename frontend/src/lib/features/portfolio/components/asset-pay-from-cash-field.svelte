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
	 * Con más de un sitio de donde sacarlo se elige cuál. La mayoría del dinero
	 * que alguien guarda está en un bolsillo —la cajita del banco, el subsaldo
	 * del bróker—, así que ofrecer solo la cuenta principal sería no ofrecer
	 * nada. Los depósitos a plazo no están en la lista: están cerrados hasta que
	 * vencen, y el backend los rechaza igual.
	 *
	 * Sin ningún saldo la casilla no se dibuja: no hay nada con qué pagar, y el
	 * backend rechazaría la compra entera. Quien la monta decide eso.
	 */
	import { formatSettled, type CashSource } from '../asset';

	let {
		checked = $bindable(false),
		pocketId = $bindable(''),
		currency,
		amount = 0,
		sources,
		paidFromPocketId = null,
		paidAmount = 0,
		editing = false
	}: {
		checked?: boolean;
		/** El bolsillo elegido; vacío es la cuenta principal. */
		pocketId?: string;
		/** La moneda de la cuenta, que es en la que se paga. */
		currency: string;
		/** Lo que va a salir del saldo. */
		amount?: number;
		sources: CashSource[];
		/**
		 * Dónde está ya el dinero de esta compra, si ya se pagaba con efectivo, y
		 * cuánto: ese saldo se enseña con el cargo restado, así que para saber si
		 * alcanza hay que volver a sumarlo.
		 */
		paidFromPocketId?: string | null;
		paidAmount?: number;
		editing?: boolean;
	} = $props();

	const selected = $derived(sources.find((s) => s.id === pocketId) ?? sources[0]);
	const available = $derived(selected?.balance ?? 0);
	const backIn = $derived(
		paidFromPocketId !== null && paidFromPocketId === selected?.id ? paidAmount : 0
	);
	const enough = $derived(amount <= available + backIn);
</script>

<div class="pay-cash">
	<label class="head">
		<input type="checkbox" name="payFromCash" bind:checked />
		<span class="text">
			<span class="name">Pagarlo con mi efectivo en esta plataforma</span>
			<span class="hint">
				{#if editing}
					Resta {amount > 0 ? formatSettled(amount, currency) : 'lo que costó'} de tu efectivo en {currency}
					en esta plataforma. Si lo desmarcas, el dinero vuelve al saldo.
				{:else}
					Resta {amount > 0 ? formatSettled(amount, currency) : 'lo que cuesta'} de tu efectivo en {currency}
					en esta plataforma. Déjalo sin marcar si transferiste el dinero para esta compra.
				{/if}
			</span>
		</span>
	</label>

	{#if checked && sources.length > 1}
		<div class="from">
			<label class="from-label" for="pay-from-pocket">De dónde sale</label>
			<select id="pay-from-pocket" class="form-select" name="payFromPocketId" bind:value={pocketId}>
				{#each sources as source (source.id)}
					<option value={source.id}>
						{source.name} · {formatSettled(source.balance, currency)}
					</option>
				{/each}
			</select>
		</div>
	{:else if checked}
		<input type="hidden" name="payFromPocketId" value={pocketId} />
	{/if}

	{#if selected}
		<p class="balance" class:short={checked && !enough}>
			{#if sources.length > 1}
				En {selected.name} tienes {formatSettled(available, currency)}.
			{:else}
				Tienes {formatSettled(available, currency)} en esta cuenta.
			{/if}
			{#if checked && !enough}
				No alcanza para esta compra.
			{/if}
		</p>
	{/if}
</div>

<style>
	.pay-cash {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.head {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: start;
		gap: 0.75rem;
		cursor: pointer;
	}

	.head input[type='checkbox'] {
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

	/* Sangrado bajo la casilla, para que se lea como parte de ella y no como
	   otro campo del formulario. */
	.from,
	.balance {
		margin: 0;
		padding-left: calc(18px + 0.75rem);
	}

	.from {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.from-label {
		font-size: 0.8rem;
		color: var(--text-muted);
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

	.form-select {
		padding: 0.5rem 2.2rem 0.5rem 0.75rem;
		border: 1.5px solid rgba(212, 145, 42, 0.35);
		border-radius: 8px;
		background-color: rgba(212, 145, 42, 0.06);
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%23d4912a' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M6 9l6 6 6-6'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.65rem center;
		background-size: 1rem;
		appearance: none;
		-webkit-appearance: none;
		color: var(--text);
		font-size: 0.85rem;
		font-family: var(--font-body);
		cursor: pointer;
	}

	.form-select:focus {
		outline: none;
		border-color: var(--amber);
	}
</style>
