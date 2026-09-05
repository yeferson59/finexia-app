<script lang="ts">
	/*
	 * La cifra: cuánto hay, cuánto de eso se ha ganado y hacia dónde va.
	 *
	 * Es lo único ruidoso de la página. Antes esta misma cifra vivía dentro de
	 * una tarjeta con borde, y el mismo total volvía a salir en la tarjeta de
	 * crecimiento y en el pie del listado de portafolios: tres cajas repitiendo
	 * dos números. Aquí sale una vez, en grande, y el resto de la página la
	 * descompone en vez de repetirla.
	 */
	import { resolve } from '$app/paths';
	import CurrencyToggle from './currency-toggle.svelte';
	import ExchangeRateNote from './exchange-rate-note.svelte';
	import Sparkline from './sparkline.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { FALLBACK_CURRENCY, partitionByCurrency } from '$lib/shared/currency';
	import { plural } from '../breakdown';
	import type { ExchangeRate, GrowthDataPoint, PortfolioSummary } from '$lib/api/types';

	interface Props {
		summaries: PortfolioSummary[];
		currency?: string;
		/** Tasa con la que están convertidas estas cifras; `null` si no hay conversión. */
		displayRate?: ExchangeRate | null;
		/** Serie del patrimonio, para la curva de al lado. */
		series?: GrowthDataPoint[];
	}

	let {
		summaries = [],
		currency = FALLBACK_CURRENCY,
		displayRate = null,
		series = []
	}: Props = $props();

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

	const values = $derived(series.map((point) => parseFloat(point.totalValue || '0') || 0));

	const money = (value: number) => privacy.money(formatCurrency(value, currency));

	/** Primera y última fecha de la curva, para decir qué tramo se está viendo. */
	const range = $derived.by(() => {
		if (series.length < 2) return '';
		const start = new Date(series[0].date + 'T00:00:00');
		return start.toLocaleDateString('es-CO', { month: 'long', year: 'numeric' });
	});
</script>

<section class="headline" aria-labelledby="net-worth">
	<div class="figure">
		<h1 class="label" id="net-worth">Patrimonio total</h1>

		<p class="amount">{money(netWorth)}</p>

		{#if counted.length > 0}
			<p class="delta" class:up class:down={!up}>
				{up ? '+' : '−'}{money(Math.abs(gain))} sobre lo invertido ({formatSignedPercent(
					gainPct,
					2
				)})
			</p>
			<p class="meta">
				Repartido en {plural(counted.length, 'portafolio', 'portafolios')} y {plural(
					positions,
					'posición',
					'posiciones'
				)}.
			</p>
		{:else if summaries.length > 0}
			<p class="delta">Ningún portafolio se puede convertir a {currency} todavía.</p>
		{:else}
			<p class="delta">Todavía no hay nada que sumar.</p>
			<p class="meta">
				<a class="start" href={resolve('/dashboard/portfolios/add')}> Crea tu primer portafolio </a>
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

	<div class="side">
		<div class="rate">
			<CurrencyToggle {currency} />
			<ExchangeRateNote rate={displayRate} />
		</div>

		{#if values.length > 1}
			<div class="spark">
				<Sparkline {values} />
				<p class="spark-caption">Desde {range}</p>
			</div>
		{/if}
	</div>
</section>

<style>
	.headline {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(220px, 300px);
		align-items: start;
		gap: 2.5rem;
		padding-bottom: 2rem;
		border-bottom: 1px solid var(--border);
	}

	/* Nombra la cifra que va justo debajo, así que no es una etiqueta de adorno.
	   En caja normal y en la tipografía del resto de la interfaz: el panel tenía
	   cinco antetítulos en versalitas monoespaciadas, uno sobre cada tarjeta. */
	.label {
		margin: 0 0 0.65rem;
		font-family: var(--font-body);
		font-size: 0.9rem;
		font-weight: 400;
		color: var(--text-muted);
	}

	/*
	 * Cifras proporcionales, no tabulares: a este tamaño `tabular-nums` da a
	 * cada dígito el ancho de un cero y el número se ve suelto. Las tabulares se
	 * quedan donde hacen falta, que es en las columnas de abajo.
	 */
	.amount {
		margin: 0;
		font-family: var(--font-mono);
		font-size: clamp(2.5rem, 6.5vw, 4rem);
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.035em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.delta {
		margin: 0.85rem 0 0;
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	.delta.up {
		color: var(--green);
	}

	.delta.down {
		color: var(--red);
	}

	.meta {
		margin: 0.3rem 0 0;
		font-size: 0.85rem;
		color: var(--text-dim);
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

	.side {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.rate {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
	}

	.spark-caption {
		margin: 0.4rem 0 0;
		font-size: 0.72rem;
		color: var(--text-dim);
		text-align: right;
	}

	@media (max-width: 860px) {
		.headline {
			grid-template-columns: minmax(0, 1fr);
			gap: 1.75rem;
		}

		.rate {
			align-items: flex-start;
		}

		.spark-caption {
			text-align: left;
		}
	}
</style>
