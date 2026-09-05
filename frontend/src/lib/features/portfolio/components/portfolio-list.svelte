<script lang="ts">
	/*
	 * Los portafolios, uno por fila, de mayor a menor.
	 *
	 * Sustituye a una rejilla de tarjetas. Comparar es lo que se viene a hacer
	 * aquí —cuál pesa más, a cuál le va mejor— y en una rejilla las cifras caen
	 * a una altura distinta en cada tarjeta y las columnas se recolocan solas al
	 * cambiar el ancho, así que no había manera de recorrerlas con la vista.
	 *
	 * La barra es la gráfica de crecimiento del portafolio contraída a un solo
	 * instante: el mismo par que dibujan sus dos series —el capital que pusiste,
	 * frío, y lo que el mercado hizo con él, cálido—. El corte cae siempre en el
	 * capital: lo que queda dentro es la ganancia y lo que asoma por fuera del
	 * extremo es lo que falta para recuperarlo. El largo de la barra es siempre
	 * el valor de mercado, que es lo que mantiene las filas comparables.
	 */
	import EmptyState from '$lib/ui/empty-state.svelte';
	import PortfolioCapitalBar from './portfolio-capital-bar.svelte';
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPct } from '../portfolio';
	import type { PortfolioRow, PortfolioTotals } from '../portfolio';

	let {
		rows,
		totals,
		scale,
		displayCurrency
	}: {
		/** Los de esta hoja, ya ordenados. */
		rows: PortfolioRow[];
		/** Totales de la lista entera, no de la hoja. */
		totals: PortfolioTotals;
		/** Ancho de referencia de las barras: el mayor importe de la lista entera. */
		scale: number;
		displayCurrency: string;
	} = $props();

	const money = (value: number, currency = displayCurrency) =>
		privacy.money(formatCurrency(value, currency));
</script>

{#if rows.length > 0}
	<table>
		<caption class="sr-only">
			Tus portafolios, de mayor a menor valor, con su nivel de riesgo, el capital que invertiste en
			cada uno y lo que llevan ganado
		</caption>
		<thead>
			<tr>
				<th scope="col">Portafolio</th>
				<th scope="col" class="col-risk">Riesgo</th>
				<th scope="col" class="col-bar">
					<!-- La cabecera de la columna es la leyenda de la barra: es el único
					     sitio donde hace falta decir qué significa cada color, y así no
					     hay que repetirlo debajo de cada fila. -->
					<span class="legend" aria-hidden="true">
						<span class="key key-cost"></span>capital
						<span class="key key-gain"></span>ganancia
					</span>
					<span class="sr-only">Capital invertido y ganancia</span>
				</th>
				<th scope="col" class="col-value num">Valor en {displayCurrency}</th>
				<th scope="col" class="col-return num">Rendimiento</th>
			</tr>
		</thead>

		<tbody>
			{#each rows as row (row.id)}
				{@const up = row.gain >= 0}
				<tr>
					<th scope="row" class="who">
						<a class="name" href={resolve('/dashboard/portfolios/[id]', { id: row.id })}>
							{row.name}
						</a>
						{#if row.isDefault}
							<span class="tag">predeterminado</span>
						{/if}
						<span class="detail">
							{row.description || row.typeLabel},
							{row.positions === 1 ? '1 posición' : `${row.positions} posiciones`}
						</span>
					</th>

					<td class="col-risk risk">{row.riskName}</td>

					<td class="col-bar">
						<PortfolioCapitalBar {row} {scale} {displayCurrency} />
					</td>

					<td class="col-value num value">
						{money(row.value, row.currency)}
						{#if !row.converted}
							<span class="qualifier">en {row.currency}, su propia moneda</span>
						{:else if row.unconverted > 0}
							<span class="qualifier flagged">
								{row.unconverted === 1
									? 'una posición suma'
									: `${row.unconverted} posiciones suman`}
								sin convertir
							</span>
						{/if}
					</td>

					<td class="col-return num return" class:up class:down={!up}>
						{row.positions > 0 ? formatPct(row.gainPct) : '—'}
					</td>
				</tr>
			{/each}
		</tbody>

		<!-- El total va al pie de su columna, que es donde vive un total. Estaba
		     en una tarjeta encima de la lista, repitiendo la cifra que el panel
		     ya enseña en grande. -->
		<tfoot>
			<tr>
				<th scope="row" class="who">
					Total
					<span class="detail">
						{totals.counted === 1 ? '1 portafolio' : `${totals.counted} portafolios`},
						{totals.positions === 1
							? '1 posición abierta'
							: `${totals.positions} posiciones abiertas`}
					</span>
				</th>
				<td class="col-risk"></td>
				<td class="col-bar">
					<span class="cost-total">Capital invertido: {money(totals.cost)}</span>
				</td>
				<td class="col-value num value">{money(totals.value)}</td>
				<td class="col-return num return" class:up={totals.gain >= 0} class:down={totals.gain < 0}>
					{totals.positions > 0 ? formatPct(totals.gainPct) : '—'}
				</td>
			</tr>
		</tfoot>
	</table>
{:else}
	<EmptyState
		bordered
		title="Todavía no tienes portafolios"
		description="Un portafolio agrupa lo que compraste con una misma intención: el largo plazo, lo especulativo, el colchón. Crea el primero y empieza a registrar posiciones dentro."
	>
		{#snippet action()}
			<a class="create" href={resolve('/dashboard/portfolios/add')}>Crear el primero</a>
		{/snippet}
	</EmptyState>
{/if}

<style>
	table {
		width: 100%;
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
		padding: 0 0.75rem 0.7rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
		text-align: left;
		vertical-align: bottom;
		white-space: nowrap;
	}

	thead th.num {
		text-align: right;
	}

	thead th:first-child {
		padding-left: 0;
	}

	thead th:last-child {
		padding-right: 0;
	}

	.legend {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
	}

	.key {
		width: 9px;
		height: 9px;
		border-radius: 2px;
	}

	.key-cost {
		background: var(--cost);
	}

	.key-gain {
		margin-left: 0.4rem;
		background: var(--green);
	}

	tbody th,
	tbody td,
	tfoot th,
	tfoot td {
		padding: 1rem 0.75rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.85rem;
		font-weight: 400;
		color: var(--text);
		text-align: left;
		vertical-align: top;
	}

	tbody th:first-child,
	tfoot th:first-child {
		padding-left: 0;
	}

	tbody td:last-child,
	tfoot td:last-child {
		padding-right: 0;
	}

	@media (hover: hover) {
		tbody tr:hover th,
		tbody tr:hover td {
			background: var(--panel);
		}
	}

	/* El pie no es una fila más de la lista: va separado por su propio filete y
	   sin el de abajo, que duplicaría el borde de la sección. */
	tfoot th,
	tfoot td {
		border-bottom: none;
		border-top: 1px solid var(--border-strong);
		color: var(--text-muted);
	}

	tbody tr:last-child th,
	tbody tr:last-child td {
		border-bottom: none;
	}

	.who {
		min-width: 0;
	}

	.name {
		font-size: 0.95rem;
		color: var(--text);
		text-decoration: none;
		overflow-wrap: anywhere;
	}

	.name:hover {
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	/* Cuál usa la aplicación cuando no eliges: es una propiedad del portafolio,
	   no un elogio, así que va en el gris de las anotaciones y no en el ámbar. */
	.tag {
		margin-left: 0.5rem;
		font-size: 0.72rem;
		/* En el gris de las anotaciones y no en el más apagado: dice cuál usa la
		   aplicación cuando no eliges, que es un dato y hay que poder leerlo. */
		color: var(--text-muted);
	}

	.detail {
		display: block;
		max-width: 46ch;
		margin-top: 0.25rem;
		font-size: 0.78rem;
		line-height: 1.4;
		color: var(--text-muted);
	}

	/*
	 * El riesgo era una etiqueta en versalitas y en verde, ámbar o rojo. En una
	 * fila que además dice «+21,22%» en verde, el rojo de «Agresivo» se leía
	 * como una pérdida: es una característica del portafolio, no un resultado.
	 */
	.risk {
		color: var(--text-muted);
		white-space: nowrap;
	}

	.cost-total {
		display: block;
		margin-top: 0.15rem;
		font-size: 0.78rem;
		color: var(--text-dim);
	}

	.num {
		text-align: right;
	}

	.value,
	.return {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.value {
		font-weight: 600;
	}

	.return.up {
		color: var(--green);
	}

	.return.down {
		color: var(--red);
	}

	.qualifier {
		display: block;
		margin-top: 0.2rem;
		font-family: var(--font-body);
		font-size: 0.72rem;
		font-weight: 400;
		color: var(--text-dim);
		white-space: normal;
	}

	.qualifier.flagged {
		color: var(--amber);
	}

	.col-risk {
		width: 7rem;
	}

	.col-bar {
		width: 28%;
	}

	.col-value {
		width: 11rem;
	}

	.col-return {
		width: 7rem;
	}

	thead th:first-child {
		width: 31%;
	}

	.create {
		display: inline-flex;
		align-items: center;
		padding: 0.75rem 1.4rem;
		border-radius: 10px;
		background: var(--amber);
		color: #0d0800;
		font-size: 0.9rem;
		font-weight: 600;
		text-decoration: none;
		transition: background 0.2s ease;
	}

	.create:hover {
		background: var(--amber-light);
	}

	/*
	 * Debajo de esto la fila no cabe en una línea. Se pliega en dos: el nombre y
	 * el valor arriba, el riesgo y el rendimiento debajo, y la barra cruzando
	 * las dos columnas. Antes se desplazaba de lado y el rendimiento —la mitad
	 * de la pregunta que trae aquí a alguien— quedaba fuera de la pantalla.
	 */
	@media (max-width: 820px) {
		thead {
			display: none;
		}

		tbody tr,
		tfoot tr {
			display: grid;
			grid-template-columns: minmax(0, 1fr) auto;
			column-gap: 1rem;
			padding: 1rem 0;
			border-bottom: 1px solid var(--border);
		}

		tbody tr:last-child {
			border-bottom: none;
		}

		tfoot tr {
			border-bottom: none;
			border-top: 1px solid var(--border-strong);
		}

		tbody th,
		tbody td,
		tfoot th,
		tfoot td {
			width: auto;
			padding: 0;
			border: none;
		}

		.who {
			grid-column: 1;
			grid-row: 1;
		}

		.col-value {
			grid-column: 2;
			grid-row: 1;
			width: auto;
		}

		.col-risk {
			grid-column: 1;
			grid-row: 2;
			width: auto;
			margin-top: 0.6rem;
			font-size: 0.78rem;
		}

		.col-return {
			grid-column: 2;
			grid-row: 2;
			width: auto;
			margin-top: 0.6rem;
		}

		.col-bar {
			grid-column: 1 / -1;
			grid-row: 3;
			width: auto;
			margin-top: 0.7rem;
		}

		@media (hover: hover) {
			tbody tr:hover th,
			tbody tr:hover td {
				background: none;
			}
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.create {
			transition: none;
		}
	}
</style>
