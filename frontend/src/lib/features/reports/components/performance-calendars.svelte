<script lang="ts">
	/** Calendario de rentabilidad mensual, un panel por año. */
	import ReportPanel from './report-panel.svelte';
	import { MONTHS, performanceClass, type PerformanceCalendar } from '../reports';

	interface Props {
		calendars: PerformanceCalendar[];
	}

	let { calendars }: Props = $props();
</script>

<section class="analytics-grid">
	{#if calendars.length > 0}
		{#each calendars as calendar (calendar.year)}
			<ReportPanel class="calendar-card" title="Performance Calendar (%)" badge={calendar.year}>
				<div class="calendar-grid">
					{#each calendar.values as value, index (`${calendar.year}-${MONTHS[index]}`)}
						{#if value === null}
							<div class="month-cell null-cell">
								<p class="month">{MONTHS[index]}</p>
								<p class="percent">–</p>
							</div>
						{:else}
							<div class={`month-cell ${performanceClass(value)}`}>
								<p class="month">{MONTHS[index]}</p>
								<p class="percent">{value > 0 ? '+' : ''}{value.toFixed(1)}%</p>
							</div>
						{/if}
					{/each}
				</div>
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
	}

	.month-cell.strong-positive {
		background: rgba(34, 201, 126, 0.26);
		border-color: rgba(34, 201, 126, 0.45);
		color: var(--green);
	}

	.month-cell.positive {
		background: rgba(34, 201, 126, 0.18);
		border-color: rgba(34, 201, 126, 0.3);
		color: var(--green);
	}

	.month-cell.flat-positive {
		background: rgba(212, 145, 42, 0.2);
		border-color: rgba(212, 145, 42, 0.35);
		color: var(--amber-light);
	}

	.month-cell.negative {
		background: rgba(224, 90, 90, 0.16);
		border-color: rgba(224, 90, 90, 0.3);
		color: var(--red);
	}

	.month-cell.strong-negative {
		background: rgba(224, 90, 90, 0.26);
		border-color: rgba(224, 90, 90, 0.46);
		color: var(--red);
	}

	.month-cell.null-cell {
		background: rgba(255, 255, 255, 0.022);
		border-color: transparent;
		color: rgba(236, 234, 229, 0.28);
	}

	@media (max-width: 1024px) {
		.analytics-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
