<script lang="ts">
	/*
	 * La lista consolidada: una fila por activo con lo que el usuario tiene de
	 * él, sumando todos sus portafolios.
	 *
	 * Es también la vista en tabla de la torta de al lado —mismos datos, mismo
	 * orden—, así que el reparto se puede leer entero aunque los colores no se
	 * distingan.
	 */
	import Card from '$lib/ui/card.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { assetTypeColor } from '$lib/shared/format/asset-type';
	import { formatQuantity, type AssetHoldingRow } from '../asset-holdings';

	let {
		rows,
		displayCurrency,
		formatValue,
		onGoToPortfolios
	}: {
		rows: AssetHoldingRow[];
		/** Moneda de la columna «Valor»: la misma en todas las filas. */
		displayCurrency: string;
		formatValue: (value: number) => string;
		/**
		 * Salida del estado vacío. Aquí no hay un portafolio al que agregar —esta
		 * vista los atraviesa todos—, así que lleva a elegir uno.
		 */
		onGoToPortfolios: () => void;
	} = $props();

	/*
	 * El precio va en la moneda del activo, no en la de la columna «Valor»: es
	 * lo que cotiza, no lo que se convirtió. Por eso no usa `formatValue`.
	 */
	function fmtPrice(row: AssetHoldingRow): string {
		if (row.marketPrice === null) return '—';

		return privacy.money(formatCurrency(row.marketPrice, row.currency));
	}
</script>

<Card variant="elevated" padding="none">
	<div class="holdings">
		<header class="panel-header">
			<h2>Tus activos</h2>
			<span>Valores en {displayCurrency}</span>
		</header>

		{#if rows.length > 0}
			<div class="table-scroll">
				<table>
					<caption class="sr-only">
						Activos que tienes, con su clase, las unidades sumadas entre portafolios, su precio y
						cuánto pesan sobre el total
					</caption>
					<thead>
						<tr>
							<th scope="col">Activo</th>
							<th scope="col">Tipo</th>
							<th scope="col" class="num">Cantidad</th>
							<th scope="col" class="num">Precio</th>
							<th scope="col" class="num">Valor</th>
							<th scope="col" class="weight-col">Peso</th>
							<th scope="col" class="num">Portafolios</th>
						</tr>
					</thead>
					<tbody>
						{#each rows as row (row.assetId)}
							<tr>
								<th scope="row" class="asset">
									<span class="ticker">{row.ticker}</span>
									<span class="name">{row.name}</span>
								</th>
								<td>
									<span class="type">
										<span class="type-dot" style="background: {assetTypeColor(row.assetType)}"
										></span>
										{row.typeLabel}
									</span>
								</td>
								<td class="num mono">{formatQuantity(row.quantity)}</td>
								<td class="num mono price">
									{fmtPrice(row)}
									{#if row.marketPrice === null}
										<span class="at-cost">a coste</span>
									{/if}
								</td>
								<td class="num mono value">
									{privacy.money(formatValue(row.value))}
									{#if !row.fxConverted}
										<span class="fx-flag" title="Sin tasa de cambio: importe sin convertir">
											sin convertir
										</span>
									{/if}
								</td>
								<td class="weight-col">
									<span class="weight">
										<span class="weight-track" aria-hidden="true">
											<span class="weight-fill" style="width: {Math.min(row.percent, 100)}%"></span>
										</span>
										<span class="weight-pct mono">{row.percent.toFixed(1)}%</span>
									</span>
								</td>
								<td class="num mono">{row.portfolios}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else}
			<EmptyState
				bordered
				title="Todavía no hay nada que listar"
				description="Cuando registres posiciones en tus portafolios, aquí aparecerá cuánto tienes de cada activo."
			>
				{#snippet action()}
					<button onclick={onGoToPortfolios} class="btn-go-portfolios">
						<svg
							width="18"
							height="18"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<path d="M12 5v14M5 12h14" />
						</svg>
						Ir a mis portafolios
					</button>
				{/snippet}
			</EmptyState>
		{/if}
	</div>
</Card>

<style>
	.holdings {
		padding: 1.5rem;
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.panel-header h2 {
		margin: 0;
		font-size: 1.15rem;
		color: var(--text);
	}

	.panel-header span {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.52);
	}

	/* La tabla es ancha por naturaleza (siete columnas): se desplaza dentro de
	   su propio contenedor en vez de empujar el ancho de la página. */
	.table-scroll {
		overflow-x: auto;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		min-width: 720px;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border: 0;
	}

	thead th {
		padding: 0.6rem 0.7rem;
		text-align: left;
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.6);
		border-bottom: 1px solid var(--border);
		white-space: nowrap;
	}

	tbody tr {
		border-bottom: 1px solid var(--border);
		transition: background 0.2s ease;
	}

	tbody tr:last-child {
		border-bottom: none;
	}

	tbody tr:hover {
		background: var(--surface-2);
	}

	tbody th,
	tbody td {
		padding: 0.75rem 0.7rem;
		font-size: 0.85rem;
		color: var(--text);
		text-align: left;
		font-weight: 400;
		vertical-align: middle;
	}

	.num {
		text-align: right;
	}

	.mono {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.asset {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}

	.ticker {
		font-family: var(--font-mono);
		font-weight: 700;
		color: var(--text);
	}

	.name {
		font-size: 0.76rem;
		color: var(--text-muted);
	}

	.type {
		display: inline-flex;
		align-items: center;
		gap: 0.45rem;
		white-space: nowrap;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.8);
	}

	.type-dot {
		width: 8px;
		height: 8px;
		border-radius: 2px;
		flex-shrink: 0;
	}

	.value {
		font-weight: 600;
	}

	.price,
	.value {
		white-space: nowrap;
	}

	.at-cost,
	.fx-flag {
		display: block;
		font-family: var(--font-body);
		font-size: 0.68rem;
		font-weight: 400;
		color: var(--amber);
	}

	.at-cost {
		color: var(--text-dim);
	}

	.weight-col {
		width: 9.5rem;
	}

	.weight {
		display: flex;
		align-items: center;
		gap: 0.6rem;
	}

	.weight-track {
		flex: 1;
		height: 6px;
		min-width: 3rem;
		border-radius: 3px;
		background: rgba(255, 255, 255, 0.07);
		overflow: hidden;
	}

	.weight-fill {
		display: block;
		height: 100%;
		border-radius: 3px;
		background: var(--amber);
	}

	.weight-pct {
		font-size: 0.78rem;
		color: rgba(236, 234, 229, 0.75);
		min-width: 3.2rem;
		text-align: right;
	}

	.btn-go-portfolios {
		display: inline-flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.8rem 1.4rem;
		border: none;
		border-radius: 10px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-go-portfolios:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}
</style>
