<script lang="ts">
	/*
	 * Lo que vale el portafolio y de dónde sale esa cifra.
	 *
	 * Ocupa el sitio de cuatro tarjetas —valor de mercado, ganancia,
	 * rentabilidad real y riesgo— de las que tres repetían números que ya
	 * estaban en otra parte de la misma página: el valor volvía a salir en el
	 * centro del donut y en el pie de la gráfica, el capital en la tarjeta de
	 * «capital invertido» de más abajo, y la rentabilidad real en el pie de la
	 * gráfica, donde además está bien puesta porque depende del tramo elegido.
	 *
	 * La barra es la misma del listado, así que abrir un portafolio no cambia de
	 * idioma: el corte cae en el capital y lo que sigue es lo que ha ganado.
	 */
	import PortfolioCapitalBar from './portfolio-capital-bar.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPct } from '../portfolio';
	import type { PortfolioRow } from '../portfolio';

	let {
		value,
		cost,
		baseCurrency
	}: {
		value: number;
		cost: number;
		baseCurrency: string;
	} = $props();

	const gain = $derived(value - cost);
	const gainPct = $derived(cost > 0 ? (gain / cost) * 100 : 0);

	const money = (amount: number) => privacy.money(formatCurrency(amount, baseCurrency));

	/*
	 * La barra habla el contrato de una fila del listado. Aquí solo hay un
	 * portafolio, así que la escala es él mismo: su barra llena el carril y lo
	 * que se lee dentro es el reparto entre capital y ganancia.
	 */
	const row = $derived<PortfolioRow>({
		id: 'self',
		name: '',
		description: '',
		typeLabel: '',
		riskName: '',
		isDefault: false,
		positions: 0,
		currency: baseCurrency,
		value,
		cost,
		gain,
		gainPct,
		converted: true,
		unconverted: 0
	});
</script>

<section class="headline" aria-labelledby="market-value">
	<h2 class="label" id="market-value">Valor de mercado</h2>
	<p class="amount">{money(value)}</p>

	{#if cost > 0}
		<p class="delta" class:up={gain >= 0} class:down={gain < 0}>
			{gain >= 0 ? '+' : '−'}{money(Math.abs(gain))} sobre los {money(cost)} que invertiste ({formatPct(
				gainPct
			)})
		</p>
	{:else}
		<p class="delta">Todavía no hay capital invertido que comparar.</p>
	{/if}

	{#if value + cost > 0}
		<div class="bar">
			<PortfolioCapitalBar
				{row}
				scale={Math.max(value, cost)}
				displayCurrency={baseCurrency}
				legend
			/>
		</div>
	{/if}
</section>

<style>
	.headline {
		padding-bottom: 2rem;
		border-bottom: 1px solid var(--border);
	}

	/* Nombra la cifra de debajo, así que no es una etiqueta de adorno. En caja
	   normal: el detalle tenía diez antetítulos en versalitas. */
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
</style>
