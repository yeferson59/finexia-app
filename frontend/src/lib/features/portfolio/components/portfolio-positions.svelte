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
	 * La tabla vive en `portfolio-positions-table`; aquí quedan la cabecera, las
	 * frases de resumen, el estado vacío y las posiciones cerradas.
	 */
	import EmptyState from '$lib/ui/empty-state.svelte';
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPercent } from '$lib/shared/format/percent';
	import { formatAssetType } from '$lib/shared/format/asset-type';
	import { formatCalendarDate } from '$lib/shared/format/date';
	import PortfolioPositionsTable from './portfolio-positions-table.svelte';
	import { formatPct, type HoldingView, type TopTransaction } from '../portfolio';
	import type { TypeBreakdownSlice } from '../portfolio';

	let {
		holdings,
		closed = [],
		typeBreakdown,
		topTransaction,
		portfolioId,
		baseCurrency
	}: {
		/** Las posiciones abiertas: las que pesan y rinden. */
		holdings: HoldingView[];
		/**
		 * Las vendidas enteras. Se enseñan aparte y plegadas: siguen siendo del
		 * portafolio y su historial vive en la página del activo, pero una fila a
		 * cero en la tabla contaba como activo sin tener nada que pesar.
		 */
		closed?: HoldingView[];
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

		<PortfolioPositionsTable {rows} {portfolioId} {baseCurrency} />

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
		<!-- Un portafolio con todo vendido no es uno vacío: invitar a «agregar tu
		     primer activo» encima de su historial sería decir que nunca tuvo nada. -->
		<EmptyState
			bordered
			title={closed.length > 0
				? 'No hay posiciones abiertas'
				: 'Este portafolio aún no tiene activos'}
			description={closed.length > 0
				? 'Todo lo que había aquí se vendió. Su historial sigue abajo, en cada activo.'
				: 'Registra lo que tienes en cada plataforma y aquí verás cuánto pesa y cómo va rindiendo.'}
		>
			{#snippet action()}
				<a class="add" href={resolve('/dashboard/portfolios/[id]/add', { id: portfolioId })}>
					{closed.length > 0 ? 'Agregar un activo' : 'Agregar tu primer activo'}
				</a>
			{/snippet}
		</EmptyState>
	{/if}

	{#if closed.length > 0}
		<details class="closed">
			<summary>
				Posiciones cerradas <span class="closed-count">{closed.length}</span>
			</summary>
			<p class="closed-hint">
				Vendidas por completo: ya no suman al valor ni al rendimiento. Sus compras y ventas siguen
				en la página de cada activo.
			</p>
			<ul class="closed-list">
				{#each closed as holding (holding.symbol)}
					<li>
						<span class="who">
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
						</span>
						<span class="type">{formatAssetType(holding.assetType)}</span>
					</li>
				{/each}
			</ul>
		</details>
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

	/* Símbolo, nombre y clase se ven igual en la tabla y en la lista de
	   cerradas: la regla vive aquí y alcanza a `portfolio-positions-table`. */
	.positions :global(.symbol) {
		display: block;
		font-family: var(--font-mono);
		font-weight: 600;
		color: var(--text);
		text-decoration: none;
	}

	.positions :global(.symbol:hover) {
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.positions :global(.name) {
		display: block;
		margin-top: 0.15rem;
		font-size: 0.78rem;
		line-height: 1.35;
		color: var(--text-muted);
		overflow-wrap: anywhere;
	}

	.positions :global(.type) {
		color: var(--text-muted);
		white-space: nowrap;
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

	/* Las cerradas van plegadas y en tono apagado: son historia del portafolio,
	   no cartera, y no deben competir con la tabla de arriba. */
	.closed {
		margin-top: 1.75rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
	}

	.closed summary {
		cursor: pointer;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.closed-count {
		margin-left: 0.3rem;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-dim);
	}

	.closed-hint {
		max-width: 68ch;
		margin: 0.6rem 0 0;
		font-size: 0.8rem;
		line-height: 1.45;
		color: var(--text-dim);
	}

	.closed-list {
		margin: 0.5rem 0 0;
		padding: 0;
		list-style: none;
	}

	.closed-list li {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: baseline;
		column-gap: 1rem;
		padding: 0.7rem 0;
		border-bottom: 1px solid var(--border);
		font-size: 0.85rem;
	}

	.closed-list li:last-child {
		border-bottom: none;
	}
</style>
