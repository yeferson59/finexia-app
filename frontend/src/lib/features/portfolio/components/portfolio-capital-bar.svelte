<script lang="ts">
	/*
	 * Cuánto puso el dueño y qué ha hecho el mercado con ello, en una barra.
	 *
	 * Es la gráfica de crecimiento del portafolio contraída a un solo instante:
	 * el mismo par que dibujan sus dos series —el capital invertido, frío, y el
	 * valor de mercado, cálido—. El corte cae siempre en el capital: lo que
	 * queda dentro de la barra es la ganancia y lo que asoma por fuera de su
	 * extremo es lo que falta para recuperarlo.
	 *
	 * El largo de la parte sólida es siempre el valor de mercado, nunca el
	 * coste: es lo que mantiene comparables las filas entre sí.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import type { PortfolioRow } from '../portfolio';

	let {
		row,
		scale,
		displayCurrency,
		legend = false
	}: {
		row: PortfolioRow;
		/** Mayor importe de la lista entera: el ancho del carril. */
		scale: number;
		displayCurrency: string;
		/**
		 * Pie con la clave de los colores. En el listado no hace falta —la
		 * cabecera de la columna la lleva, una vez para todas las filas— pero en
		 * el detalle la barra va sola y tiene que explicarse.
		 */
		legend?: boolean;
	} = $props();

	const drawable = $derived(row.converted && row.value + row.cost > 0);

	/** Ancho en porcentaje del carril, acotado por si llega un importe absurdo. */
	function width(amount: number): number {
		if (scale <= 0) return 0;

		return Math.max(0, Math.min(100, (amount / scale) * 100));
	}
</script>

{#if drawable}
	<span class="track" aria-hidden="true">
		<span class="held" style="width: {width(Math.min(row.value, row.cost))}%"></span>
		{#if row.gain > 0}
			<span class="gain" style="width: {width(row.gain)}%"></span>
		{:else if row.gain < 0}
			<span class="short" style="width: {width(-row.gain)}%"></span>
		{/if}
	</span>
	<span class="sr-only">
		Capital invertido: {privacy.money(formatCurrency(row.cost, row.currency))}.
	</span>
	{#if legend}
		<span class="key" aria-hidden="true">
			<span class="swatch swatch-cost"></span>capital invertido
			{#if row.gain < 0}
				<span class="swatch swatch-short"></span>lo que falta para recuperarlo
			{:else if row.gain > 0}
				<span class="swatch swatch-gain"></span>ganancia
			{/if}
		</span>
	{/if}
{:else if row.converted}
	<span class="idle">Sin posiciones todavía</span>
{:else}
	<span class="idle">Sin tasa a {displayCurrency}</span>
{/if}

<style>
	.track {
		display: flex;
		height: 9px;
		margin-top: 0.3rem;
		border-radius: 2px;
		background: var(--panel-2);
		overflow: hidden;
	}

	.held {
		background: var(--cost);
	}

	/*
	 * Un mínimo de tres píxeles para la ganancia y para lo que falta. Sin él,
	 * una pérdida del 2,6 % era una franja de cuatro píxeles que no se veía, y
	 * la barra de un portafolio en pérdida quedaba idéntica a la de uno que no
	 * hubiera ganado nada. Solo se dibujan cuando hay algo que dibujar, así que
	 * el mínimo no inventa una ganancia donde no la hay.
	 */
	.gain {
		min-width: 3px;
		background: var(--green);
	}

	/*
	 * Lo que falta para recuperar el capital. Va translúcido porque no es barra:
	 * es el hueco que dejó al quedarse corta, y sólido parecería otro tramo de
	 * valor. El filete derecho marca dónde está el capital, que es el extremo al
	 * que no llegó.
	 */
	.short {
		min-width: 3px;
		background: rgba(224, 90, 90, 0.32);
		box-shadow: inset -2px 0 0 var(--red);
	}

	.key {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		margin-top: 0.5rem;
		font-size: 0.75rem;
		color: var(--text-dim);
	}

	.swatch {
		width: 9px;
		height: 9px;
		border-radius: 2px;
	}

	.swatch + .swatch,
	.key .swatch:not(:first-child) {
		margin-left: 0.5rem;
	}

	.swatch-cost {
		background: var(--cost);
	}

	.swatch-gain {
		background: var(--green);
	}

	.swatch-short {
		background: rgba(224, 90, 90, 0.32);
		box-shadow: inset -2px 0 0 var(--red);
	}

	.idle {
		display: block;
		margin-top: 0.15rem;
		font-size: 0.78rem;
		color: var(--text-dim);
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}
</style>
