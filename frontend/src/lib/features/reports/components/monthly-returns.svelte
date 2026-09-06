<script lang="ts">
	/*
	 * La rentabilidad de cada mes, un año por fila.
	 *
	 * Sustituye a una rejilla de tarjetas, una por año, de doce mosaicos cada
	 * una. Comparar es lo que se viene a hacer aquí —qué meses tiraron, si el
	 * año pasado fue mejor que este— y en tarjetas cada mes caía a una altura
	 * distinta, con su leyenda de cinco colores y su pie legal repetidos debajo
	 * de cada año. Puestos en una matriz, un mes se compara con el de al lado y
	 * con el mismo mes del año anterior, que es justo lo que se busca.
	 *
	 * El color ya no pide aprenderse una escala: la celda imprime su cifra y el
	 * fondo solo dice signo e intensidad. Y el signo va en la propia cifra, así
	 * que quien no distinga los dos tonos no se pierde nada.
	 *
	 * Las cifras son rendimiento, no variación del saldo: `reports.ts` descuenta
	 * los aportes de cada tramo. El pie lo dice, porque una cuenta que crece a
	 * base de depósitos venía marcando meses del +150 % y nadie sabía por qué.
	 */
	import { formatSignedPercent } from '$lib/shared/format/percent';
	import { MONTHS, returnBackground, type PerformanceCalendar } from '../reports';

	let { calendars }: { calendars: PerformanceCalendar[] } = $props();

	const anyPartial = $derived(calendars.some((calendar) => calendar.partialMonths.length > 0));
</script>

<section class="matrix" aria-labelledby="monthly-returns">
	<h2 id="monthly-returns">Rentabilidad mes a mes</h2>

	<!-- Trece columnas no caben en un móvil: el carril se desplaza de lado. El
	     `tabindex` es lo que deja recorrerlo con el teclado —sin él, el contenido
	     que se sale por la derecha no hay forma de alcanzarlo sin ratón—, y por
	     eso la regla que lo prohíbe en un elemento sin interacción no aplica. -->
	<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
	<div class="scroller" role="region" aria-labelledby="monthly-returns" tabindex="0">
		<table>
			<caption class="sr-only">
				Rentabilidad de cada mes y del año entero, del año más reciente al más antiguo. Es
				rendimiento de lo invertido: los aportes y retiros no cuentan.
			</caption>
			<thead>
				<tr>
					<th scope="col" class="year">Año</th>
					{#each MONTHS as month (month)}
						<th scope="col">{month}</th>
					{/each}
					<th scope="col" class="total">Total</th>
				</tr>
			</thead>
			<tbody>
				{#each calendars as calendar (calendar.year)}
					<tr>
						<th scope="row" class="year">{calendar.year}</th>

						{#each calendar.values as value, index (`${calendar.year}-${MONTHS[index]}`)}
							<td class="cell">
								{#if value === null}
									<span class="chip empty" aria-hidden="true">–</span><span class="sr-only"
										>sin dato</span
									>
								{:else}
									<span class="chip" style:background-color={returnBackground(value)}>
										{formatSignedPercent(value)}{#if calendar.partialMonths.includes(index)}<span
												class="partial"
												aria-label="mes incompleto">*</span
											>{/if}
									</span>
								{/if}
							</td>
						{/each}

						<td
							class="total"
							class:up={(calendar.total ?? 0) >= 0}
							class:down={(calendar.total ?? 0) < 0}
						>
							{calendar.total === null ? '–' : formatSignedPercent(calendar.total)}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<p class="footnote">
		Rendimiento de lo invertido: el dinero que aportas o retiras dentro del mes no cuenta como
		rentabilidad.
	</p>

	{#if anyPartial}
		<p class="footnote">
			Un asterisco marca el mes que tu historial no cubre entero —aquel en el que empieza y el que
			está en curso—; esos no compiten por el mejor ni el peor mes.
		</p>
	{/if}
</section>

<style>
	.matrix {
		padding: 2rem 0;
		border-bottom: 1px solid var(--border);
	}

	h2 {
		margin: 0 0 1.1rem;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	.scroller {
		overflow-x: auto;
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
		/* Por debajo de esto las celdas se comen su propia cifra; el carril se
		   desplaza de lado en vez de apretarlas. */
		min-width: 46rem;
		/* Y por encima, el total del año se despegaba de diciembre y quedaba
		   solo contra el borde de la página. */
		max-width: 62rem;
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
		padding: 0 0.35rem 0.6rem;
		border-bottom: 1px solid var(--border);
		font-family: var(--font-body);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
		text-align: center;
	}

	thead th.year {
		text-align: left;
	}

	tbody th,
	tbody td {
		padding: 0.55rem 0.35rem;
		border-bottom: 1px solid var(--border);
		font-family: var(--font-mono);
		font-size: 0.78rem;
		font-variant-numeric: tabular-nums;
		text-align: center;
		white-space: nowrap;
	}

	tbody tr:last-child th,
	tbody tr:last-child td {
		border-bottom: none;
	}

	/* La columna del año se queda quieta mientras los meses se desplazan por
	   debajo: sin ella, a mitad de recorrido no se sabe qué año se está leyendo. */
	.year {
		position: sticky;
		left: 0;
		width: 3.5rem;
		padding-left: 0;
		background: var(--bg);
		text-align: left;
	}

	tbody th.year {
		font-weight: 500;
		color: var(--text);
	}

	.cell {
		width: 6%;
		padding: 0.35rem 0.2rem;
		color: var(--text);
	}

	/*
	 * El tinte va en una ficha y no en la celda entera: pegados unos a otros,
	 * tres meses verdes seguidos se fundían en una sola banda y había que contar
	 * las cifras para saber dónde acababa cada mes.
	 */
	.chip {
		display: block;
		padding: 0.3rem 0.15rem;
		border-radius: 4px;
	}

	.chip.empty {
		color: var(--text-dim);
	}

	.partial {
		margin-left: 0.05rem;
		color: var(--text-muted);
	}

	/* El total del año cierra la fila y se lee como su conclusión: con su propio
	   filete a la izquierda y en el color del signo. */
	.total {
		width: 5rem;
		padding-right: 0;
		border-left: 1px solid var(--border);
		font-weight: 600;
		text-align: right;
	}

	thead th.total {
		border-left: none;
	}

	td.total.up {
		color: var(--green);
	}

	td.total.down {
		color: var(--red);
	}

	.footnote {
		max-width: 78ch;
		margin: 0.9rem 0 0;
		/* Dos pies seguidos son dos notas, no un párrafo partido. */
		font-size: 0.75rem;
		line-height: 1.55;
		color: var(--text-dim);
	}

	.footnote + .footnote {
		margin-top: 0.35rem;
	}
</style>
