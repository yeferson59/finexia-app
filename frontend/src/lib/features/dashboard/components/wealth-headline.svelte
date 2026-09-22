<script lang="ts">
	/*
	 * La cifra: cuánto hay, cuánto de eso se ha ganado y hacia dónde va.
	 *
	 * Es lo único ruidoso de la página. Antes esta misma cifra vivía dentro de
	 * una tarjeta con borde, y el mismo total volvía a salir en la tarjeta de
	 * crecimiento y en el pie del listado de portafolios: tres cajas repitiendo
	 * dos números. Aquí sale una vez, en grande, y el resto de la página la
	 * descompone en vez de repetirla: la franja del reparto va pegada debajo.
	 *
	 * Tuvo al lado una curva en miniatura, que era la gráfica de crecimiento de
	 * más abajo en pequeño. Se fue por lo mismo que se fueron las tarjetas.
	 */
	import { resolve } from '$app/paths';
	import CurrencySelect from '$lib/ui/currency-select.svelte';
	import ExchangeRateNote from './exchange-rate-note.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { FALLBACK_CURRENCY, partitionByCurrency } from '$lib/shared/currency';
	import { plural } from '../breakdown';
	import type { ExchangeRate, PortfolioSummary } from '$lib/api/types';

	interface Props {
		summaries: PortfolioSummary[];
		currency?: string;
		/** Tasa con la que están convertidas estas cifras; `null` si no hay conversión. */
		displayRate?: ExchangeRate | null;
	}

	let { summaries = [], currency = FALLBACK_CURRENCY, displayRate = null }: Props = $props();

	// Un portafolio que el backend no pudo convertir viene en su propia moneda.
	// Sumarlo daría un patrimonio que no está en ninguna: se deja fuera y se
	// dice cuántos faltan, que es un total honesto en vez de uno inventado.
	const split = $derived(partitionByCurrency(summaries, currency));
	const counted = $derived(split.converted);
	const excluded = $derived(split.unconverted.length);

	const sum = (pick: (s: PortfolioSummary) => string) =>
		counted.reduce((acc, s) => acc + (parseFloat(pick(s) || '0') || 0), 0);

	const netWorth = $derived(sum((s) => s.totalMarketValue));
	const gain = $derived(sum((s) => s.totalGainLoss));
	const invested = $derived(sum((s) => s.totalCostBase));
	const positions = $derived(counted.reduce((acc, s) => acc + (s.totalPositions ?? 0), 0));
	const gainPct = $derived(invested > 0 ? (gain / invested) * 100 : 0);
	const up = $derived(gain >= 0);

	const money = (value: number) => privacy.money(formatCurrency(value, currency));

	/*
	 * La cifra en dos piezas: lo entero en grande y los centavos a media altura.
	 * Son la parte que menos informa y a este tamaño pesaban como otros tres
	 * dígitos. Si no hay centavos —el peso se escribe sin ellos— o el modo
	 * privado tapa la cifra, va entera.
	 */
	const figure = $derived.by(() => {
		const text = money(netWorth);
		const match = /^(.*\d)([.,]\d{2})(\D*)$/.exec(text);
		return match ? { whole: match[1], cents: match[2] + match[3] } : { whole: text, cents: '' };
	});
</script>

<section class="headline" aria-labelledby="net-worth">
	<div class="figure">
		<h1 class="label" id="net-worth">Patrimonio total</h1>

		<p class="amount">
			{figure.whole}{#if figure.cents}<span class="cents">{figure.cents}</span>{/if}
		</p>

		{#if counted.length > 0}
			<p class="delta">
				<span class:up class:down={!up}>
					{up ? '+' : '−'}{money(Math.abs(gain))} sobre lo invertido ({formatSignedPercent(
						gainPct,
						2
					)})</span
				>, en {plural(counted.length, 'portafolio', 'portafolios')} y {plural(
					positions,
					'posición',
					'posiciones'
				)}.
			</p>
		{:else if summaries.length > 0}
			<p class="delta">Ningún portafolio se puede convertir a {currency} todavía.</p>
		{:else}
			<p class="delta">
				Todavía no hay nada que sumar.
				<a class="start" href={resolve('/dashboard/portfolios/add')}>Crea tu primer portafolio</a>
				y registra lo que tienes en cada plataforma.
			</p>
		{/if}

		{#if excluded > 0 && counted.length > 0}
			<p class="fx">
				{excluded === 1 ? 'Falta un portafolio' : `Faltan ${excluded} portafolios`}: no hay tasa
				para pasarlo{excluded === 1 ? '' : 's'} a {currency}.
			</p>
		{/if}
	</div>

	<div class="rate">
		<CurrencySelect {currency} />
		<ExchangeRateNote rate={displayRate} />
	</div>
</section>

<style>
	.headline {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1.5rem 2.5rem;
		padding-bottom: 2.75rem;
	}

	.figure {
		min-width: 0;
	}

	/* Nombra la cifra que va justo debajo, así que no es una etiqueta de adorno. */
	.label {
		margin: 0 0 0.5rem;
		font-family: var(--font-body);
		font-size: 0.95rem;
		font-weight: 400;
		color: var(--text-muted);
	}

	/*
	 * Fraunces en un corte fino y a tamaño de cartel, con el eje óptico
	 * abierto: es un saldo, no una lectura de terminal, y la serie de la marca
	 * es la única cara de la aplicación que aguanta este tamaño con carácter.
	 * Cifras de caja alta y proporcionales: las tabulares se quedan en las
	 * columnas de abajo.
	 */
	.amount {
		margin: 0;
		font-family: var(--font-display);
		font-size: clamp(3rem, 8vw, 5.5rem);
		font-weight: 350;
		font-variant-numeric: lining-nums proportional-nums;
		font-optical-sizing: auto;
		line-height: 0.95;
		letter-spacing: -0.03em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.cents {
		margin-left: 0.04em;
		font-size: 0.45em;
		letter-spacing: 0;
		vertical-align: 0.95em;
		color: var(--text-muted);
	}

	.delta {
		max-width: 60ch;
		margin: 1.1rem 0 0;
		font-size: 0.95rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.up {
		color: var(--green);
	}

	.down {
		color: var(--red);
	}

	.start {
		color: var(--text);
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.fx {
		max-width: 44ch;
		margin: 1rem 0 0;
		padding-left: 0.75rem;
		border-left: 2px solid rgba(212, 145, 42, 0.45);
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.rate {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
	}

	@media (max-width: 860px) {
		.rate {
			align-items: flex-start;
		}
	}
</style>
