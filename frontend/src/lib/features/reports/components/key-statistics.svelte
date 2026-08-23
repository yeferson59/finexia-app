<script lang="ts">
	/*
	 * Estadísticas del historial: rendimiento, riesgo y de cuánto historial
	 * salen los dos.
	 *
	 * Antes eran dos filas —máxima caída y volatilidad— y la segunda decía `N/A`
	 * sin explicar por qué, con el resto de la serie sin usar. Ahora los tres
	 * bloques reparten doce métricas, cada una lleva en `title` qué mide, y la
	 * que no se puede calcular dice a la vista qué historial le falta.
	 *
	 * Una métrica que sí sale puede traer `note`: el reparo que hay que leer con
	 * la cifra —una estimación de margen amplio, un mes incompleto— y que en el
	 * `title` no vería nadie.
	 */
	import ReportPanel from './report-panel.svelte';
	import { UNAVAILABLE, type KeyStatGroup } from '../reports';

	interface Props {
		groups: KeyStatGroup[];
	}

	let { groups }: Props = $props();
</script>

<ReportPanel class="stats-card" title="Estadísticas clave">
	{#if groups.length > 0}
		{#each groups as group (group.title)}
			<section class="stat-group">
				<h3>{group.title}</h3>
				<!-- Etiqueta y valor son un par: `dl` los relaciona, dos `p` no. -->
				<dl class="stats-list">
					{#each group.stats as stat (stat.label)}
						<div class="stat-tile" title={stat.hint}>
							<dt>{stat.label}</dt>
							<dd class={stat.tone ?? 'neutral'}>{stat.value}</dd>
							{#if stat.value === UNAVAILABLE && stat.hint}
								<!-- Sin esto, `N/A` no distingue «falta historial» de «algo se rompió». -->
								<p class="reason">{stat.hint}</p>
							{:else if stat.note}
								<!-- El reparo de una cifra que sí sale va a la vista, no al `title`:
								     un Sharpe estimado con tres meses se lee mal sin él. -->
								<p class="reason">{stat.note}</p>
							{/if}
						</div>
					{/each}
				</dl>
			</section>
		{/each}
	{:else}
		<p class="empty-text">
			Sin datos históricos todavía. Las estadísticas aparecen con el primer cierre diario de tu
			cartera.
		</p>
	{/if}
</ReportPanel>

<style>
	.stat-group + .stat-group {
		margin-top: 1.1rem;
	}

	.stat-group h3 {
		margin: 0 0 0.5rem;
		font-family: var(--font-mono);
		font-size: 0.62rem;
		font-weight: 600;
		letter-spacing: 0.09em;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.45);
	}

	.stats-list {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		gap: 0.45rem;
		margin: 0;
	}

	.stat-tile {
		display: grid;
		align-content: start;
		gap: 0.2rem;
		background: rgba(255, 255, 255, 0.022);
		padding: 0.6rem 0.7rem;
		border-radius: 8px;
	}

	.stat-tile dt,
	.stat-tile dd {
		margin: 0;
	}

	.stat-tile dt {
		font-size: 0.68rem;
		line-height: 1.3;
		color: rgba(236, 234, 229, 0.62);
	}

	.stat-tile dd {
		font-family: var(--font-mono);
		font-size: 0.86rem;
		font-weight: 700;
		font-variant-numeric: tabular-nums;
		color: var(--amber-light);
	}

	.stat-tile dd.up {
		color: var(--green);
	}

	.stat-tile dd.down {
		color: var(--red);
	}

	.reason {
		margin: 0.1rem 0 0;
		font-size: 0.62rem;
		line-height: 1.4;
		color: rgba(236, 234, 229, 0.38);
	}
</style>
