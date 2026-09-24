<script lang="ts">
	/**
	 * La gráfica de un fondo: su valor de unidad en el tiempo. Una sola serie, así
	 * que no lleva leyenda —el título del diálogo dice qué es— y la línea va en el
	 * ámbar de la aplicación, como la del crecimiento del panel.
	 *
	 * El cursor sigue al ratón o a las flechas y fija el punto más cercano; la
	 * cifra sale en una etiqueta encima. Para quien no ve el SVG, los mismos
	 * puntos están en una tabla oculta a la vista.
	 */
	import { FUND_CHART, fundChartGeometry, nearestPoint, type FundChartPoint } from '../chart';

	interface Props {
		points: FundChartPoint[];
		/** Qué es la serie, para la tabla oculta y el lector de pantalla. */
		caption: string;
		formatValue: (value: number) => string;
		formatAxis: (value: number) => string;
		formatDate: (iso: string) => string;
	}

	let { points, caption, formatValue, formatAxis, formatDate }: Props = $props();

	const { width, height, padL, padR, padT, padB } = FUND_CHART;

	const geometry = $derived(fundChartGeometry(points));

	const line = $derived(
		points.map((p) => `${geometry.x(p.date)},${geometry.y(p.value)}`).join(' ')
	);
	const area = $derived(
		points.length < 2
			? ''
			: `${line} ${geometry.x(points[points.length - 1].date)},${height - padB} ${geometry.x(points[0].date)},${height - padB}`
	);

	let active = $state<number | null>(null);
	let svgEl: SVGSVGElement | undefined = $state();

	const activePoint = $derived(active === null ? null : (points[active] ?? null));
	const last = $derived(points[points.length - 1]);

	function onPointer(event: PointerEvent) {
		if (!svgEl) return;
		const box = svgEl.getBoundingClientRect();
		active = nearestPoint(points, geometry, ((event.clientX - box.left) / box.width) * width);
	}

	function onKeyDown(event: KeyboardEvent) {
		const end = points.length - 1;

		if (event.key === 'ArrowRight') active = Math.min(end, (active ?? -1) + 1);
		else if (event.key === 'ArrowLeft') active = Math.max(0, (active ?? points.length) - 1);
		else if (event.key === 'Home') active = 0;
		else if (event.key === 'End') active = end;
		else if (event.key === 'Escape') active = null;
		else return;

		event.preventDefault();
	}

	const valueText = $derived(
		activePoint
			? `${formatDate(activePoint.date)}: ${formatValue(activePoint.value)}`
			: 'Ningún punto seleccionado'
	);

	/* La etiqueta del cursor, en porcentaje del ancho para que siga al lienzo
	   cuando se encoge; se ancla hacia dentro en los extremos. */
	const tipLeft = $derived(activePoint ? (geometry.x(activePoint.date) / width) * 100 : 0);
</script>

<div
	class="chart-hit"
	role="slider"
	tabindex="0"
	aria-label="Recorrer la gráfica del fondo"
	aria-valuemin={0}
	aria-valuemax={Math.max(0, points.length - 1)}
	aria-valuenow={active ?? 0}
	aria-valuetext={valueText}
	onpointermove={onPointer}
	onpointerleave={() => (active = null)}
	onkeydown={onKeyDown}
	onblur={() => (active = null)}
>
	{#if activePoint}
		<div
			class="tip"
			class:start={tipLeft < 20}
			class:end={tipLeft > 80}
			style:left="{tipLeft}%"
			aria-hidden="true"
		>
			<strong>{formatValue(activePoint.value)}</strong>
			<span>{formatDate(activePoint.date)}</span>
		</div>
	{/if}

	<svg
		bind:this={svgEl}
		class="chart"
		viewBox="0 0 {width} {height}"
		preserveAspectRatio="xMidYMid meet"
		aria-hidden="true"
	>
		{#each geometry.yTicks as tick (tick)}
			<line x1={padL} y1={geometry.y(tick)} x2={width - padR} y2={geometry.y(tick)} class="grid" />
			<text x={padL - 8} y={geometry.y(tick) + 3.5} text-anchor="end" class="axis">
				{formatAxis(tick)}
			</text>
		{/each}

		{#if area}
			<polygon points={area} class="area" />
		{/if}
		<polyline points={line} class="line" />

		<circle cx={geometry.x(last.date)} cy={geometry.y(last.value)} r="4" class="dot" />

		{#each geometry.xTicks as date, i (date)}
			<text
				x={geometry.x(date)}
				y={height - padB + 20}
				text-anchor={geometry.xTicks.length === 1
					? 'middle'
					: i === 0
						? 'start'
						: i === geometry.xTicks.length - 1
							? 'end'
							: 'middle'}
				class="axis"
			>
				{formatDate(date)}
			</text>
		{/each}

		{#if activePoint}
			<line
				x1={geometry.x(activePoint.date)}
				y1={padT}
				x2={geometry.x(activePoint.date)}
				y2={height - padB}
				class="cursor"
			/>
			<circle
				cx={geometry.x(activePoint.date)}
				cy={geometry.y(activePoint.value)}
				r="4.5"
				class="dot"
			/>
		{/if}
	</svg>

	<div class="sr-only">
		<table>
			<caption>{caption}</caption>
			<thead>
				<tr>
					<th scope="col">Fecha</th>
					<th scope="col">Valor</th>
				</tr>
			</thead>
			<tbody>
				{#each points as point (point.date)}
					<tr>
						<th scope="row"><time datetime={point.date}>{formatDate(point.date)}</time></th>
						<td>{formatValue(point.value)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>

<style>
	.chart-hit {
		position: relative;
		padding-top: 2.4rem;
		touch-action: pan-y;
	}

	.chart-hit:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 4px;
		border-radius: 8px;
	}

	.chart {
		display: block;
		width: 100%;
	}

	.grid {
		stroke: var(--border);
		stroke-width: 1;
	}

	.axis {
		fill: var(--text-dim);
		font-family: var(--font-mono);
		font-size: 9px;
	}

	/* Un velo del color de la serie, no un bloque: la línea es el dato. */
	.area {
		fill: var(--amber);
		fill-opacity: 0.1;
		stroke: none;
	}

	.line {
		fill: none;
		stroke: var(--amber);
		stroke-width: 2;
		stroke-linecap: round;
		stroke-linejoin: round;
	}

	/* Anillo del color del fondo: mantiene el punto legible donde cruza la línea. */
	.dot {
		fill: var(--amber-light);
		stroke: var(--bg);
		stroke-width: 2;
	}

	.cursor {
		stroke: rgba(236, 234, 229, 0.35);
		stroke-width: 1;
	}

	/* La cifra manda y la fecha acompaña; el texto nunca va en el color de la serie. */
	.tip {
		position: absolute;
		top: 0;
		display: grid;
		gap: 0.1rem;
		padding: 0.3rem 0.55rem;
		border: 1px solid var(--border-strong);
		border-radius: 7px;
		background: var(--bg);
		transform: translateX(-50%);
		pointer-events: none;
		white-space: nowrap;
	}

	.tip.start {
		transform: none;
	}

	.tip.end {
		transform: translateX(-100%);
	}

	.tip strong {
		font-family: var(--font-mono);
		font-size: 0.82rem;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.tip span {
		font-size: 0.7rem;
		color: var(--text-muted);
	}

	@media (max-width: 600px) {
		.axis {
			font-size: 12px;
		}
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
