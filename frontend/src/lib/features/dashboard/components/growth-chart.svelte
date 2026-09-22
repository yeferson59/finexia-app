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
		/**
		 * Índice del primer punto proyectado; `null` cuando no se proyecta nada.
		 *
		 * Los puntos vienen en una sola serie —el historial y luego la proyección—
		 * para que el eje, el cursor y la escala los traten igual. Lo único que
		 * cambia a partir de aquí es cómo se dibujan: la línea pasa a discontinua
		 * y el relleno se corta, porque lo de la derecha no ocurrió.
		 */
		forecastFrom?: number | null;
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
		forecastFrom = null,
		onactivate
	}: Props = $props();

	const { padL, padR, padT, plotH, svgW, svgH } = PLOT;

	let svgEl: SVGSVGElement | undefined = $state();

	const toX = (i: number) => toPlotX(i, points.length);
	const toY = (v: number) => toPlotY(v, scale);

	/*
	 * Dónde acaba lo que de verdad pasó. Sin proyección es la serie entera, y con
	 * ella el último día con dato: es el punto que las dos líneas comparten, así
	 * que entra en los dos trazos y la curva no se rompe en el empalme.
	 */
	const historyEnd = $derived(forecastFrom === null ? points.length : Math.max(0, forecastFrom));
	const hasForecast = $derived(historyEnd > 0 && historyEnd < points.length);

	const coords = (from: number, to: number, pick: (p: GrowthPoint) => number) =>
		points
			.slice(from, to)
			.map((p, i) => `${toX(from + i)},${toY(pick(p))}`)
			.join(' ');

	const mvPoints = $derived(coords(0, historyEnd, (p) => p.mv));
	const cbPoints = $derived(coords(0, historyEnd, (p) => p.cb));
	/*
	 * La banda proyectada: `mv` lleva el techo —si se repitiera el mejor mes— y
	 * `cb` el suelo —si se repitiera el peor—. Los dos trazos arrancan en el
	 * último punto real, así que la banda se abre desde donde está la cartera
	 * hoy en vez de aparecer ya separada.
	 *
	 * El coste no se prolonga: proyectar aportes futuros sería inventarse lo que
	 * su dueño va a ingresar, y por eso `cb` significa otra cosa a partir de aquí.
	 */
	const forecastHigh = $derived(
		hasForecast ? coords(historyEnd - 1, points.length, (p) => p.mv) : ''
	);
	const forecastLow = $derived(
		hasForecast ? coords(historyEnd - 1, points.length, (p) => p.cb) : ''
	);
	/* El relleno de la banda: el techo de ida y el suelo de vuelta. */
	const forecastBand = $derived.by(() => {
		if (!hasForecast) return '';
		const back = points
			.slice(historyEnd - 1)
			.map((p, i) => ({ x: toX(historyEnd - 1 + i), y: toY(p.cb) }))
			.reverse()
			.map((c) => `${c.x},${c.y}`)
			.join(' ');

		return `${forecastHigh} ${back}`;
	});
	/*
	 * El relleno cierra contra la línea de referencia, no siempre contra el
	 * suelo: en porcentaje una racha negativa tiene que verse colgando por
	 * debajo del cero, y rellenarla hasta el borde la pintaba como si fuera
	 * terreno ganado.
	 */
	const floorY = $derived(
		baseline === null ? padT + plotH : Math.min(Math.max(toY(baseline), padT), padT + plotH)
	);
	/* El relleno solo cubre el historial: sombrear la proyección la pintaría
	   con el mismo peso que lo que de verdad ocurrió. */
	const mvFill = $derived(
		historyEnd < 2 ? '' : `${mvPoints} ${toX(historyEnd - 1)},${floorY} ${toX(0)},${floorY}`
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
	/** Si el punto `i` es proyección y no historial. */
	const isForecast = (i: number) => hasForecast && i >= historyEnd;

	/*
	 * Las filas de la tabla oculta: el historial entero y, de la proyección, solo
	 * los cierres de mes.
	 *
	 * La curva proyectada avanza al mismo paso que el historial para que el eje
	 * no mienta, así que un año son cientos de puntos. Dictarlos uno a uno no
	 * informa de nada —es una exponencial, y entre dos días consecutivos no pasa
	 * nada— y sepulta el historial, que sí es dato.
	 */
	const tableRows = $derived(
		points
			.map((point, i) => ({ point, i }))
			.filter(
				({ point, i }) =>
					!isForecast(i) ||
					i === points.length - 1 ||
					points[i + 1].date.slice(0, 7) !== point.date.slice(0, 7)
			)
	);

	const valueText = $derived(
		activePoint && active !== null
			? isForecast(active)
				? `${formatDate(activePoint.date)}, proyección: entre ${formatValue(activePoint.cb)} y ${formatValue(activePoint.mv)}`
				: `${formatDate(activePoint.date)}: ${primaryLabel} ${formatValue(activePoint.mv)}, ${secondaryLabel} ${formatValue(activePoint.cb)}`
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

		{#if hasForecast}
			<!-- La banda va debajo de sus bordes y de la raya de hoy: es el fondo
			     sobre el que se leen, no una figura más. -->
			<polygon points={forecastBand} class="band" />
			<!-- La raya de hoy: sin ella la banda parece parte del historial, y lo
			     que separa las dos mitades es justo que una pasó. -->
			<line
				x1={toX(historyEnd - 1)}
				y1={padT}
				x2={toX(historyEnd - 1)}
				y2={padT + plotH}
				class="today"
			/>
			<polyline points={forecastLow} class="line-forecast" />
			<polyline points={forecastHigh} class="line-forecast" />
		{/if}

		{#if historyEnd > 0}
			{@const lastIndex = historyEnd - 1}
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
			<!-- Sobre la banda los dos puntos son sus dos bordes, así que el de
			     abajo se marca en ámbar como el de arriba: ahí no hay coste que
			     señalar, porque la proyección no supone aportes. -->
			<circle
				cx={toX(active)}
				cy={toY(activePoint.cb)}
				r={isForecast(active) ? 4 : 3.5}
				class="cursor-dot"
				class:cost={!isForecast(active)}
				class:value={isForecast(active)}
			/>
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
				{#each tableRows as { point, i } (point.date)}
					<tr>
						<th scope="row">
							<!-- En la misma línea que la fecha: el salto entre las dos dejaba un
							     espacio colgando al final de cada fila del historial. -->
							<time datetime={point.date}>{formatFullDate(point.date)}</time>{#if isForecast(i)}
								(proyección: mejor y peor mes){/if}
						</th>
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

	/* La proyección, del mismo color pero discontinua y más fina: es la misma
	   cartera, y la línea rota dice que a partir de ahí no es un dato. */
	.line-forecast {
		fill: none;
		stroke: var(--amber);
		stroke-width: 2;
		stroke-dasharray: 5 5;
		stroke-linecap: round;
		stroke-linejoin: round;
		opacity: 0.7;
	}

	.today {
		stroke: var(--border-strong);
		stroke-width: 1;
		stroke-dasharray: 2 4;
	}

	/* El rango entre los dos extremos. Muy tenue a propósito: lo que hay dentro
	   no es un dato, es todo lo que cabe entre el mejor mes y el peor. */
	.band {
		fill: var(--amber);
		fill-opacity: 0.09;
		stroke: none;
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
