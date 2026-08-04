<script lang="ts">
	/*
	 * Proyección a cinco años extrapolando el CAGR del historial.
	 *
	 * La gráfica es un SVG propio (el proyecto no tiene librería de charts): las
	 * coordenadas las calcula `projectionCoordinates`, en `reports.ts`.
	 *
	 * Antes el área se pintaba encima de la línea y la tapaba, y el eje vertical
	 * no tenía ni una cifra: la curva subía, pero no se sabía hasta dónde. Ahora
	 * el relleno va debajo, cada año lleva su valor y la misma tabla está
	 * disponible para el lector de pantalla.
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
	const area = $derived(
		points.length > 0 ? `${line} ${points[points.length - 1].x},230 40,230` : ''
	);

	/** Marcas del eje vertical: el valor proyectado en cada línea de la rejilla. */
	const yTicks = $derived.by(() => {
		if (projection.length === 0) return [];
		const values = projection.map((p) => p.value);
		const min = Math.min(...values);
		const max = Math.max(...values);
		// `projectionCoordinates` reparte el rango entre y=230 (mínimo) y y=50.
		return Array.from({ length: 5 }, (_, i) => ({
			y: 50 + i * 45,
			value: max - ((max - min) * i) / 4
		}));
	});

	function fmtAbbrev(value: number): string {
		const abs = Math.abs(value);
		if (abs >= 1_000_000) return `$${(value / 1_000_000).toFixed(1)}M`;
		if (abs >= 1_000) return `$${Math.round(value / 1_000)}k`;
		return `$${Math.round(value)}`;
	}

	function fmtFull(value: number): string {
		return '$' + new Intl.NumberFormat('es-CO', { maximumFractionDigits: 0 }).format(value);
	}
</script>

<ReportPanel class="projection-card" title="Proyección de crecimiento">
	{#if points.length > 0}
		<svg
			class="projection-chart"
			viewBox="0 0 600 280"
			preserveAspectRatio="xMidYMid meet"
			aria-hidden="true"
		>
			<defs>
				<linearGradient id="projectionGradient" x1="0%" y1="0%" x2="0%" y2="100%">
					<stop offset="0%" style="stop-color: var(--amber); stop-opacity: 0.25" />
					<stop offset="100%" style="stop-color: var(--amber); stop-opacity: 0" />
				</linearGradient>
			</defs>

			{#each yTicks as tick (tick.y)}
				<line x1="40" y1={tick.y} x2="560" y2={tick.y} stroke="var(--border)" stroke-width="1" />
				<text x="34" y={tick.y + 3.5} text-anchor="end" class="axis">{fmtAbbrev(tick.value)}</text>
			{/each}

			<!-- El relleno primero: si va después, tapa la línea que debe destacar. -->
			<polygon points={area} fill="url(#projectionGradient)" />
			<polyline
				points={line}
				fill="none"
				stroke="var(--amber)"
				stroke-width="3"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>

			{#each points as point, i (point.period)}
				<circle
					cx={point.x}
					cy={point.y}
					r="4"
					fill="var(--amber-light)"
					stroke="#08090a"
					stroke-width="2"
				/>
				<text x={point.x} y={point.y - 13} text-anchor="middle" class="value-label">
					{fmtAbbrev(projection[i].value)}
				</text>
				<text x={point.x} y="260" text-anchor="middle" class="axis">{point.period}</text>
			{/each}
		</svg>

		<table class="sr-only">
			<caption>Valor proyectado del portafolio por año</caption>
			<thead>
				<tr><th scope="col">Año</th><th scope="col">Valor proyectado</th></tr>
			</thead>
			<tbody>
				{#each projection as entry (entry.period)}
					<tr><td>{entry.period}</td><td>{fmtFull(entry.value)}</td></tr>
				{/each}
			</tbody>
		</table>

		<p class="footnote">
			Extrapolación del crecimiento anual compuesto de tu historial. No es una previsión de mercado.
		</p>
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

	.axis {
		fill: rgba(236, 234, 229, 0.5);
		font-size: 11px;
		font-family: var(--font-mono);
	}

	.value-label {
		fill: var(--amber-light);
		font-size: 11px;
		font-weight: 600;
		font-family: var(--font-mono);
	}

	.footnote {
		margin: 0.5rem 0 0;
		font-size: 0.72rem;
		line-height: 1.5;
		color: rgba(236, 234, 229, 0.42);
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

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
		border: 0;
	}
</style>
