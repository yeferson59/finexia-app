<script lang="ts">
	/*
	 * Lo que te costó la posición, y la cuenta de la que sale.
	 *
	 * Es lo único que esta pantalla pide mirar con atención: es el número que se
	 * contrasta contra la confirmación del bróker antes de guardar. Por eso está
	 * solo, y por eso enseña la operación —«5 × $100,00»— en vez de pedir que te
	 * fíes de él.
	 *
	 * Sustituye a dos cosas que decían lo mismo: el recuadro «Valor Total
	 * Invertido» de los detalles de compra y la tarjeta ámbar «Resumen de
	 * Inversión» del final, que repetía activo, cantidad y precio unitario ya
	 * escritos veinte píxeles más arriba.
	 */
	let {
		units,
		unitPrice,
		currency,
		costCurrency,
		rate,
		converted,
		formatCurrency
	}: {
		units: number;
		unitPrice: number;
		/** Moneda en la que cotizó la operación. */
		currency: string;
		/** Moneda en la que liquidó la cuenta. */
		costCurrency: string;
		rate: number;
		converted: boolean;
		formatCurrency: (value: number, code: string) => string;
	} = $props();

	const traded = $derived(units * unitPrice);
	const settled = $derived(traded * rate);
	const priced = $derived(units > 0 && unitPrice > 0);
	// Con conversión declarada y sin tasa, el total sería cero: un número
	// plausible y falso en el sitio donde más se mira.
	const ready = $derived(priced && rate > 0);

	const pendingReason = $derived(
		priced ? 'Falta la tasa de la operación.' : 'Escribe la cantidad y el precio y aparecerá aquí.'
	);
</script>

<div class="total">
	<span class="label">Total invertido</span>

	<span class="figure" class:pending={!ready}>
		{ready ? formatCurrency(settled, costCurrency || currency) : '—'}
	</span>

	{#if ready}
		<span class="working">
			{units.toLocaleString('es-CO', { maximumFractionDigits: 8 })} × {formatCurrency(
				unitPrice,
				currency
			)}
			<!-- Con conversión, la cuenta entera: es lo que permite ver si el que no
			     cuadra es el precio o la tasa, sin salir de la pantalla. -->
			{#if converted}
				<span class="fx">
					= {formatCurrency(traded, currency)} al cambio de {rate.toLocaleString('es-CO', {
						maximumFractionDigits: 6
					})}
				</span>
			{/if}
		</span>
	{:else}
		<span class="working">{pendingReason}</span>
	{/if}
</div>

<style>
	.total {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: baseline;
		gap: 0.35rem 1.5rem;
		padding-top: 1.35rem;
		border-top: 1px solid var(--border-strong);
	}

	.label {
		grid-column: 1;
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
	}

	/*
	 * El único ámbar de la página. Estaba repartido entre este total, el resumen,
	 * los cinco datos del activo, los tickers del buscador y los bordes de todos
	 * los campos, así que no destacaba nada.
	 */
	.figure {
		grid-column: 2;
		font-family: var(--font-mono);
		font-size: 1.5rem;
		font-variant-numeric: tabular-nums;
		font-weight: 600;
		color: var(--amber);
		white-space: nowrap;
	}

	.figure.pending {
		color: var(--text-dim);
	}

	.working {
		grid-column: 1 / -1;
		justify-self: end;
		max-width: 46ch;
		font-family: var(--font-mono);
		font-size: 0.78rem;
		font-variant-numeric: tabular-nums;
		line-height: 1.5;
		color: var(--text-muted);
		text-align: right;
	}

	.fx {
		display: block;
	}

	/* Sin la tipografía de máquina: es una frase, no una cuenta. */
	.figure.pending ~ .working {
		font-family: var(--font-body);
	}

	@media (max-width: 520px) {
		.figure {
			font-size: 1.25rem;
		}

		.working {
			justify-self: start;
			text-align: left;
		}
	}
</style>
