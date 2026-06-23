<script lang="ts">
	import type { PageProps } from './$types';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Card from '$components/ui/card.svelte';

	const { params, data }: PageProps = $props();

	const portfolio = $derived(data.portfolio);

	interface HoldingView {
		symbol: string;
		name: string;
		assetType: string;
		quantity: number;
		marketPrice: number;
		costBasis: number;
		value: number;
		gainLoss: number;
		gainLossPct: number;
		allocation: number;
	}

	// Group entries by ticker so the same asset held in multiple platforms
	// appears as a single row with aggregated quantity and cost basis.
	const holdings = $derived.by<HoldingView[]>(() => {
		const list = portfolio?.holdings ?? [];
		const grouped: Record<string, HoldingView> = {};

		for (const h of list) {
			const quantity = parseFloat(h.quantity) || 0;
			const costPrice = parseFloat(h.price) || 0;
			const marketPrice = parseFloat(h.marketPrice) || costPrice;
			const costBasis = quantity * costPrice;
			const value = quantity * marketPrice;

			const existing = grouped[h.ticker];
			if (existing) {
				existing.quantity += quantity;
				existing.costBasis += costBasis;
				existing.value += value;
				existing.gainLoss = existing.value - existing.costBasis;
				existing.gainLossPct =
					existing.costBasis > 0 ? (existing.gainLoss / existing.costBasis) * 100 : 0;
			} else {
				grouped[h.ticker] = {
					symbol: h.ticker,
					name: h.name,
					assetType: h.assetType,
					quantity,
					marketPrice,
					costBasis,
					value,
					gainLoss: value - costBasis,
					gainLossPct: costBasis > 0 ? ((value - costBasis) / costBasis) * 100 : 0,
					allocation: 0
				};
			}
		}

		const rows = Object.values(grouped);
		const total = rows.reduce((sum, h) => sum + h.value, 0);
		return rows.map((h) => ({ ...h, allocation: total > 0 ? (h.value / total) * 100 : 0 }));
	});

	const totalValue = $derived(holdings.reduce((sum, h) => sum + h.value, 0));
	const totalCost = $derived(holdings.reduce((sum, h) => sum + h.costBasis, 0));
	const totalGainLoss = $derived(totalValue - totalCost);
	const totalGainLossPct = $derived(totalCost > 0 ? (totalGainLoss / totalCost) * 100 : 0);
	const baseCurrency = $derived(portfolio?.baseCurrency?.trim() || 'USD');

	function formatPct(value: number): string {
		return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
	}

	function formatCurrency(value: number): string {
		return new Intl.NumberFormat('es-CO', {
			style: 'currency',
			currency: baseCurrency,
			minimumFractionDigits: 2
		}).format(value);
	}

	const ASSET_TYPE_LABELS: Record<string, string> = {
		stock: 'Acciones',
		etf: 'ETFs',
		crypto: 'Cripto',
		bond: 'Bonos',
		cash: 'Efectivo',
		real_estate: 'Inmobiliario',
		commodity: 'Materias primas',
		other: 'Otros'
	};

	const ASSET_TYPE_COLORS: Record<string, string> = {
		stock: '#4f8ef7',
		etf: '#d4912a',
		crypto: '#f97316',
		bond: '#22c55e',
		cash: '#94a3b8',
		real_estate: '#a855f7',
		commodity: '#b45309',
		other: '#6b7280'
	};

	const typeBreakdown = $derived.by(() => {
		const grouped: Record<string, { label: string; value: number; color: string }> = {};
		for (const h of holdings) {
			const key = h.assetType;
			if (!grouped[key]) {
				grouped[key] = {
					label: ASSET_TYPE_LABELS[key] ?? key,
					value: 0,
					color: ASSET_TYPE_COLORS[key] ?? '#6b7280'
				};
			}
			grouped[key].value += h.value;
		}
		const total = Object.values(grouped).reduce((s, v) => s + v.value, 0);
		return Object.entries(grouped)
			.map(([type, data]) => ({
				type,
				...data,
				pct: total > 0 ? (data.value / total) * 100 : 0
			}))
			.sort((a, b) => b.value - a.value);
	});

	function goBack() {
		goto(resolve('/dashboard/portfolios'));
	}

	function addAsset() {
		goto(resolve('/dashboard/portfolios/[id]/add', { id: params.id }));
	}

	function viewAssetDetails(symbol: string) {
		goto(resolve('/dashboard/portfolios/[id]/assets/[symbol]', { id: params.id, symbol }));
	}
</script>

<svelte:head>
	<title>Portafolio - FINEXIA</title>
	<meta name="description" content="Detalle de posiciones y asignación de portafolio" />
</svelte:head>

<header class="page-header">
	<div class="header-top">
		<div>
			<button onclick={goBack} class="btn-back" aria-label="Volver a portafolios">
				<svg
					width="20"
					height="20"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<path d="M19 12H5M12 19l-7-7 7-7" />
				</svg>
			</button>
			<h1 class="page-title">{portfolio?.name ?? 'Portafolio'}</h1>
			<p class="page-subtitle">
				{portfolio?.description || 'Visión detallada de posiciones y asignación.'}
			</p>
		</div>
		<button onclick={addAsset} class="btn-add-asset">
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
			Agregar Activo
		</button>
	</div>
</header>

<section class="cards-grid">
	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Valor de mercado</p>
		<h2 class="hero-value">{formatCurrency(totalValue)}</h2>
		<p class="hero-delta">Costo: {formatCurrency(totalCost)} · {baseCurrency}</p>
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Ganancia / Pérdida</p>
		<h2 class="hero-value {totalGainLoss >= 0 ? 'positive' : 'negative'}">
			{formatCurrency(totalGainLoss)}
		</h2>
		<p class="hero-delta {totalGainLoss >= 0 ? 'positive' : 'negative'}">
			{formatPct(totalGainLossPct)} sobre costo
		</p>
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Riesgo · Activos</p>
		<h2 class="hero-value">{portfolio?.riskName ?? '—'}</h2>
		<p class="hero-delta">
			{holdings.length}
			{holdings.length === 1 ? 'activo' : 'activos'}
		</p>
	</Card>
</section>

{#if typeBreakdown.length > 0}
	<Card variant="elevated" padding="none">
		<div class="distribution">
			<header class="panel-header">
				<h2>Distribución por tipo</h2>
				<span>{typeBreakdown.length} {typeBreakdown.length === 1 ? 'tipo' : 'tipos'}</span>
			</header>

			<div class="stacked-bar">
				{#each typeBreakdown as slice (slice.type)}
					<div
						class="stack-segment"
						style="width: {slice.pct}%; background: {slice.color};"
						title="{slice.label}: {slice.pct.toFixed(1)}%"
					></div>
				{/each}
			</div>

			<div class="type-legend">
				{#each typeBreakdown as slice (slice.type)}
					<div class="legend-item">
						<span class="legend-dot" style="background: {slice.color};"></span>
						<span class="legend-label">{slice.label}</span>
						<span class="legend-pct">{slice.pct.toFixed(1)}%</span>
						<span class="legend-value">{formatCurrency(slice.value)}</span>
					</div>
				{/each}
			</div>
		</div>
	</Card>
{/if}

<Card variant="elevated" padding="none">
	<div class="holdings">
		<header class="panel-header">
			<h2>Posiciones</h2>
			<span>{holdings.length} {holdings.length === 1 ? 'activo' : 'activos'}</span>
		</header>

		{#if holdings.length > 0}
			<div class="holdings-list">
				{#each holdings as holding (holding.symbol)}
					<button
						class="holding-row"
						onclick={() => viewAssetDetails(holding.symbol)}
						aria-label={`Ver detalles de ${holding.symbol}`}
					>
						<div class="holding-main">
							<p class="symbol">{holding.symbol}</p>
							<p class="name">{holding.name}</p>
						</div>
						<div class="bar-wrap" aria-label={`Asignación ${holding.allocation.toFixed(1)}%`}>
							<div class="bar-fill" style={`width: ${holding.allocation}%`}></div>
						</div>
						<p class="metric">{formatCurrency(holding.value)}</p>
						<p class="metric delta {holding.gainLoss >= 0 ? 'positive' : 'negative'}">
							{formatPct(holding.gainLossPct)}
						</p>
						<svg
							class="arrow-icon"
							width="20"
							height="20"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<path d="M9 18l6-6-6-6" />
						</svg>
					</button>
				{/each}
			</div>
		{:else}
			<div class="empty-holdings">
				<p class="empty-text">Este portafolio aún no tiene activos.</p>
				<button onclick={addAsset} class="btn-add-asset">
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
					Agregar tu primer activo
				</button>
			</div>
		{/if}
	</div>
</Card>

<style>
	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.header-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
		font-weight: 300;
		letter-spacing: -0.02em;
		color: var(--text);
		font-family: var(--font-display);
	}

	.page-subtitle {
		margin: 0;
		color: rgba(236, 234, 229, 0.62);
		font-size: 1rem;
	}

	.btn-add-asset {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.85rem 1.5rem;
		border: none;
		border-radius: 10px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
		white-space: nowrap;
	}

	.btn-add-asset:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.btn-back {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		padding: 0;
		margin-right: 1rem;
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: var(--border);
		color: var(--amber);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-back:hover {
		background: rgba(212, 145, 42, 0.2);
		border-color: rgba(212, 145, 42, 0.5);
		transform: translateX(-2px);
	}

	.header-top > div {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
	}

	.btn-add-asset {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.85rem 1.5rem;
		border: none;
		border-radius: 10px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
		white-space: nowrap;
	}

	.btn-add-asset:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.cards-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.eyebrow {
		margin: 0 0 0.55rem;
		font-size: 0.72rem;
		letter-spacing: 0.7px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.46);
	}

	.hero-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		margin: 0;
		font-size: 1.6rem;
		color: var(--text);
	}

	.hero-delta {
		margin: 0.4rem 0 0;
		font-size: 0.82rem;
		color: rgba(236, 234, 229, 0.55);
	}

	.distribution {
		padding: 1.5rem;
	}

	.stacked-bar {
		display: flex;
		height: 12px;
		border-radius: 999px;
		overflow: hidden;
		background: rgba(236, 234, 229, 0.08);
		gap: 2px;
		margin-bottom: 1.25rem;
	}

	.stack-segment {
		height: 100%;
		border-radius: 999px;
		transition: opacity 0.2s ease;
		min-width: 4px;
	}

	.stack-segment:hover {
		opacity: 0.8;
	}

	.type-legend {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		gap: 0.55rem;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 0.55rem;
		padding: 0.55rem 0.7rem;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.022);
	}

	.legend-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.legend-label {
		flex: 1;
		font-size: 0.82rem;
		color: var(--text);
	}

	.legend-pct {
		font-size: 0.82rem;
		font-weight: 700;
		font-family: var(--font-mono);
		color: var(--text);
	}

	.legend-value {
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.5);
		font-family: var(--font-mono);
		min-width: 4rem;
		text-align: right;
	}

	.holdings {
		padding: 1.5rem;
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}

	.panel-header h2 {
		margin: 0;
		font-size: 1.2rem;
		color: var(--text);
	}

	.panel-header span {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.52);
	}

	.holdings-list {
		display: grid;
		gap: 0.75rem;
	}

	.empty-holdings {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		padding: 2.5rem 1.5rem;
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.022);
		border: 1px dashed rgba(212, 145, 42, 0.25);
		text-align: center;
	}

	.empty-text {
		margin: 0;
		font-size: 0.95rem;
		color: rgba(236, 234, 229, 0.6);
	}

	.holding-row {
		display: grid;
		grid-template-columns: 1.25fr 1.4fr 0.8fr 0.45fr auto;
		align-items: center;
		gap: 0.9rem;
		padding: 0.85rem;
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.022);
		cursor: pointer;
		transition: all 0.3s ease;
		border: 1px solid rgba(212, 145, 42, 0);
	}

	.holding-row:hover {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(212, 145, 42, 0.2);
		transform: translateX(4px);
	}

	.holding-main .symbol {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 700;
		color: var(--amber-light);
	}

	.holding-main .name {
		margin: 0.15rem 0 0;
		font-size: 0.78rem;
		color: rgba(236, 234, 229, 0.6);
	}

	.bar-wrap {
		height: 8px;
		border-radius: 999px;
		background: rgba(236, 234, 229, 0.12);
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		border-radius: inherit;
		background: linear-gradient(90deg, var(--amber), var(--green));
	}

	.metric {
		margin: 0;
		font-size: 0.84rem;
		font-weight: 600;
		text-align: right;
		color: var(--text);
	}

	.metric.delta {
		color: var(--amber-light);
	}

	.metric.delta.positive,
	.hero-value.positive {
		color: var(--green);
	}

	.metric.delta.negative,
	.hero-value.negative {
		color: var(--red);
	}

	.hero-delta.positive {
		color: var(--green);
	}

	.hero-delta.negative {
		color: var(--red);
	}

	.arrow-icon {
		color: rgba(212, 145, 42, 0.4);
		transition: all 0.3s ease;
	}

	.holding-row:hover .arrow-icon {
		color: var(--amber-light);
		transform: translateX(2px);
	}

	@media (max-width: 1024px) {
		.cards-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}

		.header-top {
			flex-direction: column;
		}

		.btn-add-asset {
			width: 100%;
			justify-content: center;
		}

		.holding-row {
			grid-template-columns: 1fr;
			gap: 0.55rem;
		}

		.metric {
			text-align: left;
		}
	}
</style>
