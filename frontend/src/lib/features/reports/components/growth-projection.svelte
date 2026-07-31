<script lang="ts">
	/**
	 * Proyección a cinco años extrapolando el CAGR del historial.
	 *
	 * La gráfica es un SVG propio (el proyecto no tiene librería de charts): las
	 * coordenadas las calcula `projectionCoordinates`, en `reports.ts`.
	 */
	import ReportPanel from './report-panel.svelte';
	import { projectionCoordinates, type GrowthProjectionEntry } from '../reports';

	interface Props {
		projection: GrowthProjectionEntry[];
	}

	let { projection }: Props = $props();

	const points = $derived(projectionCoordinates(projection));
	const line = $derived(points.map((p) => `${p.x},${p.y}`).join(' '));
	// El área cierra contra la base del viewBox, de la esquina derecha a la izquierda.
	const area = $derived(`${line} 560,230 40,230`);
</script>

<ReportPanel class="projection-card" title="Growth Projection">
	{#if points.length > 0}
		<svg class="projection-chart" viewBox="0 0 600 280" preserveAspectRatio="xMidYMid meet">
			<defs>
				<linearGradient id="projectionGradient" x1="0%" y1="0%" x2="0%" y2="100%">
					<stop offset="0%" style="stop-color: var(--amber); stop-opacity: 0.25" />
					<stop offset="100%" style="stop-color: var(--amber); stop-opacity: 0" />
				</linearGradient>
			</defs>
			{#each Array.from({ length: 5 }) as _, i (i)}
				<line
					x1="40"
					y1={35 + i * 50}
					x2="560"
					y2={35 + i * 50}
					stroke="var(--border)"
					stroke-width="1"
				/>
			{/each}
			<polyline
				points={line}
				fill="none"
				stroke="var(--amber)"
				stroke-width="3"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
			<polygon points={area} fill="url(#projectionGradient)" />
			{#each points as point (point.period)}
				<circle
					cx={point.x}
					cy={point.y}
					r="4"
					fill="var(--amber-light)"
					stroke="rgba(255, 255, 255, 0.022)"
					stroke-width="2"
				/>
				<text
					x={point.x}
					y="260"
					text-anchor="middle"
					fill="rgba(236, 234, 229,0.56)"
					font-size="12"
				>
					{point.period}
				</text>
			{/each}
		</svg>
	{:else}
		<div class="empty-chart">
			<p>Proyección disponible con al menos 6 meses de historial.</p>
		</div>
	{/if}
</ReportPanel>

<style>
	.projection-chart {
		width: 100%;
		min-height: 280px;
		display: block;
	}

	.empty-chart {
		padding: 3rem 2rem;
		text-align: center;
		color: rgba(236, 234, 229, 0.45);
		font-size: 0.82rem;
		border: 1px dashed var(--border);
		border-radius: 8px;
		line-height: 1.6;
	}

	.empty-chart p {
		margin: 0;
	}
</style>
