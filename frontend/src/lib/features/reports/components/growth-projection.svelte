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
	 *
	 * Cada punto lleva además el porcentaje acumulado, y la cabecera la tasa
	 * anual de la que sale todo. El importe proyectado depende de cuánto haya
	 * hoy en la cuenta —un aporte de mañana lo mueve entero sin que la
	 * proyección haya cambiado de opinión—, así que quien quiera leer lo que
	 * esta gráfica de verdad extrapola tiene que poder leerlo en porcentaje.
	 */
	import ReportPanel from './report-panel.svelte';
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { PROJECTION_GUTTER, PROJECTION_MIN_DAYS, projectionCoordinates } from '../projection';
	import type { GrowthProjectionSeries } from '../projection';

	interface Props {
		/** `null` mientras el historial no dé para proyectar. */
		projection: GrowthProjectionSeries | null;
		/** Días de historial, para que el estado vacío diga cuánto falta. */
		historyDays: number;
	}

	let { projection, historyDays }: Props = $props();

	const missingDays = $derived(Math.max(PROJECTION_MIN_DAYS - historyDays, 0));

	const entries = $derived(projection?.entries ?? []);

	const points = $derived(projectionCoordinates(entries));
	const line = $derived(points.map((p) => `${p.x},${p.y}`).join(' '));
	// El área cierra contra la base del viewBox, de la esquina derecha a la izquierda.
	const area = $derived(
		points.length > 0 ? `${line} ${points[points.length - 1].x},230 ${PROJECTION_GUTTER},230` : ''
	);

	/** Lo que abarca el eje vertical, de la marca más baja a la más alta. */
	const span = $derived.by(() => {
		if (entries.length === 0) return 0;
		const values = entries.map((p) => p.value);
		return Math.max(...values) - Math.min(...values);
	});

	/** Marcas del eje vertical: el valor proyectado en cada línea de la rejilla. */
	const yTicks = $derived.by(() => {
		if (entries.length === 0) return [];
		const values = entries.map((p) => p.value);
		const min = Math.min(...values);
		const max = Math.max(...values);
		// `projectionCoordinates` reparte el rango entre y=230 (mínimo) y y=50.
		return Array.from({ length: 5 }, (_, i) => ({
			y: 50 + i * 45,
			value: max - ((max - min) * i) / 4
		}));
	});

	/**
	 * Decimales que hacen falta para que dos marcas contiguas no salgan iguales.
	 *
	 * Con una tasa cercana a cero las cinco marcas caben en menos de mil dólares,
	 * y redondear a miles imprimía «$89k» cinco veces: un eje sin escala, que es
	 * justo lo que la rejilla venía a resolver.
	 */
	function decimalsFor(unit: number): number {
		const steps = span / unit;
		if (steps >= 10) return 0;
		if (steps >= 1) return 1;
		return 2;
	}

	function fmtAbbrev(value: number): string {
		const abs = Math.abs(value);
		if (abs >= 1_000_000) return `$${(value / 1_000_000).toFixed(decimalsFor(1_000_000))}M`;
		if (abs >= 1_000) return `$${(value / 1_000).toFixed(decimalsFor(1_000))}k`;
		return `$${Math.round(value)}`;
	}

	function fmtFull(value: number): string {
		return '$' + new Intl.NumberFormat('es-CO', { maximumFractionDigits: 0 }).format(value);
	}
</script>

<ReportPanel
	class="projection-card"
	title="Proyección de crecimiento"
	badge={projection ? `${formatSignedPercent(projection.annualRatePct)} anual` : ''}
>
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
				<line
					x1={PROJECTION_GUTTER}
					y1={tick.y}
					x2="572"
					y2={tick.y}
					stroke="var(--border)"
					stroke-width="1"
				/>
				<text x={PROJECTION_GUTTER - 6} y={tick.y + 3.5} text-anchor="end" class="axis">
					{fmtAbbrev(tick.value)}
				</text>
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
					{fmtAbbrev(entries[i].value)}
				</text>
				<!-- El primer año es el punto de partida: su «+0,0 %» no dice nada y
				     encima empuja la etiqueta de dinero contra el borde del lienzo. -->
				{#if i > 0}
					<text x={point.x} y={point.y - 26} text-anchor="middle" class="pct-label">
						{formatSignedPercent(entries[i].returnPct)}
					</text>
				{/if}
				<text x={point.x} y="260" text-anchor="middle" class="axis">{point.period}</text>
			{/each}
		</svg>

		<table class="sr-only">
			<caption>Valor proyectado del portafolio por año y rentabilidad acumulada desde hoy</caption>
			<thead>
				<tr>
					<th scope="col">Año</th>
					<th scope="col">Valor proyectado</th>
					<th scope="col">Acumulado desde hoy</th>
				</tr>
			</thead>
			<tbody>
				{#each entries as entry (entry.period)}
					<tr>
						<td>{entry.period}</td>
						<td>{fmtFull(entry.value)}</td>
						<td>{formatSignedPercent(entry.returnPct)}</td>
					</tr>
				{/each}
			</tbody>
		</table>

		<p class="footnote">
			Extrapola tu rentabilidad anualizada ({formatSignedPercent(projection?.annualRatePct ?? 0)} al año)
			sobre el valor actual, sin contar aportes futuros. El porcentaje es lo que de verdad se proyecta;
			el importe se mueve con cada aporte que hagas. No es una previsión de mercado.
		</p>
	{:else}
		<div class="empty-chart">
			<p>Proyección disponible con al menos 6 meses de historial.</p>
			<!-- Decir cuánto falta ahorra volver cada semana a comprobarlo. -->
			<p class="countdown">
				{#if historyDays > 0}
					Llevas {historyDays}
					{historyDays === 1 ? 'día' : 'días'}; faltan {missingDays}.
				{:else}
					Empieza a contar con el primer cierre diario de tu cartera.
				{/if}
			</p>
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

	/* Más apagada que la de dinero: acompaña a la cifra principal, no compite
	   con ella. */
	.pct-label {
		fill: rgba(236, 234, 229, 0.5);
		font-size: 9.5px;
		font-family: var(--font-mono);
	}

	.footnote {
		margin: 0.5rem 0 0;
		font-size: 0.72rem;
		line-height: 1.5;
		color: var(--text-dim);
	}

	.empty-chart {
		padding: 3rem 2rem;
		text-align: center;
		color: var(--text-dim);
		font-size: 0.82rem;
		border: 1px dashed var(--border);
		border-radius: 8px;
		line-height: 1.6;
	}

	.empty-chart p {
		margin: 0;
	}

	/* `.empty-chart p` pone `margin: 0`; hace falta ganarle en especificidad. */
	.empty-chart .countdown {
		margin-top: 0.4rem;
		font-family: var(--font-mono);
		font-size: 0.72rem;
		color: var(--text-dim);
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
