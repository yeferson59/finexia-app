<script lang="ts">
	/*
	 * Resumen de portafolios del dashboard.
	 *
	 * Era una rejilla de `div` con pinta de tabla: quien la recorre con lector de
	 * pantalla oía una fila de textos sueltos, sin saber a qué columna
	 * correspondía cada cifra. Ahora es una `<table>` de verdad, con cabeceras
	 * asociadas y el nombre del portafolio como cabecera de su fila.
	 */
	import { resolve } from '$app/paths';
	import Card from '$lib/ui/card.svelte';
	import CardHeader from '$lib/ui/card-header.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import Stat from '$lib/ui/stat.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';

	interface PortfolioSummary {
		id: string;
		name: string;
		type: string;
		baseCurrency: string;
		displayCurrency?: string;
		totalPositions: number;
		totalCostBase: string;
		totalMarketValue: string;
		totalGainLoss: string;
		totalGainLossPct: string;
	}

	const { summaries = [], currency = 'USD' }: { summaries: PortfolioSummary[]; currency?: string } =
		$props();

	const totalInvested = $derived(
		summaries.reduce((acc, s) => acc + parseFloat(s.totalCostBase || '0'), 0)
	);
	const totalValue = $derived(
		summaries.reduce((acc, s) => acc + parseFloat(s.totalMarketValue || '0'), 0)
	);
	const totalGainLoss = $derived(
		summaries.reduce((acc, s) => acc + parseFloat(s.totalGainLoss || '0'), 0)
	);
	const totalGainLossPct = $derived(totalInvested > 0 ? (totalGainLoss / totalInvested) * 100 : 0);

	function fmtMoney(value: number, currencyCode = currency): string {
		return privacy.money(formatCurrency(value, currencyCode));
	}

	function fmtPct(value: number): string {
		return new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		}).format(value);
	}
</script>

<Card>
	<CardHeader eyebrow="Resumen" title="Portafolios" />

	{#if summaries.length === 0}
		<EmptyState
			title="Sin portafolios registrados"
			description="Crea tu primer portafolio para agrupar los activos que tienes en cada plataforma."
		>
			{#snippet icon()}
				<svg
					width="48"
					height="48"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.5"
				>
					<rect x="2" y="3" width="20" height="14" rx="2" />
					<path d="M8 21h8M12 17v4" />
				</svg>
			{/snippet}
			{#snippet action()}
				<a href={resolve('/dashboard/portfolios/add')} class="empty-cta">Crear portafolio</a>
			{/snippet}
		</EmptyState>
	{:else}
		<DataTable class="portfolio-table" caption="Tus portafolios, con su valor actual y su ganancia">
			<thead>
				<tr>
					<th scope="col">Nombre</th>
					<th scope="col" class="col-type">Tipo</th>
					<th scope="col" class="num col-invested">Invertido</th>
					<th scope="col" class="num">Valor actual</th>
					<th scope="col" class="num">Ganancia/Pérdida</th>
				</tr>
			</thead>
			<tbody>
				{#each summaries as s (s.id)}
					{@const gainLoss = parseFloat(s.totalGainLoss || '0')}
					{@const marketValue = parseFloat(s.totalMarketValue || '0')}
					{@const costBase = parseFloat(s.totalCostBase || '0')}
					{@const pct = parseFloat(s.totalGainLossPct || '0')}
					{@const isUp = gainLoss >= 0}
					{@const rowCurrency = s.displayCurrency || s.baseCurrency}
					<tr>
						<th scope="row" class="name-cell">
							<a href={resolve(`/dashboard/portfolios/${s.id}`)} class="name">{s.name}</a>
							<span class="currency">{rowCurrency} · {s.totalPositions} pos.</span>
						</th>
						<td class="col-type"><span class="type-badge">{s.type}</span></td>
						<td class="num dim col-invested">{fmtMoney(costBase, rowCurrency)}</td>
						<td class="num">{fmtMoney(marketValue, rowCurrency)}</td>
						<td class="num gain-cell" class:positive={isUp} class:negative={!isUp}>
							<span>{isUp ? '+' : '−'}{fmtMoney(Math.abs(gainLoss), rowCurrency)}</span>
							<span class="pct">{isUp ? '+' : ''}{fmtPct(pct)}%</span>
						</td>
					</tr>
				{/each}
			</tbody>
		</DataTable>
	{/if}

	<div class="chart-stats">
		<Stat label="Total Invertido" value={fmtMoney(totalInvested)} />
		<Stat label="Valor Actual" tone="highlight" value={fmtMoney(totalValue)} />
		<Stat
			label="Ganancia Total"
			tone={totalGainLoss >= 0 ? 'positive' : 'negative'}
			value="{totalGainLoss >= 0 ? '+' : ''}{fmtPct(totalGainLossPct)}%"
		/>
	</div>
</Card>

<style>
	.empty-cta {
		display: inline-block;
		padding: 0.6rem 1.2rem;
		border: 1px solid rgba(212, 145, 42, 0.35);
		border-radius: 6px;
		background: rgba(212, 145, 42, 0.08);
		color: var(--amber-light);
		font-size: 0.82rem;
		font-weight: 600;
		text-decoration: none;
		transition:
			background 0.2s ease,
			border-color 0.2s ease;
	}

	.empty-cta:hover {
		background: rgba(212, 145, 42, 0.14);
		border-color: rgba(212, 145, 42, 0.55);
	}

	:global(.portfolio-table) {
		margin-bottom: 1.5rem;
	}

	/* La cabecera de fila es un `th`; hay que devolverle el aspecto de celda. */
	.name-cell {
		font-family: var(--font-body);
		font-size: 0.875rem;
		font-weight: 400;
		letter-spacing: normal;
		text-transform: none;
		color: var(--text);
		padding: 0.75rem 1.25rem;
		white-space: normal;
	}

	.name {
		display: block;
		font-weight: 600;
		color: var(--text);
		text-decoration: none;
		overflow-wrap: anywhere;
		transition: color 0.2s ease;
	}

	.name:hover {
		color: var(--amber-light);
	}

	.currency {
		display: block;
		margin-top: 0.2rem;
		font-size: 0.72rem;
		color: rgba(236, 234, 229, 0.45);
		text-transform: uppercase;
		font-family: var(--font-mono);
	}

	.type-badge {
		font-size: 0.72rem;
		color: rgba(212, 145, 42, 0.75);
		text-transform: uppercase;
		letter-spacing: 0.3px;
		font-weight: 600;
	}

	.dim {
		color: rgba(236, 234, 229, 0.55);
	}

	.gain-cell {
		white-space: nowrap;
	}

	.gain-cell .pct {
		display: block;
		font-size: 0.72rem;
	}

	.gain-cell.positive {
		color: var(--green);
	}

	.gain-cell.negative {
		color: var(--red);
	}

	.chart-stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1.5rem;
		padding-top: 1.5rem;
		border-top: 1px solid var(--border);
	}

	/* En pantallas estrechas la tabla ya hace scroll horizontal; ocultar las dos
	   columnas secundarias evita que lo necesite para lo esencial. */
	@media (max-width: 1024px) {
		.col-invested {
			display: none;
		}

		.chart-stats {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 768px) {
		.col-type {
			display: none;
		}

		.chart-stats {
			grid-template-columns: 1fr;
		}
	}
</style>
