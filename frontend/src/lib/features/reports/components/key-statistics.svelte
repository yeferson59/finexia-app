<script lang="ts">
	/** Métricas de riesgo derivadas del historial (drawdown y volatilidad). */
	import ReportPanel from './report-panel.svelte';
	import type { KeyStat } from '../reports';

	interface Props {
		statistics: KeyStat[];
	}

	let { statistics }: Props = $props();
</script>

<ReportPanel class="stats-card" title="Key Statistics">
	<div class="stats-list">
		{#if statistics.length > 0}
			{#each statistics as stat (stat.label)}
				<div class="stat-row">
					<p>{stat.label}</p>
					<p>{stat.value}</p>
				</div>
			{/each}
		{:else}
			<p class="empty-text">Sin datos</p>
		{/if}
	</div>
</ReportPanel>

<style>
	.stats-list {
		display: grid;
		gap: 0.45rem;
	}

	.stat-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: rgba(255, 255, 255, 0.022);
		padding: 0.6rem 0.75rem;
		border-radius: 8px;
	}

	.stat-row p {
		margin: 0;
		font-size: 0.8rem;
	}

	.stat-row p:first-child {
		color: rgba(236, 234, 229, 0.62);
	}

	.stat-row p:last-child {
		font-weight: 700;
		color: var(--amber-light);
	}
</style>
