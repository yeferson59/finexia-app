<script lang="ts">
	/*
	 * La tabla de las posiciones abiertas, ya ordenadas por peso.
	 *
	 * Vive aparte de `portfolio-positions`, que pone la cabecera, las frases de
	 * resumen, el estado vacío y las cerradas: juntas pasaban del presupuesto de
	 * tamaño. El aspecto del símbolo, el nombre y la clase no está aquí: lo
	 * comparte con la lista de cerradas y lo aporta el contenedor.
	 *
	 * Las filas eran botones con `aria-label`, que tapaba su contenido: quien
	 * usa lector de pantalla oía «ver detalles de AAPL» y ni el valor ni el
	 * rendimiento. Ahora el símbolo es un enlace y las cabeceras nombran cada
	 * columna.
	 */
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPercent } from '$lib/shared/format/percent';
	import { formatAssetType } from '$lib/shared/format/asset-type';
	import { formatPct, type HoldingView } from '../portfolio';

	let {
		rows,
		portfolioId,
		baseCurrency
	}: {
		/** Las posiciones abiertas, de mayor a menor peso. */
		rows: HoldingView[];
		portfolioId: string;
		baseCurrency: string;
	} = $props();

	const money = (amount: number) => privacy.money(formatCurrency(amount, baseCurrency));
</script>

<table>
	<caption class="sr-only">
		Los activos de este portafolio, de mayor a menor peso, con su clase, cuánto pesan sobre el
		total, lo que valen y cuánto han rendido
	</caption>
	<thead>
		<tr>
			<th scope="col">Activo</th>
			<th scope="col" class="col-class">Clase</th>
			<th scope="col" class="col-weight">Peso</th>
			<th scope="col" class="col-value num">Valor en {baseCurrency}</th>
			<th scope="col" class="col-return num">Rendimiento</th>
		</tr>
	</thead>
	<tbody>
		{#each rows as holding (holding.symbol)}
			{@const up = holding.gainLoss >= 0}
			<tr>
				<th scope="row" class="who">
					<a
						class="symbol"
						href={resolve('/dashboard/portfolios/[id]/assets/[symbol]', {
							id: portfolioId,
							symbol: holding.symbol
						})}
					>
						{holding.symbol}
					</a>
					<span class="name">{holding.name}</span>
				</th>

				<td class="col-class type">{formatAssetType(holding.assetType)}</td>

				<td class="col-weight">
					<span class="weight">
						<span class="track" aria-hidden="true">
							<span class="fill" style="width: {Math.min(holding.allocation, 100).toFixed(2)}%"
							></span>
						</span>
						<span class="pct">{formatPercent(holding.allocation)}</span>
					</span>
				</td>

				<td class="col-value num value">
					{money(holding.value)}
					{#if !holding.fxConverted}
						<span class="qualifier">sin convertir a {baseCurrency}</span>
					{/if}
				</td>

				<td class="col-return num return" class:up class:down={!up}>
					{formatPct(holding.gainLossPct)}
				</td>
			</tr>
		{/each}
	</tbody>
</table>

<style>
	table {
		width: 100%;
		margin-top: 1.35rem;
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
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
		text-align: left;
		white-space: nowrap;
	}

	thead th.num {
		text-align: right;
	}

	thead th:first-child {
		padding-left: 0;
		width: 28%;
	}

	thead th:last-child {
		padding-right: 0;
	}

	tbody th,
	tbody td {
		padding: 0.85rem 0.75rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.85rem;
		font-weight: 400;
		color: var(--text);
		text-align: left;
		vertical-align: middle;
	}

	tbody th:first-child {
		padding-left: 0;
	}

	tbody td:last-child {
		padding-right: 0;
	}

	tbody tr:last-child th,
	tbody tr:last-child td {
		border-bottom: none;
	}

	@media (hover: hover) {
		tbody tr:hover th,
		tbody tr:hover td {
			background: var(--panel);
		}
	}

	/*
	 * Una sola serie, un solo color: el largo dice el peso y el nombre de la
	 * fila, de quién es. La barra llevaba un degradado de ámbar a verde que no
	 * codificaba nada —el verde es la ganancia en toda la aplicación, y aquí
	 * aparecía también en las posiciones que perdían—.
	 */
	/*
	 * El carril va de 0 a 100 y no al mayor peso de la lista: aquí la pregunta
	 * no es solo cuál pesa más sino si alguna posición se ha comido el
	 * portafolio, y eso solo se lee contra el total. Es la tarjeta
	 * «Concentración» convertida en la propia columna.
	 */
	.weight {
		display: flex;
		align-items: center;
		gap: 0.6rem;
	}

	.track {
		flex: 1;
		min-width: 2.5rem;
		height: 6px;
		border-radius: 3px;
		background: var(--panel-2);
		overflow: hidden;
	}

	.fill {
		display: block;
		height: 100%;
		min-width: 2px;
		border-radius: 0 3px 3px 0;
		background: var(--amber);
	}

	.pct {
		flex-shrink: 0;
		min-width: 3.2rem;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-variant-numeric: tabular-nums;
		text-align: right;
		color: var(--text-muted);
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
		color: var(--amber);
		white-space: normal;
	}

	.col-class {
		width: 8rem;
	}

	.col-weight {
		width: 24%;
	}

	.col-value {
		width: 11rem;
	}

	.col-return {
		width: 8rem;
	}

	/* Debajo de esto la fila se pliega en dos, como en el listado: el activo y
	   su valor arriba, la clase y el rendimiento debajo, y la barra cruzando. */
	@media (max-width: 820px) {
		thead {
			display: none;
		}

		tbody tr {
			display: grid;
			grid-template-columns: minmax(0, 1fr) auto;
			column-gap: 1rem;
			padding: 0.9rem 0;
			border-bottom: 1px solid var(--border);
		}

		tbody tr:last-child {
			border-bottom: none;
		}

		tbody th,
		tbody td {
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

		.col-class {
			grid-column: 1;
			grid-row: 2;
			width: auto;
			margin-top: 0.5rem;
			font-size: 0.78rem;
		}

		.col-return {
			grid-column: 2;
			grid-row: 2;
			width: auto;
			margin-top: 0.5rem;
		}

		.col-weight {
			grid-column: 1 / -1;
			grid-row: 3;
			width: auto;
			margin-top: 0.6rem;
		}

		.col-weight .track {
			min-width: 0;
		}
	}
</style>
