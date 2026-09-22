<script lang="ts">
	/**
	 * Las cifras de la proyección, debajo de la gráfica que la dibuja.
	 *
	 * La banda dice la forma; esto dice los números, que es lo que alguien
	 * apunta. Cada marca es un rango y no una cifra: entre lo que daría el peor
	 * mes repetido y lo que daría el mejor. Un punto medio inventaría una
	 * precisión que no hay.
	 *
	 * Tres marcas —trimestre, semestre y año— en vez de los doce meses: doce
	 * filas serían una tabla de amortización y dicen lo mismo.
	 *
	 * El encabezado lleva siempre de cuántos meses salen los dos extremos. No es
	 * un adorno: un rango sacado de dos meses y otro de dos años se escriben
	 * igual y no valen lo mismo.
	 */
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import type { ValueForecast } from '../projection';
	import { FORECAST_MIN_MONTHS, forecastMilestones } from '../projection';

	interface Props {
		forecast: ValueForecast | null;
		/** Cuántos meses completos hay hoy, para decir qué falta cuando no hay proyección. */
		completeMonths: number;
		/** Cómo escribir un importe; lo impone el panel, que conoce la moneda. */
		formatMoney: (value: number) => string;
	}

	let { forecast, completeMonths, formatMoney }: Props = $props();

	const milestones = $derived(forecastMilestones(forecast));

	const label = (monthsAhead: number) =>
		monthsAhead === 12 ? 'En un año' : `En ${monthsAhead} meses`;
</script>

<section class="forecast" aria-labelledby="growth-forecast-title">
	<div class="head">
		<h3 id="growth-forecast-title">Proyección</h3>
		{#if forecast}
			<p class="basis">
				Entre repetir doce veces tu peor mes
				<strong>({formatSignedPercent(forecast.worstMonthPct, 2)})</strong>
				y repetir tu mejor mes <strong>({formatSignedPercent(forecast.bestMonthPct, 2)})</strong>,
				de los {forecast.basisMonths} que llevas cerrados.
			</p>
		{/if}
	</div>

	{#if forecast}
		<dl class="marks">
			{#each milestones as entry (entry.monthsAhead)}
				<div class="mark">
					<dt>{label(entry.monthsAhead)}</dt>
					<dd>
						<span class="value">{formatMoney(entry.low)} – {formatMoney(entry.high)}</span>
						<span class="pct">
							{formatSignedPercent(entry.lowPct, 1)} a {formatSignedPercent(entry.highPct, 1)}
						</span>
					</dd>
				</div>
			{/each}
		</dl>

		<p class="caveat">
			No es una previsión: son los dos meses que ya tuviste, repetidos. Ni el mercado ni tus aportes
			futuros entran en la cuenta, y lo que pase de verdad casi nunca es un mes repetido doce veces.
			Los extremos equivalen a {formatSignedPercent(forecast.worstAnnualPct, 1)} y
			{formatSignedPercent(forecast.bestAnnualPct, 1)} anuales.
		</p>
	{:else if completeMonths < FORECAST_MIN_MONTHS}
		<p class="caveat">
			Todavía no hay con qué proyectar: hacen falta {FORECAST_MIN_MONTHS} meses cerrados para tener un
			mejor y un peor mes distintos, y por ahora {completeMonths === 1
				? 'ha cerrado uno'
				: 'no ha cerrado ninguno'}. El mes en que abriste la cuenta y el mes en curso no cuentan,
			porque van a medias, así que con tres meses de historial ya se ve.
		</p>
	{:else}
		<p class="caveat">
			No hay banda que dibujar: algún mes se llevó todo lo invertido, y repetirlo doce veces no
			lleva a ninguna cifra.
		</p>
	{/if}
</section>

<style>
	.forecast {
		margin-top: 1.5rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--border);
	}

	.head {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 0.3rem 0.75rem;
		margin-bottom: 1rem;
	}

	h3 {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1rem;
		font-weight: 400;
		color: var(--text);
	}

	.basis {
		margin: 0;
		max-width: 60ch;
		font-size: 0.82rem;
		color: var(--text-muted);
	}

	.basis strong {
		font-weight: 500;
		color: var(--text);
	}

	.marks {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.75rem;
		margin: 0;
	}

	.mark {
		padding: 0.75rem 0.85rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.02);
	}

	dt {
		font-size: 0.78rem;
		color: var(--text-muted);
	}

	dd {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 0.15rem 0.5rem;
		margin: 0.3rem 0 0;
	}

	.value {
		font-family: var(--font-mono);
		font-size: 0.95rem;
		color: var(--text);
	}

	.pct {
		font-size: 0.8rem;
		color: var(--amber);
	}

	.caveat {
		max-width: 68ch;
		margin: 0.9rem 0 0;
		font-size: 0.78rem;
		line-height: 1.5;
		color: var(--text-dim);
	}

	/* En pantalla estrecha las tres marcas se apilan: a un tercio del ancho de un
	   móvil, un importe de seis cifras no cabe y se parte por la mitad. */
	@media (max-width: 560px) {
		.marks {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
