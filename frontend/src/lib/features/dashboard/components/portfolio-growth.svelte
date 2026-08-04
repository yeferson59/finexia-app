<script lang="ts">
	import CardHeader from '$lib/ui/card-header.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import Stat from '$lib/ui/stat.svelte';
	import GrowthChart from './growth-chart.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import {
		PERIODS,
		filterByPeriod,
		growthScale,
		type GrowthPoint,
		type Period
	} from '../dashboard';

	import type { GrowthDataPoint, GrowthSummary } from '$lib/api/types';

	const {
		data = [],
		summary = { firstDate: '', initialValue: '0', currentValue: '0', totalGrowthPct: '0' }
	}: { data: GrowthDataPoint[]; summary: GrowthSummary } = $props();

	let selectedPeriod = $state<Period>('Todo');
	/** Punto bajo el cursor (ratón o teclado); `null` cuando no hay ninguno. */
	let activeIndex = $state<number | null>(null);

	const periods = PERIODS;

	const filteredData = $derived(filterByPeriod(data, selectedPeriod));

	const points = $derived<GrowthPoint[]>(
		filteredData.map((d) => ({
			mv: parseFloat(d.totalValue || '0'),
			cb: parseFloat(d.totalCostBase || '0'),
			date: d.date
		}))
	);

	const scale = $derived(growthScale(points.flatMap((p) => [p.mv, p.cb])));

	const activePoint = $derived(activeIndex === null ? null : (points[activeIndex] ?? null));
	const activeGain = $derived(activePoint ? activePoint.mv - activePoint.cb : 0);

	// Métricas del resumen (la foto completa, no la del punto señalado).
	const currentVal = $derived(parseFloat(summary.currentValue || '0'));
	const initialVal = $derived(parseFloat(summary.initialValue || '0'));
	const totalGrowthPct = $derived(parseFloat(summary.totalGrowthPct || '0'));
	const absoluteGain = $derived(currentVal - initialVal);
	const isPositive = $derived(absoluteGain >= 0);

	/** Al cambiar de periodo el índice anterior ya no señala al mismo día. */
	function selectPeriod(period: Period) {
		selectedPeriod = period;
		activeIndex = null;
	}

	function fmt(v: number): string {
		return new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		}).format(v);
	}

	function fmtMoney(v: number): string {
		return privacy.money('$' + fmt(v));
	}

	/*
	 * Etiquetas del eje. Entre 1.000 y 10.000 hace falta un decimal: redondeando
	 * a miles, una serie de 1.500 a 1.900 pintaba "$2k" en las cinco marcas.
	 */
	function fmtAbbrev(v: number): string {
		const abs = Math.abs(v);
		if (abs >= 1_000_000) return privacy.money(`$${(v / 1_000_000).toFixed(1)}M`);
		if (abs >= 10_000) return privacy.money(`$${(v / 1_000).toFixed(0)}k`);
		if (abs >= 1_000) return privacy.money(`$${(v / 1_000).toFixed(1)}k`);
		return privacy.money(`$${v.toFixed(0)}`);
	}

	/*
	 * Con más de un año a la vista, día y mes no bastan: el eje repetía "1 de
	 * jun" para dos junios distintos. El año solo aparece cuando hace falta,
	 * para no recargar los rangos cortos.
	 */
	const spansYears = $derived(
		points.length > 1 && points[0].date.slice(0, 4) !== points[points.length - 1].date.slice(0, 4)
	);

	/*
	 * En rangos largos el eje pasa a "mes año" y suelta el día: con
	 * "01 de jun de 25" las seis etiquetas se pisaban unas a otras, y a esa
	 * escala el día no aporta nada —el detalle exacto lo da el cursor—.
	 */
	function fmtDate(iso: string): string {
		const d = new Date(iso + 'T00:00:00');
		return spansYears
			? d.toLocaleDateString('es-CO', { month: 'short', year: '2-digit' })
			: d.toLocaleDateString('es-CO', { day: '2-digit', month: 'short' });
	}

	function fmtLongDate(iso: string): string {
		const d = new Date(iso + 'T00:00:00');
		return d.toLocaleDateString('es-CO', { day: '2-digit', month: 'long', year: 'numeric' });
	}
</script>

<div class="growth-card">
	<div class="card-top">
		<CardHeader eyebrow="Portafolio" title="Crecimiento del portafolio" divider={false} />
		<div class="period-tabs" role="tablist" aria-label="Período">
			{#each periods as p (p)}
				<button
					role="tab"
					aria-selected={selectedPeriod === p}
					class="period-btn"
					class:active={selectedPeriod === p}
					onclick={() => selectPeriod(p)}>{p}</button
				>
			{/each}
		</div>
	</div>

	<div class="divider"></div>

	<div class="metrics-row">
		<Stat
			label="Total ganancia"
			tone={isPositive ? 'positive' : 'negative'}
			value="{isPositive ? '+' : '−'}{fmtMoney(Math.abs(absoluteGain))}"
		/>
		<Stat
			label="Desde creación"
			tone={isPositive ? 'positive' : 'negative'}
			value="{isPositive ? '+' : ''}{fmt(totalGrowthPct)}%"
		/>
		<Stat label="Valor actual" tone="highlight" value={fmtMoney(currentVal)} />
	</div>

	{#if points.length < 2}
		<EmptyState
			bordered
			title="Aún no hay suficiente historial"
			description="El gráfico se dibuja a medida que el sistema registra las capturas diarias de tu portafolio."
		/>
	{:else}
		<!-- El detalle del punto señalado. Sin `aria-live`: la gráfica es un
		     `slider` y su `aria-valuetext` ya lo anuncia; duplicarlo aquí haría
		     que el lector de pantalla lo dijera dos veces. -->
		<p class="readout">
			{#if activePoint}
				<span class="readout-date">{fmtLongDate(activePoint.date)}</span>
				<span class="readout-item">
					<i class="swatch value"></i>Valor <b>{fmtMoney(activePoint.mv)}</b>
				</span>
				<span class="readout-item">
					<i class="swatch cost"></i>Invertido <b>{fmtMoney(activePoint.cb)}</b>
				</span>
				<span class="readout-item" class:up={activeGain >= 0} class:down={activeGain < 0}>
					Ganancia <b>{activeGain >= 0 ? '+' : '−'}{fmtMoney(Math.abs(activeGain))}</b>
				</span>
			{:else}
				<span class="readout-hint">
					Pasa el cursor por la gráfica —o enfócala y usa las flechas— para ver el detalle de cada
					día.
				</span>
			{/if}
		</p>

		<GrowthChart
			{points}
			{scale}
			active={activeIndex}
			formatAbbrev={fmtAbbrev}
			formatDate={fmtDate}
			formatFullDate={fmtLongDate}
			formatMoney={fmtMoney}
			onactivate={(index) => (activeIndex = index)}
		/>

		<div class="legend">
			<div class="legend-item">
				<span class="legend-line amber"></span>
				<span>Valor de mercado</span>
			</div>
			<div class="legend-item">
				<span class="legend-line gray"></span>
				<span>Capital invertido</span>
			</div>
		</div>
	{/if}
</div>

<style>
	.growth-card {
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		padding: 2rem;
		backdrop-filter: blur(10px);
	}

	.card-top {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.25rem;
	}

	.divider {
		height: 1px;
		background: var(--border);
		margin-bottom: 1.5rem;
	}

	.period-tabs {
		display: flex;
		gap: 0.2rem;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 0.2rem;
		flex-shrink: 0;
	}

	.period-btn {
		padding: 0.3rem 0.65rem;
		border: none;
		background: transparent;
		color: var(--text-dim);
		border-radius: 6px;
		font-size: 0.72rem;
		font-weight: 600;
		font-family: var(--font-mono);
		cursor: pointer;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.period-btn.active {
		background: rgba(212, 145, 42, 0.18);
		color: var(--amber-light);
	}

	.period-btn:hover:not(.active) {
		background: rgba(255, 255, 255, 0.05);
		color: var(--text);
	}

	.metrics-row {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1.5rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	/* Alto fijo: el detalle aparece y desaparece sin mover la gráfica. */
	.readout {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.5rem 1.25rem;
		min-height: 2.4rem;
		margin: 0 0 0.5rem;
		font-size: 0.78rem;
		color: var(--text-muted);
	}

	.readout-date {
		font-family: var(--font-mono);
		font-size: 0.7rem;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.readout-item {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
	}

	.readout-item b {
		font-family: var(--font-mono);
		font-weight: 600;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.readout-item.up b {
		color: var(--green);
	}

	.readout-item.down b {
		color: var(--red);
	}

	.swatch {
		width: 10px;
		height: 2px;
		border-radius: 1px;
	}

	.swatch.value {
		background: var(--amber);
	}

	.swatch.cost {
		background: rgba(236, 234, 229, 0.4);
	}

	.readout-hint {
		color: var(--text-dim);
	}

	.legend {
		display: flex;
		gap: 1.5rem;
		margin-top: 0.75rem;
		font-size: 0.72rem;
		font-family: var(--font-mono);
		color: var(--text-dim);
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.legend-line {
		display: inline-block;
		width: 22px;
		height: 0;
	}

	.legend-line.amber {
		border-top: 2.5px solid var(--amber);
	}

	.legend-line.gray {
		border-top: 1.5px dashed rgba(236, 234, 229, 0.28);
	}

	@media (max-width: 768px) {
		.growth-card {
			padding: 1.5rem;
		}
	}

	@media (max-width: 600px) {
		.metrics-row {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.card-top {
			flex-direction: column;
			align-items: flex-start;
		}

		.readout {
			min-height: 3.4rem;
		}
	}
</style>
