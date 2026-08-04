<script lang="ts">
	/** Métricas de riesgo derivadas del historial (drawdown y volatilidad). */
	import ReportPanel from './report-panel.svelte';
	import type { KeyStat } from '../reports';

	interface Props {
		statistics: KeyStat[];
	}

	let { statistics }: Props = $props();
</script>

<ReportPanel class="stats-card" title="Estadísticas clave">
	{#if statistics.length > 0}
		<!-- Etiqueta y valor son un par: `dl` los relaciona, dos `p` no. -->
		<dl class="stats-list">
			{#each statistics as stat (stat.label)}
				<div class="stat-row">
					<dt>{stat.label}</dt>
					<dd>{stat.value}</dd>
				</div>
			{/each}
		</dl>
	{:else}
		<p class="empty-text">Sin datos</p>
	{/if}
</ReportPanel>

<style>
	.stats-list {
		display: grid;
		gap: 0.45rem;
		margin: 0;
	}

	.stat-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		background: rgba(255, 255, 255, 0.022);
		padding: 0.6rem 0.75rem;
		border-radius: 8px;
	}

	.stat-row dt,
	.stat-row dd {
		margin: 0;
		font-size: 0.8rem;
	}

	.stat-row dt {
		color: rgba(236, 234, 229, 0.62);
	}

	.stat-row dd {
		font-family: var(--font-mono);
		font-weight: 700;
		font-variant-numeric: tabular-nums;
		color: var(--amber-light);
	}
</style>
