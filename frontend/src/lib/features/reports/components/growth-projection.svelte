<script lang="ts">
	/*
	 * Adónde llega la cuenta a cinco años si el ritmo del historial se mantiene.
	 *
	 * La curva dibuja el porcentaje acumulado y el eje incluye siempre el cero.
	 * Antes dibujaba los importes y estiraba su rango al alto del lienzo: con
	 * una tasa del −0,3 % anual, los noventa dólares que separaban el primer año
	 * del último ocupaban el canvas entero y la proyección se leía como un
	 * desplome. Un ritmo que no mueve nada tiene que dibujar una línea que no se
	 * mueve.
	 *
	 * El dinero está en la tabla de debajo, con una columna por año. Esa tabla
	 * es la gráfica en cifras, así que sirve igual a quien la lee con un lector
	 * de pantalla: la versión oculta que había antes desaparece con ella.
	 */
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPercent, formatSignedPercent } from '$lib/shared/format/percent';
	import { PROJECTION_MIN_DAYS, projectionGeometry } from '../projection';
	import type { GrowthProjectionSeries } from '../projection';

	interface Props {
		/** `null` mientras el historial no dé para proyectar. */
		projection: GrowthProjectionSeries | null;
		/** Días de historial, para que el estado vacío diga cuánto falta. */
		historyDays: number;
		/** Moneda en la que están los importes proyectados. */
		currency: string;
	}

	let { projection, historyDays, currency }: Props = $props();

	const missingDays = $derived(Math.max(PROJECTION_MIN_DAYS - historyDays, 0));
	const entries = $derived(projection?.entries ?? []);
	const geometry = $derived(projectionGeometry(entries));

	const money = (value: number) => privacy.money(formatCurrency(value, currency));
</script>

<section class="projection" aria-labelledby="projection">
	<h2 id="projection">Si el ritmo se mantiene</h2>

	{#if projection && geometry.points.length > 0}
		<p class="lead">
			Tu cuenta ha rendido un {formatSignedPercent(projection.annualRatePct)} anual. Extendido cinco años
			sobre lo que tienes hoy, sin contar aportes futuros, llegaría aquí.
		</p>

		<svg class="chart" viewBox="0 0 600 208" preserveAspectRatio="xMidYMid meet" aria-hidden="true">
			{#each geometry.ticks as tick (tick.y)}
				<line x1="54" y1={tick.y} x2="570" y2={tick.y} stroke="var(--border)" stroke-width="1" />
				<text x="46" y={tick.y + 3.5} text-anchor="end" class="axis">
					{formatSignedPercent(tick.value)}
				</text>
			{/each}

			<!-- El cero se dibuja aparte y más marcado: es la referencia contra la
			     que se lee si la curva sube o baja. -->
			<line
				x1="54"
				y1={geometry.zeroY}
				x2="570"
				y2={geometry.zeroY}
				stroke="var(--border-strong)"
				stroke-width="1"
			/>
			<text x="46" y={geometry.zeroY + 3.5} text-anchor="end" class="axis zero">
				{formatPercent(0)}
			</text>

			<!-- Sin relleno bajo la curva: lo que hay que leer aquí es la distancia
			     al cero, y un bloque de color entre los dos la tapaba. -->
			<polyline
				points={geometry.line}
				fill="none"
				stroke="var(--amber)"
				stroke-width="2.5"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>

			{#each geometry.points as point (point.period)}
				<circle cx={point.x} cy={point.y} r="3.5" fill="var(--amber-light)" />
				<text x={point.x} y="192" text-anchor="middle" class="axis">{point.period}</text>
			{/each}
		</svg>

		<!-- Como la matriz: el `tabindex` es lo que deja desplazar la tabla con el
		     teclado cuando no cabe entera. -->
		<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
		<!-- Con nombre propio, como la matriz: la sección ya se llama así. -->
		<div
			class="scroller"
			role="region"
			aria-label="Proyección año por año, tabla desplazable"
			tabindex="0"
		>
			<table>
				<caption class="sr-only">
					Valor proyectado de la cuenta y rentabilidad acumulada desde hoy, año por año
				</caption>
				<thead>
					<tr>
						<td class="corner"></td>
						{#each entries as entry, index (entry.period)}
							<th scope="col">{entry.period}{index === 0 ? ', hoy' : ''}</th>
						{/each}
					</tr>
				</thead>
				<tbody>
					<tr>
						<th scope="row">Valor proyectado</th>
						{#each entries as entry (entry.period)}
							<td>{money(entry.value)}</td>
						{/each}
					</tr>
					<tr>
						<th scope="row">Acumulado desde hoy</th>
						{#each entries as entry (entry.period)}
							<td class="pct" class:up={entry.returnPct > 0} class:down={entry.returnPct < 0}>
								{formatSignedPercent(entry.returnPct)}
							</td>
						{/each}
					</tr>
				</tbody>
			</table>
		</div>

		<p class="footnote">
			No es una previsión de mercado: es tu propio ritmo repetido cinco veces. El porcentaje es lo
			que de verdad se proyecta; el importe se mueve con cada aporte que hagas, sin que la
			proyección haya cambiado de opinión.
		</p>
	{:else}
		<p class="empty">
			La proyección necesita al menos seis meses de historial.
			<!-- Decir cuánto falta ahorra volver cada semana a comprobarlo. -->
			{#if historyDays > 0}
				Llevas {historyDays}
				{historyDays === 1 ? 'día' : 'días'}, así que faltan {missingDays}.
			{:else}
				Empieza a contar con el primer cierre diario de tu cartera.
			{/if}
		</p>
	{/if}
</section>

<style>
	.projection {
		padding: 2rem 0;
		border-bottom: 1px solid var(--border);
	}

	h2 {
		margin: 0 0 0.6rem;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	.lead,
	.empty {
		max-width: 64ch;
		margin: 0;
		font-size: 0.9rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.chart {
		width: 100%;
		max-width: 52rem;
		min-height: 220px;
		margin-top: 0.75rem;
		display: block;
	}

	.axis {
		fill: var(--text-dim);
		font-size: 11px;
		font-family: var(--font-mono);
	}

	/* La marca del cero acompaña a su filete, que es más marcado que el resto. */
	.zero {
		fill: var(--text-muted);
	}

	/* Al ancho de la gráfica: la tabla es la misma proyección en cifras, y a lo
	   ancho de la página se leían como dos bloques distintos. */
	/* `position: relative` por lo mismo que en la matriz: aquí el que se escapa
	   es el `caption` accesible de la tabla. */
	.scroller {
		position: relative;
		max-width: 52rem;
		overflow-x: auto;
		margin-top: 0.5rem;
	}

	/*
	 * Sombra en el borde derecho mientras quede tabla por ver, y solo mientras
	 * quede: la capa opaca viaja con el contenido (`local`) y tapa a la sombra
	 * —fija al carril (`scroll`)— justo al llegar al final. Sin ella, en un
	 * móvil nada dice que a la derecha siguen media docena de columnas.
	 */
	.scroller {
		background:
			linear-gradient(var(--bg), var(--bg)) right center / 1.25rem 100% no-repeat local,
			linear-gradient(to left, rgba(8, 9, 10, 0.92), rgba(8, 9, 10, 0)) right center / 2.5rem 100%
				no-repeat scroll;
	}

	.scroller:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 4px;
	}

	table {
		width: 100%;
		min-width: 40rem;
		border-collapse: collapse;
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
	}

	thead th {
		padding: 0 0.75rem 0.6rem;
		border-bottom: 1px solid var(--border);
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
		text-align: right;
	}

	/*
	 * La columna de las etiquetas se queda quieta mientras los años se
	 * desplazan: sin ella, a mitad de recorrido no se sabe si la cifra que se
	 * está leyendo es el valor proyectado o lo acumulado.
	 */
	.corner,
	tbody th {
		position: sticky;
		left: 0;
		width: 12rem;
		background: var(--bg);
	}

	tbody th,
	tbody td {
		padding: 0.7rem 0.75rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.82rem;
		font-weight: 400;
		text-align: right;
	}

	tbody th {
		padding-left: 0;
		font-family: var(--font-body);
		color: var(--text-muted);
		text-align: left;
	}

	tbody td {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--text);
		white-space: nowrap;
	}

	tbody tr:last-child th,
	tbody tr:last-child td {
		border-bottom: none;
	}

	.pct.up {
		color: var(--green);
	}

	.pct.down {
		color: var(--red);
	}

	.footnote {
		max-width: 78ch;
		margin: 0.9rem 0 0;
		font-size: 0.75rem;
		line-height: 1.55;
		color: var(--text-dim);
	}

	/*
	 * En un móvil la curva se encoge a un tercio y sus etiquetas quedan en seis
	 * píxeles: ilegibles, y encima empujando la tabla —que dice lo mismo con
	 * cifras que sí se leen— fuera de la pantalla. Ahí se queda la tabla sola.
	 */
	@media (max-width: 640px) {
		.chart {
			display: none;
		}

		/* Y la columna de las etiquetas se queda con la mitad del ancho de la
		   pantalla si no se le pone coto. */
		table {
			min-width: 32rem;
		}

		.corner {
			width: 8.5rem;
		}
	}
</style>
