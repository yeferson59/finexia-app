<script lang="ts">
	/*
	 * Calendario de rentabilidad mensual, un panel por año.
	 *
	 * El color solo era color: un mes verde y uno rojo se distinguían mirando,
	 * pero no leyendo. Ahora cada celda dice su mes y su signo en el nombre
	 * accesible, hay leyenda de la escala y el año lleva su acumulado, que es lo
	 * primero que se busca al abrir la página.
	 *
	 * Las cifras son rendimiento, no variación del saldo: `reports.ts` descuenta
	 * los aportes de cada tramo. El pie lo dice, porque una cuenta que crece a
	 * base de depósitos venía marcando meses del +150 % y nadie sabía por qué.
	 */
	import ReportPanel from './report-panel.svelte';
	import { MONTHS, performanceClass, type PerformanceCalendar } from '../reports';

	interface Props {
		calendars: PerformanceCalendar[];
	}

	let { calendars }: Props = $props();

	function fmtPct(value: number): string {
		const formatted = new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 1,
			maximumFractionDigits: 1
		}).format(value);
		return `${value > 0 ? '+' : ''}${formatted}%`;
	}

	/** Suma compuesta de los meses con dato: lo que rindió el año. */
	function yearTotal(values: (number | null)[]): number | null {
		const months = values.filter((v): v is number => v !== null);
		if (months.length === 0) return null;
		return (months.reduce((acc, v) => acc * (1 + v / 100), 1) - 1) * 100;
	}

	function cellLabel(month: string, value: number | null, partial: boolean): string {
		if (value === null) return `${month}: sin dato`;
		const sign = value > 0 ? 'positivo' : value < 0 ? 'negativo' : 'plano';
		const scope = partial ? ', mes parcial' : '';
		return `${month}: ${fmtPct(value)}, ${sign}${scope}`;
	}
</script>

<section class="analytics-grid">
	{#if calendars.length > 0}
		{#each calendars as calendar (calendar.year)}
			{@const total = yearTotal(calendar.values)}
			<ReportPanel class="calendar-card" title="Rentabilidad mensual (%)" badge={calendar.year}>
				{#if total !== null}
					<p class="year-total">
						<span>Acumulado del año</span>
						<b class:up={total >= 0} class:down={total < 0}>{fmtPct(total)}</b>
					</p>
				{/if}

				<div class="calendar-grid">
					{#each calendar.values as value, index (`${calendar.year}-${MONTHS[index]}`)}
						{@const partial = value !== null && calendar.partialMonths.includes(index)}
						<div
							class={`month-cell ${value === null ? 'null-cell' : performanceClass(value)}`}
							role="img"
							aria-label={cellLabel(MONTHS[index], value, partial)}
						>
							<p class="month" aria-hidden="true">
								{MONTHS[index]}{#if partial}<span class="partial">*</span>{/if}
							</p>
							<p class="percent" aria-hidden="true">{value === null ? '–' : fmtPct(value)}</p>
						</div>
					{/each}
				</div>

				<ul class="scale" aria-hidden="true">
					<li><i class="swatch strong-negative"></i>≤ −1%</li>
					<li><i class="swatch negative"></i>&lt; 0%</li>
					<li><i class="swatch flat-positive"></i>0–1%</li>
					<li><i class="swatch positive"></i>1–2%</li>
					<li><i class="swatch strong-positive"></i>≥ 2%</li>
				</ul>

				<p class="footnote">
					Rendimiento de lo invertido: los aportes y retiros del mes no cuentan como rentabilidad.{#if calendar.partialMonths.length > 0}
						<br />* Mes parcial: el historial no lo cubre entero, porque empieza dentro de él o
						porque el mes sigue en curso. No entra en el mejor ni en el peor mes de las
						estadísticas.{/if}
				</p>
			</ReportPanel>
		{/each}
	{:else}
		<ReportPanel tag="div" class="empty-panel">
			<p class="empty-text">Sin datos históricos</p>
		</ReportPanel>
	{/if}
</section>

<style>
	.analytics-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.analytics-grid :global(.calendar-card) {
		padding: 1rem;
	}

	.analytics-grid :global(.empty-panel) {
		padding: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		grid-column: 1 / -1;
	}

	.year-total {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.75rem;
		margin: 0 0 0.8rem;
		padding-bottom: 0.7rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.7rem;
		color: var(--text-dim);
	}

	.year-total b {
		font-family: var(--font-mono);
		font-size: 0.95rem;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}

	.year-total b.up {
		color: var(--green);
	}

	.year-total b.down {
		color: var(--red);
	}

	.calendar-grid {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 0.45rem;
	}

	.month-cell {
		padding: 0.5rem;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.022);
		border: 1px solid transparent;
	}

	.month {
		margin: 0;
		font-size: 0.65rem;
		color: rgba(236, 234, 229, 0.55);
	}

	.percent {
		margin: 0.18rem 0 0;
		font-size: 0.76rem;
		font-weight: 700;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.month-cell.strong-positive,
	.swatch.strong-positive {
		background: rgba(34, 201, 126, 0.26);
		border-color: rgba(34, 201, 126, 0.45);
		color: var(--green);
	}

	.month-cell.positive,
	.swatch.positive {
		background: rgba(34, 201, 126, 0.18);
		border-color: rgba(34, 201, 126, 0.3);
		color: var(--green);
	}

	.month-cell.flat-positive,
	.swatch.flat-positive {
		background: rgba(212, 145, 42, 0.2);
		border-color: rgba(212, 145, 42, 0.35);
		color: var(--amber-light);
	}

	.month-cell.negative,
	.swatch.negative {
		background: rgba(224, 90, 90, 0.16);
		border-color: rgba(224, 90, 90, 0.3);
		color: var(--red);
	}

	.month-cell.strong-negative,
	.swatch.strong-negative {
		background: rgba(224, 90, 90, 0.26);
		border-color: rgba(224, 90, 90, 0.46);
		color: var(--red);
	}

	.month-cell.null-cell {
		background: rgba(255, 255, 255, 0.022);
		border-color: transparent;
		color: var(--text-dim);
	}

	.scale {
		display: flex;
		flex-wrap: wrap;
		gap: 0.3rem 0.7rem;
		margin: 0.85rem 0 0;
		padding: 0;
		list-style: none;
		font-family: var(--font-mono);
		font-size: 0.58rem;
		color: var(--text-dim);
	}

	.scale li {
		display: flex;
		align-items: center;
		gap: 0.3rem;
	}

	.swatch {
		width: 9px;
		height: 9px;
		border-radius: 2px;
		border: 1px solid transparent;
	}

	.partial {
		margin-left: 0.1rem;
		color: var(--text-dim);
	}

	.footnote {
		margin: 0.6rem 0 0;
		font-size: 0.6rem;
		line-height: 1.5;
		color: var(--text-dim);
	}

	@media (max-width: 1024px) {
		.analytics-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
