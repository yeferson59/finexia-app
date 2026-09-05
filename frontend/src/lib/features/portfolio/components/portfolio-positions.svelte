<script lang="ts">
	/*
	 * Qué hay dentro del portafolio: una fila por activo, de mayor a menor peso.
	 *
	 * Absorbe cuatro tarjetas y un donut. «Mejor activo», «peor activo» y
	 * «concentración» eran tres tarjetas que decían cuál es la primera y cuál la
	 * última de esta misma lista; ahora la lista está ordenada y una frase
	 * nombra los extremos, que es lo único que costaba encontrar. El donut
	 * repartía el portafolio entre dos o tres clases de activo: eso cabe en una
	 * línea, y la columna «Clase» deja comprobarla fila a fila.
	 *
	 * Las filas eran botones con `aria-label`, que tapaba su contenido: quien
	 * usa lector de pantalla oía «ver detalles de AAPL» y ni el valor ni el
	 * rendimiento. Ahora el símbolo es un enlace y las cabeceras nombran cada
	 * columna.
	 */
	import EmptyState from '$lib/ui/empty-state.svelte';
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPercent } from '$lib/shared/format/percent';
	import { formatAssetType } from '$lib/shared/format/asset-type';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import { formatPct, type HoldingView, type TopTransaction } from '../portfolio';
	import type { TypeBreakdownSlice } from '../portfolio';

	let {
		holdings,
		typeBreakdown,
		topTransaction,
		portfolioId,
		baseCurrency
	}: {
		holdings: HoldingView[];
		/** Reparto por clase de activo, ya ordenado de mayor a menor. */
		typeBreakdown: TypeBreakdownSlice[];
		/** La operación de mayor importe registrada; `null` si no hay ninguna. */
		topTransaction: TopTransaction | null;
		portfolioId: string;
		baseCurrency: string;
	} = $props();

	const money = (amount: number) => privacy.money(formatCurrency(amount, baseCurrency));

	// De mayor a menor peso: así la primera fila es la posición dominante y no
	// hace falta una tarjeta que lo diga.
	const rows = $derived([...holdings].sort((a, b) => b.value - a.value));

	/** El de mejor y el de peor rendimiento, cuando hay más de uno que comparar. */
	const standouts = $derived.by(() => {
		if (rows.length < 2) return null;

		const byReturn = [...rows].sort((a, b) => b.gainLossPct - a.gainLossPct);
		const best = byReturn[0];
		const worst = byReturn[byReturn.length - 1];

		return best.symbol === worst.symbol ? null : { best, worst };
	});
</script>

<section class="positions" aria-labelledby="positions-title">
	<header class="head">
		<h2 id="positions-title">Posiciones</h2>
		<p class="count">{rows.length === 1 ? '1 activo' : `${rows.length} activos`}</p>
	</header>

	{#if rows.length > 0}
		{#if typeBreakdown.length > 1}
			<p class="mix">
				{#each typeBreakdown as slice, i (slice.type)}{i > 0 ? ', ' : ''}{slice.label}
					{formatPercent(slice.pct)}{/each}.
			</p>
		{/if}

		{#if standouts}
			<p class="standouts">
				{standouts.best.symbol} es la que más ha rendido ({formatPct(standouts.best.gainLossPct)}); {standouts
					.worst.symbol}, la que menos ({formatPct(standouts.worst.gainLossPct)}).
			</p>
		{/if}

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

		{#if topTransaction}
			<p class="top-txn">
				La mayor operación registrada aquí: {topTransaction.assetTicker}, {money(
					parseFloat(topTransaction.value) || 0
				)}, el {formatCalendarDate(topTransaction.transactionDate, {
					year: 'numeric',
					month: 'long',
					day: 'numeric'
				})}.
			</p>
		{/if}
	{:else}
		<EmptyState
			bordered
			title="Este portafolio aún no tiene activos"
			description="Registra lo que tienes en cada plataforma y aquí verás cuánto pesa y cómo va rindiendo."
		>
			{#snippet action()}
				<a class="add" href={resolve('/dashboard/portfolios/[id]/add', { id: portfolioId })}>
					Agregar tu primer activo
				</a>
			{/snippet}
		</EmptyState>
	{/if}
</section>

<style>
	.positions {
		padding-top: 2rem;
	}

	.head {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.5rem 1rem;
	}

	h2 {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	.count {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-dim);
	}

	/* Lo que el donut dibujaba: con dos o tres clases, una frase lo dice y la
	   columna «Clase» deja comprobarlo fila a fila. */
	.mix,
	.standouts {
		max-width: 68ch;
		margin: 0.6rem 0 0;
		font-size: 0.85rem;
		line-height: 1.45;
		color: var(--text-muted);
	}

	.standouts {
		margin-top: 0.3rem;
	}

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

	.symbol {
		display: block;
		font-family: var(--font-mono);
		font-weight: 600;
		color: var(--text);
		text-decoration: none;
	}

	.symbol:hover {
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.name {
		display: block;
		margin-top: 0.15rem;
		font-size: 0.78rem;
		line-height: 1.35;
		color: var(--text-muted);
		overflow-wrap: anywhere;
	}

	.type {
		color: var(--text-muted);
		white-space: nowrap;
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

	.top-txn {
		max-width: 68ch;
		margin: 1.25rem 0 0;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-dim);
	}

	.add {
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

	.add:hover {
		background: var(--amber-light);
	}

	@media (prefers-reduced-motion: reduce) {
		.add {
			transition: none;
		}
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
