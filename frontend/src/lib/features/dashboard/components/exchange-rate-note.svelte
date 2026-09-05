<script lang="ts">
	/*
	 * La tasa con la que están convertidas las cifras de la tarjeta.
	 *
	 * El selector de moneda cambia el número grande sin decir con qué se
	 * convirtió, y una cifra en pesos que nadie puede cuadrar con su banco no
	 * sirve de mucho. Esta línea enseña la tasa, su fecha y de dónde salió.
	 *
	 * No se muestra si no hay tasa: el backend la trae de un feed público y una
	 * caída suya es mejor callarla que rellenarla con un valor inventado.
	 */
	import { formatCalendarDate } from '$lib/shared/format/date';
	import type { ExchangeRate } from '$lib/api/types';

	const { rate }: { rate: ExchangeRate | null } = $props();

	/*
	 * Etiqueta de la fuente. `dolarapi` es la TRM: el backend solo publica ese
	 * dato desde ese feed (ver internal/platform/marketdata/dolarapi), y la TRM
	 * es el nombre que reconoce quien la va a comparar con su extracto.
	 */
	const SOURCE_LABELS: Record<string, string> = {
		dolarapi: 'TRM · dolarapi.com',
		manual: 'Tasa fijada manualmente'
	};

	const amount = $derived(rate ? Number.parseFloat(rate.rate) : 0);

	const formatted = $derived(
		new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		}).format(amount)
	);

	const sourceLabel = $derived(rate ? (SOURCE_LABELS[rate.source] ?? rate.source) : '');

	const asOf = $derived(
		rate
			? formatCalendarDate(rate.rateDate, { day: 'numeric', month: 'short', year: 'numeric' })
			: ''
	);
</script>

{#if rate && Number.isFinite(amount) && amount > 0}
	<p class="rate">
		<span class="pair">1&nbsp;{rate.fromCurrency} = {formatted}&nbsp;{rate.toCurrency}</span>
		<span class="meta">{sourceLabel} · {asOf}</span>
	</p>
{/if}

<style>
	.rate {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.1rem;
		margin: 0.4rem 0 0;
		text-align: right;
	}

	.pair {
		font-family: var(--font-mono);
		font-size: 0.72rem;
		font-weight: 600;
		color: var(--text-dim);
		font-variant-numeric: tabular-nums;
	}

	.meta {
		font-size: 0.66rem;
		color: var(--text-dim);
	}

	@media (max-width: 480px) {
		.rate {
			align-items: flex-start;
			text-align: left;
		}
	}
</style>
