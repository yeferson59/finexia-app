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
	 *
	 * El lienzo no sabe qué está dibujando: recibe dos series, cómo se llaman y
	 * cómo se escriben sus cifras. Así la misma gráfica sirve para el dinero y
	 * para la rentabilidad —quien decide es `portfolio-growth`— sin duplicar
	 * ejes, cursor ni tabla accesible.
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
		/** Cómo se llama el trazo principal (línea ámbar) en la vista activa. */
		primaryLabel: string;
		/** Cómo se llama el trazo de referencia (línea gris) en la vista activa. */
		secondaryLabel: string;
		/** Título de la tabla accesible; dice qué serie es esta. */
		caption: string;
		/**
		 * Altura contra la que cierra el relleno y en la que se marca la línea de
		 * referencia. `null` en la vista de dinero, donde el suelo es el borde del
		 * lienzo; `0` en la de porcentaje, donde el equilibrio es la cifra que hay
		 * que poder ver de un vistazo.
		 */
		baseline?: number | null;
		/** Etiqueta corta del eje vertical. */
		formatAbbrev: (value: number) => string;
		/** Etiqueta del eje horizontal; se abrevia y puede repetirse entre puntos. */
		formatDate: (iso: string) => string;
		/**
		 * Fecha completa de un punto. La tabla oculta necesita una fecha que
		 * identifique la fila sin ambigüedad: `formatDate` abrevia a mes y año en
		 * los rangos largos, así que allí varias filas se llamarían igual.
		 */
		formatFullDate: (iso: string) => string;
		/** Texto para el lector de pantalla y la tabla oculta. */
		formatValue: (value: number) => string;
		onactivate: (index: number | null) => void;
	}

	let {
		points,
		scale,
		active,
		primaryLabel,
		secondaryLabel,
		caption,
		baseline = null,
		formatAbbrev,
		formatDate,
		formatFullDate,
		formatValue,
		onactivate
	}: Props = $props();

	const { padL, padR, padT, plotH, svgW, svgH } = PLOT;

	let svgEl: SVGSVGElement | undefined = $state();

	const toX = (i: number) => toPlotX(i, points.length);
	const toY = (v: number) => toPlotY(v, scale);

	const mvPoints = $derived(points.map((p, i) => `${toX(i)},${toY(p.mv)}`).join(' '));
	const cbPoints = $derived(points.map((p, i) => `${toX(i)},${toY(p.cb)}`).join(' '));
	/*
	 * El relleno cierra contra la línea de referencia, no siempre contra el
	 * suelo: en porcentaje una racha negativa tiene que verse colgando por
	 * debajo del cero, y rellenarla hasta el borde la pintaba como si fuera
	 * terreno ganado.
	 */
	const floorY = $derived(
		baseline === null ? padT + plotH : Math.min(Math.max(toY(baseline), padT), padT + plotH)
	);
	const mvFill = $derived(
		points.length < 2 ? '' : `${mvPoints} ${toX(points.length - 1)},${floorY} ${toX(0)},${floorY}`
	);

	const yTicks = $derived(scale.ticks.map((value) => ({ value, y: toY(value) })));

	/*
	 * Hasta seis etiquetas, repartidas por igual entre el primer punto y el
	 * último. Tomarlas cada `n/6` dejaba la penúltima pegada a la última —con 70
	 * puntos caían a nueve de distancia— y los dos textos se solapaban.
	 */
	const X_LABELS = 6;
	const xLabels = $derived.by(() => {
		const n = points.length;
		if (n === 0) return [];
		if (n <= X_LABELS) return points.map((p, i) => ({ i, date: p.date }));

		const indices = Array.from({ length: X_LABELS }, (_, k) =>
			Math.round((k * (n - 1)) / (X_LABELS - 1))
		);
		return [...new Set(indices)].map((i) => ({ i, date: points[i].date }));
	});

	const activePoint = $derived(active === null ? null : (points[active] ?? null));

	/*
	 * El cursor recorre índices de la serie, que es justo lo que describe el
	 * patrón `slider` de ARIA: el lector de pantalla anuncia `aria-valuetext` en
	 * cada flecha sin necesidad de una región `aria-live` que lo repita.
	 */
	const valueText = $derived(
		activePoint
			? `${formatDate(activePoint.date)}: ${primaryLabel} ${formatValue(activePoint.mv)}, ${secondaryLabel} ${formatValue(activePoint.cb)}`
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

		<!-- La línea de equilibrio, más marcada que la rejilla: en porcentaje es la
		     frontera entre ganar y perder, y no puede confundirse con una marca más. -->
		{#if baseline !== null}
			<line x1={padL} y1={floorY} x2={svgW - padR} y2={floorY} class="baseline" />
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

	<!-- Los mismos datos que dibuja el SVG, con las dos series y la fecha
	     completa: es la única vía de acceso para quien no puede leer la gráfica,
	     así que no puede quedarse corta respecto a lo que promete el título.

	     El `sr-only` va en el div y no en la tabla: para una tabla, `height` es
	     un mínimo y no un máximo, así que las reglas de ocultación no la
	     recortaban y sus 1.700 px de filas se sumaban al alto del documento.
	     La página terminaba con una franja vacía del tamaño de la serie. -->
	<div class="sr-only">
		<table>
			<caption>{caption}</caption>
			<thead>
				<tr>
					<th scope="col">Fecha</th>
					<th scope="col">{primaryLabel}</th>
					<th scope="col">{secondaryLabel}</th>
				</tr>
			</thead>
			<tbody>
				{#each points as point (point.date)}
					<tr>
						<th scope="row"><time datetime={point.date}>{formatFullDate(point.date)}</time></th>
						<td>{formatValue(point.mv)}</td>
						<td>{formatValue(point.cb)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
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
		fill: var(--text-dim);
		font-size: 9px;
		font-family: var(--font-mono);
	}

	.baseline {
		stroke: rgba(236, 234, 229, 0.35);
		stroke-width: 1.25;
	}

	/*
	 * El capital invertido, en frío contra el ámbar del valor de mercado: el
	 * dinero que se puso y lo que el mercado hizo con él. Era un gris sin
	 * identidad que se confundía con la rejilla.
	 */
	.line-cost {
		fill: none;
		stroke: var(--cost);
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

	/* Anillo del color del fondo: es lo que mantiene el punto legible donde
	   cruza la línea, en vez de un borde dibujado alrededor. */
	.cursor-dot {
		stroke: var(--bg);
		stroke-width: 2;
	}

	.cursor-dot.value {
		fill: var(--amber-light);
	}

	.cursor-dot.cost {
		fill: var(--cost);
	}

	/*
	 * En pantallas estrechas el lienzo se encoge y con él el texto del SVG, que
	 * va en unidades del viewBox: a 370 px de ancho las marcas de los ejes
	 * quedaban en unos 5,5 px reales. Subir el cuerpo aquí las devuelve al
	 * borde de lo legible sin tocar la geometría, que es compartida.
	 */
	@media (max-width: 600px) {
		.axis {
			font-size: 12px;
		}
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
