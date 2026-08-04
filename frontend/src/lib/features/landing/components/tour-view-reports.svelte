<script lang="ts">
	/* Maqueta de `/dashboard/reports`: calendario, estadísticas y descargas. */
	import {
		TOUR_CALENDAR,
		TOUR_KEY_STATS,
		TOUR_REPORTS,
		tourPerformanceClass
	} from '../product-tour';

	/* Coma decimal, como el resto de cifras de la landing y del dashboard. */
	const fmt = (value: number | null) =>
		value === null ? '—' : `${value > 0 ? '+' : ''}${value.toFixed(1).replace('.', ',')}`;
</script>

<div class="reports">
	<div class="mk-card">
		<div class="cal-top">
			<div>
				<div class="mk-eyebrow">Rendimiento</div>
				<div class="mk-title">Rentabilidad mes a mes</div>
			</div>
			<span class="mk-pill">2026</span>
		</div>
		<div class="calendar">
			{#each TOUR_CALENDAR as cell (cell.month)}
				<div class="cell {tourPerformanceClass(cell.value)}">
					<span class="m">{cell.month}</span>
					<span class="v">{fmt(cell.value)}</span>
				</div>
			{/each}
		</div>
	</div>

	<div class="bottom">
		<div class="mk-card">
			<div class="mk-eyebrow">Estadísticas clave</div>
			<div class="stats">
				{#each TOUR_KEY_STATS as stat (stat.label)}
					<div class="stat-row">
						<span>{stat.label}</span>
						<b class:mk-up={stat.value.startsWith('+')} class:mk-dn={stat.value.startsWith('−')}>
							{stat.value}
						</b>
					</div>
				{/each}
			</div>
		</div>

		<div class="mk-card">
			<div class="mk-eyebrow">Descargas</div>
			<div class="downloads">
				{#each TOUR_REPORTS as report (report.title)}
					<div class="dl">
						<span class="fmt">{report.format}</span>
						<span class="dl-title">{report.title}</span>
						<span class="dl-btn">Descargar</span>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>

<style>
	.reports {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.cal-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
		padding-bottom: 14px;
		border-bottom: 1px solid var(--border);
	}

	.calendar {
		display: grid;
		grid-template-columns: repeat(12, minmax(0, 1fr));
		gap: 4px;
		margin-top: 14px;
	}

	.cell {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		padding: 9px 2px;
		border-radius: 5px;
		border: 1px solid var(--border);
		font-family: var(--font-mono);
	}

	.cell .m {
		font-size: 8px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.cell .v {
		font-size: 9.5px;
		font-weight: 600;
		color: var(--text-muted);
	}

	.mk-cal-strong-up {
		background: rgba(34, 201, 126, 0.16);
		border-color: rgba(34, 201, 126, 0.3);
	}

	.mk-cal-strong-up .v {
		color: var(--green);
	}

	.mk-cal-up {
		background: rgba(34, 201, 126, 0.07);
	}

	.mk-cal-up .v {
		color: rgba(34, 201, 126, 0.85);
	}

	.mk-cal-down {
		background: rgba(224, 90, 90, 0.08);
	}

	.mk-cal-down .v,
	.mk-cal-strong-down .v {
		color: var(--red);
	}

	.mk-cal-strong-down {
		background: rgba(224, 90, 90, 0.16);
		border-color: rgba(224, 90, 90, 0.3);
	}

	.mk-cal-empty {
		opacity: 0.42;
		border-style: dashed;
	}

	.bottom {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr);
		gap: 10px;
	}

	.stats {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
	}

	.stat-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		padding: 8px 0;
		border-top: 1px solid var(--border);
		font-size: 11px;
		color: var(--text-muted);
	}

	.stat-row:first-child {
		border-top: none;
	}

	.stat-row b {
		font-family: var(--font-mono);
		font-size: 11.5px;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.downloads {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 7px;
	}

	.dl {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 10px;
		border: 1px solid var(--border);
		border-radius: 7px;
		background: var(--surface);
	}

	.fmt {
		flex-shrink: 0;
		padding: 2px 7px;
		border-radius: 999px;
		background: var(--border-strong);
		font-family: var(--font-mono);
		font-size: 8.5px;
		font-weight: 700;
		letter-spacing: 0.06em;
		color: var(--amber-light);
	}

	.dl-title {
		flex: 1;
		min-width: 0;
		font-size: 11px;
		color: var(--text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.dl-btn {
		flex-shrink: 0;
		font-family: var(--font-mono);
		font-size: 9px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	@media (max-width: 700px) {
		.calendar {
			grid-template-columns: repeat(6, minmax(0, 1fr));
		}

		.bottom {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
