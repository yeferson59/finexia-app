<script lang="ts">
	/*
	 * Los dos conmutadores de la cabecera de la gráfica: en qué unidad se dibuja
	 * —dinero o rentabilidad— y qué tramo del historial se mira.
	 *
	 * Salieron de `portfolio-growth` cuando la tarjeta pasó del presupuesto de
	 * 500 líneas; aquí, juntos, es además más difícil que uno de los dos acabe
	 * pareciendo otra cosa que el otro.
	 *
	 * Interno de la feature.
	 */
	import { GROWTH_VIEWS, PERIODS, type GrowthView, type Period } from '../dashboard';

	interface Props {
		view: GrowthView;
		period: Period;
		onview: (view: GrowthView) => void;
		onperiod: (period: Period) => void;
	}

	let { view, period, onview, onperiod }: Props = $props();
</script>

<div class="chart-controls">
	<!-- Botones de alternancia, no pestañas: no cambian de panel, cambian la
	     unidad de lo que ya se está mirando. -->
	<div class="tab-group" role="group" aria-label="Unidad de la gráfica">
		{#each GROWTH_VIEWS as option (option.id)}
			<button
				type="button"
				class="tab-btn"
				class:active={view === option.id}
				aria-pressed={view === option.id}
				title={option.hint}
				onclick={() => onview(option.id)}>{option.label}</button
			>
		{/each}
	</div>

	<div class="tab-group" role="tablist" aria-label="Período">
		{#each PERIODS as option (option)}
			<button
				role="tab"
				aria-selected={period === option}
				class="tab-btn"
				class:active={period === option}
				onclick={() => onperiod(option)}>{option}</button
			>
		{/each}
	</div>
</div>

<style>
	/* Los dos grupos comparten forma: son la misma clase de control, y verlos
	   distintos haría pensar que uno de ellos navega a otro sitio. */
	.chart-controls {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.tab-group {
		display: flex;
		gap: 0.15rem;
		padding: 0.2rem;
		border: 1px solid var(--border);
		border-radius: 9px;
		flex-shrink: 0;
	}

	/*
	 * La misma forma que el conmutador de «Dónde está»: en el panel los dos
	 * viven a dos dedos uno de otro y son la misma clase de control. El activo
	 * era una pastilla ámbar, que en esta página es el color del valor de
	 * mercado y no el de «pestaña seleccionada».
	 */
	.tab-btn {
		padding: 0.35rem 0.75rem;
		border: none;
		border-radius: 7px;
		background: none;
		color: var(--text-muted);
		font-family: inherit;
		font-size: 0.8rem;
		cursor: pointer;
		transition:
			background 0.15s ease,
			color 0.15s ease;
	}

	.tab-btn.active {
		background: var(--panel-2, rgba(255, 255, 255, 0.08));
		color: var(--text);
	}

	.tab-btn:hover:not(.active) {
		color: var(--text);
	}
</style>
