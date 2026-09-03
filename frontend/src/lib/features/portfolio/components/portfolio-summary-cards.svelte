<script lang="ts">
	import Card from '$lib/ui/card.svelte';
	import { formatPct } from '../portfolio';

	let {
		totalValue,
		totalCost,
		baseCurrency,
		totalGainLoss,
		totalGainLossPct,
		realReturn,
		riskName,
		holdingsCount,
		formatCurrency
	}: {
		totalValue: number;
		totalCost: number;
		baseCurrency: string;
		totalGainLoss: number;
		totalGainLossPct: number;
		/**
		 * Rentabilidad ponderada por tiempo del historial, en porcentaje; `null`
		 * mientras la serie de crecimiento no dé ni un tramo que medir.
		 */
		realReturn: number | null;
		riskName: string | undefined;
		holdingsCount: number;
		formatCurrency: (value: number) => string;
	} = $props();
</script>

<section class="cards-grid">
	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Valor de mercado</p>
		<h2 class="hero-value">{formatCurrency(totalValue)}</h2>
		<p class="hero-delta">Costo: {formatCurrency(totalCost)} · {baseCurrency}</p>
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Ganancia / Pérdida</p>
		<h2 class="hero-value {totalGainLoss >= 0 ? 'positive' : 'negative'}">
			{formatCurrency(totalGainLoss)}
		</h2>
		<p class="hero-delta {totalGainLoss >= 0 ? 'positive' : 'negative'}">
			{formatPct(totalGainLossPct)} sobre costo
		</p>
	</Card>

	<!-- La ganancia de al lado divide por lo invertido hoy, así que un aporte
	     hecho después de una subida la hunde sin que la cartera haya perdido
	     nada. Esta encadena los tramos del historial y solo se mueve con el
	     mercado: es lo que rindió el dinero, no cuánto dinero entró. -->
	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Rentabilidad real</p>
		{#if realReturn === null}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Necesita al menos dos cierres diarios del portafolio.</p>
		{:else}
			<h2 class="hero-value {realReturn >= 0 ? 'positive' : 'negative'}">
				{formatPct(realReturn)}
			</h2>
			<p class="hero-delta">Limpia de aportes y retiros</p>
		{/if}
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Riesgo · Activos</p>
		<h2 class="hero-value">{riskName ?? '—'}</h2>
		<p class="hero-delta">
			{holdingsCount}
			{holdingsCount === 1 ? 'activo' : 'activos'}
		</p>
	</Card>
</section>

<style>
	.cards-grid {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.eyebrow {
		margin: 0 0 0.55rem;
		font-size: 0.72rem;
		letter-spacing: 0.7px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.46);
	}

	.hero-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		margin: 0;
		font-size: 1.6rem;
		color: var(--text);
	}

	.hero-value.positive {
		color: var(--green);
	}

	.hero-value.negative {
		color: var(--red);
	}

	.hero-delta {
		margin: 0.4rem 0 0;
		font-size: 0.82rem;
		color: rgba(236, 234, 229, 0.55);
	}

	.hero-delta.positive {
		color: var(--green);
	}

	.hero-delta.negative {
		color: var(--red);
	}

	/* Con cuatro tarjetas, una sola columna a los 1024px dejaba media pantalla
	   de scroll antes de la gráfica; caen de dos en dos hasta el móvil. */
	@media (max-width: 1024px) {
		.cards-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 640px) {
		.cards-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
