<script lang="ts">
	/*
	 * Cuánto tienes de este activo aquí, y cómo le va.
	 *
	 * Ocupa el sitio de doce tarjetas —seis de «Resumen de Posición» y seis de
	 * «Información del Activo»— de las que la mitad repetía algo que ya estaba
	 * en la misma pantalla: «precio actual» era el precio gigante de la
	 * cabecera, «ROI» era el porcentaje de «ganancia/pérdida», «tipo» y
	 * «exchange» eran la insignia de al lado, «moneda» era la línea que
	 * repetían debajo otras cuatro tarjetas, y «transacciones» era el contador
	 * del historial de abajo.
	 *
	 * La barra es la misma del listado y del detalle del portafolio: lo que
	 * queda a la izquierda del corte es el capital y lo que sigue, la ganancia.
	 *
	 * Las tres monedas de la ficha —la de coste, la de cotización y la base del
	 * portafolio— las lleva ahora el símbolo de cada importe en vez de una
	 * línea `USD` bajo cada tarjeta. Solo se nombran cuando de verdad difieren,
	 * porque entonces restar un precio de otro deja de tener sentido y hay que
	 * decirlo.
	 */
	import PortfolioCapitalBar from './portfolio-capital-bar.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPercent } from '$lib/shared/format/percent';
	import { formatPct } from '../portfolio';
	import { formatUnits, unitNoun, type AssetPosition } from '../asset';
	import type { PortfolioRow } from '../portfolio';

	let {
		position,
		portfolioName
	}: {
		position: AssetPosition;
		/** Nombre del portafolio, para situar el peso de la posición dentro de él. */
		portfolioName: string;
	} = $props();

	const money = (amount: number, currency: string) =>
		privacy.money(formatCurrency(amount, currency));

	/** Los totales, en la única moneda en la que se pueden sumar. */
	const base = (amount: number) => money(amount, position.baseCurrency);

	const units = $derived(formatUnits(position.totalQty, position.assetType));

	/** El coste se pagó en una moneda y el mercado cotiza en otra. */
	const mixedCurrencies = $derived(position.costCurrency !== position.currency);

	/*
	 * Una ganancia que redondea a cero no es una ganancia: el efectivo de una
	 * cuenta la tiene siempre, y salía «+$0.00 (0,00%)» escrito en verde, que
	 * promete algo que la propia cifra desmiente.
	 */
	const flat = $derived(Math.abs(position.gainLoss) < 0.005);

	/*
	 * Con el precio de compra y el de mercado iguales, la frase de abajo no
	 * añade nada a la de arriba. Solo se puede afirmar cuando además están en
	 * la misma moneda: si no, son dos cifras que ni siquiera se comparan.
	 */
	const sameUnitPrice = $derived(
		!mixedCurrencies && Math.abs(position.averageCost - position.marketPrice) < 0.005
	);

	/*
	 * La barra habla el contrato de una fila del listado de portafolios: aquí
	 * la fila es la posición, y la escala es ella misma, así que llena el carril.
	 */
	const row = $derived<PortfolioRow>({
		id: 'self',
		name: '',
		description: '',
		typeLabel: '',
		riskName: '',
		isDefault: false,
		positions: 0,
		currency: position.baseCurrency,
		value: position.totalValue,
		cost: position.totalCost,
		gain: position.gainLoss,
		gainPct: position.gainLossPercent,
		converted: true,
		unconverted: 0
	});
</script>

<section class="headline" aria-labelledby="position-value">
	<h2 class="label" id="position-value">Tienes {units}</h2>
	<p class="amount">{base(position.totalValue)}</p>

	{#if position.totalCost <= 0}
		<p class="delta">Todavía no hay capital invertido que comparar.</p>
	{:else if flat}
		<p class="delta">Vale lo mismo que los {base(position.totalCost)} que invertiste.</p>
	{:else}
		<p class="delta" class:up={position.gainLoss > 0} class:down={position.gainLoss < 0}>
			{position.gainLoss > 0 ? '+' : '−'}{base(Math.abs(position.gainLoss))} sobre los {base(
				position.totalCost
			)} que invertiste ({formatPct(position.gainLossPercent)})
		</p>
	{/if}

	{#if position.totalValue + position.totalCost > 0}
		<div class="bar">
			<PortfolioCapitalBar
				{row}
				scale={Math.max(position.totalValue, position.totalCost)}
				displayCurrency={position.baseCurrency}
				legend
			/>
		</div>
	{/if}

	{#if !sameUnitPrice}
		<p class="unit-prices">
			Pagaste {money(position.averageCost, position.costCurrency)}
			por {unitNoun(position.assetType, 1)}; hoy cotiza a {money(
				position.marketPrice,
				position.currency
			)}.
			{#if mixedCurrencies}
				<span class="qualifier">
					La compra se liquidó en {position.costCurrency} y el mercado cotiza en {position.currency},
					así que los dos precios no se restan entre sí.
				</span>
			{/if}
		</p>
	{/if}

	{#if position.allocation > 0}
		<p class="weight">
			Es el {formatPercent(position.allocation)} de{portfolioName
				? ` ${portfolioName}`
				: ' este portafolio'}.
		</p>
	{/if}
</section>

<style>
	.headline {
		padding-bottom: 2rem;
		border-bottom: 1px solid var(--border);
	}

	/* Nombra la cifra de debajo y de paso dice la cantidad, que era otra
	   tarjeta. En caja normal: la ficha tenía seis antetítulos en versalitas. */
	.label {
		margin: 0 0 0.5rem;
		font-family: var(--font-body);
		font-size: 0.9rem;
		font-weight: 400;
		color: var(--text-muted);
	}

	.amount {
		margin: 0;
		font-family: var(--font-mono);
		font-size: clamp(2rem, 4.5vw, 2.75rem);
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.03em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.delta {
		max-width: 62ch;
		margin: 0.85rem 0 0;
		font-size: 0.9rem;
		line-height: 1.45;
		color: var(--text-muted);
	}

	.delta.up {
		color: var(--green);
	}

	.delta.down {
		color: var(--red);
	}

	.bar {
		max-width: 34rem;
		margin-top: 1.25rem;
	}

	.unit-prices,
	.weight {
		max-width: 62ch;
		margin: 1.25rem 0 0;
		font-size: 0.88rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.weight {
		margin-top: 0.35rem;
	}

	/* El aviso de monedas distintas va en ámbar porque cambia cómo se leen los
	   dos precios de la frase, no porque haga falta adornarla. */
	.qualifier {
		display: block;
		margin-top: 0.35rem;
		font-size: 0.82rem;
		color: var(--amber);
	}
</style>
