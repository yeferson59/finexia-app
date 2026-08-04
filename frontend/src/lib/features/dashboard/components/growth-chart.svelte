<script lang="ts">
	/*
	 * Lienzo de la gráfica de crecimiento: ejes, líneas y el cursor que sigue al
	 * ratón o a las flechas del teclado.
	 *
	 * La gráfica antes solo se miraba: no había forma de saber cuánto valía un
	 * punto concreto. Ahora el cursor fija un punto y la tarjeta que lo envuelve
	 * pinta sus cifras. Para quien no ve el SVG, los mismos datos están en una
	 * tabla oculta a la vista pero disponible para el lector de pantalla, así que
	 * la información no depende de poder apuntar con el ratón.
	 */
	import {
		PLOT,
		nearestIndex,
		toPlotX,
		toPlotY,
		type GrowthPoint,
		type GrowthScale
	} from '../dashboard';

	interface Props {
		points: GrowthPoint[];
		scale: GrowthScale;
		/** Índice fijado por el usuario; `null` cuando no hay ninguno. */
		active: number | null;
		formatAbbrev: (value: number) => string;
		formatDate: (iso: string) => string;
		/** Texto para el lector de pantalla y la tabla oculta. */
		formatMoney: (value: number) => string;
		onactivate: (index: number | null) => void;
	}

	let { points, scale, active, formatAbbrev, formatDate, formatMoney, onactivate }: Props =
		$props();

	const { padL, padR, padT, plotH, svgW, svgH } = PLOT;

	let svgEl: SVGSVGElement | undefined = $state();

	const toX = (i: number) => toPlotX(i, points.length);
	const toY = (v: number) => toPlotY(v, scale);

	const mvPoints = $derived(points.map((p, i) => `${toX(i)},${toY(p.mv)}`).join(' '));
	const cbPoints = $derived(points.map((p, i) => `${toX(i)},${toY(p.cb)}`).join(' '));
	const mvFill = $derived(
		points.length < 2
			? ''
			: `${mvPoints} ${toX(points.length - 1)},${padT + plotH} ${toX(0)},${padT + plotH}`
	);

	const yTicks = $derived(scale.ticks.map((value) => ({ value, y: toY(value) })));

	// Como mucho seis etiquetas en el eje horizontal: más se solapan.
	const xLabels = $derived(
		points.length === 0
			? []
			: points
					.map((p, i) => ({ i, date: p.date }))
					.filter(({ i }) => {
						const n = points.length;
						if (n <= 6) return true;
						return i % Math.ceil(n / 6) === 0 || i === n - 1;
					})
	);

	const activePoint = $derived(active === null ? null : (points[active] ?? null));

	/*
	 * El cursor recorre índices de la serie, que es justo lo que describe el
	 * patrón `slider` de ARIA: el lector de pantalla anuncia `aria-valuetext` en
	 * cada flecha sin necesidad de una región `aria-live` que lo repita.
	 */
	const valueText = $derived(
		activePoint
			? `${formatDate(activePoint.date)}: valor ${formatMoney(activePoint.mv)}, invertido ${formatMoney(activePoint.cb)}`
			: 'Ningún punto seleccionado'
	);

	/** Pasa de píxeles de pantalla a coordenadas del viewBox antes de buscar el punto. */
	function indexFromEvent(event: PointerEvent): number {
		if (!svgEl) return 0;
		const box = svgEl.getBoundingClientRect();
		const x = ((event.clientX - box.left) / box.width) * svgW;
		return nearestIndex(x, points.length);
	}

	function onKeyDown(event: KeyboardEvent) {
		const last = points.length - 1;
		let next: number | null;

		if (event.key === 'ArrowRight') next = Math.min(last, (active ?? -1) + 1);
		else if (event.key === 'ArrowLeft') next = Math.max(0, (active ?? points.length) - 1);
		else if (event.key === 'Home') next = 0;
		else if (event.key === 'End') next = last;
		else if (event.key === 'Escape') next = null;
		else return;

		event.preventDefault();
		onactivate(next);
	}
</script>

<div
	class="chart-hit"
	role="slider"
	tabindex="0"
	aria-label="Recorrer la gráfica de crecimiento del portafolio"
	aria-valuemin={0}
	aria-valuemax={Math.max(0, points.length - 1)}
	aria-valuenow={active ?? 0}
	aria-valuetext={valueText}
	onpointermove={(event) => onactivate(indexFromEvent(event))}
	onpointerleave={() => onactivate(null)}
	onkeydown={onKeyDown}
	onblur={() => onactivate(null)}
>
	<svg
		bind:this={svgEl}
		class="chart"
		viewBox="0 0 {svgW} {svgH}"
		preserveAspectRatio="xMidYMid meet"
		aria-hidden="true"
	>
		<defs>
			<linearGradient id="growthGradient" x1="0%" y1="0%" x2="0%" y2="100%">
				<stop offset="0%" style="stop-color: var(--amber); stop-opacity: 0.18" />
				<stop offset="100%" style="stop-color: var(--amber); stop-opacity: 0" />
			</linearGradient>
		</defs>

		{#each yTicks as tick (tick.value)}
			<line x1={padL} y1={tick.y} x2={svgW - padR} y2={tick.y} class="grid" />
			<text x={padL - 8} y={tick.y + 3.5} text-anchor="end" class="axis">
				{formatAbbrev(tick.value)}
			</text>
		{/each}

		{#if mvFill}
			<polygon points={mvFill} fill="url(#growthGradient)" />
		{/if}

		<polyline points={cbPoints} class="line-cost" />
		<polyline points={mvPoints} class="line-value" />

		{#if points.length > 0}
			{@const lastIndex = points.length - 1}
			<circle cx={toX(lastIndex)} cy={toY(points[lastIndex].mv)} r="4" class="last-dot" />
		{/if}

		<!-- Las etiquetas de los extremos se anclan hacia dentro: centradas, la
		     primera y la última se salían del lienzo y quedaban cortadas. -->
		{#each xLabels as { i, date } (i)}
			<text
				x={toX(i)}
				y={padT + plotH + 20}
				text-anchor={i === 0 ? 'start' : i === points.length - 1 ? 'end' : 'middle'}
				class="axis"
			>
				{formatDate(date)}
			</text>
		{/each}

		{#if active !== null && activePoint}
			<line x1={toX(active)} y1={padT} x2={toX(active)} y2={padT + plotH} class="cursor" />
			<circle cx={toX(active)} cy={toY(activePoint.cb)} r="3.5" class="cursor-dot cost" />
			<circle cx={toX(active)} cy={toY(activePoint.mv)} r="4.5" class="cursor-dot value" />
		{/if}
	</svg>

	<!-- Los mismos datos, para quien no puede leer el SVG. -->
	<table class="sr-only">
		<caption>Valor de mercado y capital invertido del portafolio, por fecha</caption>
		<thead>
			<tr><th scope="col">Fecha</th><th scope="col">Valor de mercado</th></tr>
		</thead>
		<tbody>
			{#each points as point (point.date)}
				<tr>
					<td>{formatDate(point.date)}</td>
					<td>{formatMoney(point.mv)}</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<style>
	.chart-hit {
		position: relative;
		touch-action: pan-y;
	}

	.chart-hit:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 4px;
		border-radius: 8px;
	}

	.chart {
		width: 100%;
		display: block;
	}

	.grid {
		stroke: var(--border);
		stroke-width: 1;
	}

	.axis {
		fill: rgba(236, 234, 229, 0.42);
		font-size: 9px;
		font-family: var(--font-mono);
	}

	.line-cost {
		fill: none;
		stroke: rgba(236, 234, 229, 0.3);
		stroke-width: 1.5;
		stroke-dasharray: 6 4;
		stroke-linecap: round;
		stroke-linejoin: round;
	}

	.line-value {
		fill: none;
		stroke: var(--amber);
		stroke-width: 2.5;
		stroke-linecap: round;
		stroke-linejoin: round;
	}

	.last-dot {
		fill: var(--amber-light);
		stroke: rgba(0, 0, 0, 0.35);
		stroke-width: 1.5;
	}

	.cursor {
		stroke: rgba(236, 234, 229, 0.35);
		stroke-width: 1;
		stroke-dasharray: 3 3;
	}

	.cursor-dot {
		stroke: #08090a;
		stroke-width: 2;
	}

	.cursor-dot.value {
		fill: var(--amber-light);
	}

	.cursor-dot.cost {
		fill: rgba(236, 234, 229, 0.6);
	}

	/* Fuera de la vista pero en el árbol de accesibilidad. */
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
