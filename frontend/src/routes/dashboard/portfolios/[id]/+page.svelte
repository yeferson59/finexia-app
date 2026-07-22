<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { untrack } from 'svelte';
	import Card from '$lib/ui/card.svelte';
	import { PortfolioGrowth } from '$lib/features/dashboard';
	import { privacy } from '$lib/stores/privacy.svelte';
	import { formatCalendarDate } from '$lib/utils';
	import type { PageProps } from './$types';

	interface TopTransactionData {
		value: string;
		type: string;
		currency: string;
		assetTicker: string;
		assetName: string;
		transactionDate: string;
	}

	interface GrowthDataPoint {
		date: string;
		totalValue: string;
		totalCostBase: string;
		gainLoss: string;
		gainLossPct: string;
	}

	interface GrowthSummary {
		initialValue: string;
		currentValue: string;
		totalGrowthPct: string;
	}

	const { params, data }: PageProps = $props();

	const portfolio = $derived(data.portfolio);
	const risks = $derived(
		(data as unknown as { risks: { id: string; name: string }[] }).risks ?? []
	);
	const topTransaction = $derived(
		(data as unknown as { topTransaction: TopTransactionData | null }).topTransaction
	);
	const growth = $derived(
		(data as unknown as { growth: { points: GrowthDataPoint[]; summary: GrowthSummary } | null })
			.growth
	);

	let isEditing = $state(false);
	let isSubmitting = $state(false);
	let submitSuccess = $state(false);
	let submitError = $state('');

	let editName = $state('');
	let editDescription = $state('');
	let editType = $state('');
	let editRiskId = $state('');
	let editIsDefault = $state(false);

	$effect(() => {
		if (portfolio) {
			untrack(() => {
				editName = portfolio.name;
				editDescription = portfolio.description ?? '';
				editType = portfolio.type;
				editRiskId = (portfolio as unknown as { riskId: string }).riskId ?? '';
				editIsDefault = portfolio.isDefault;
			});
		}
	});

	const portfolioTypes = [
		{ value: 'stocks_etfs', label: 'Acciones y ETF' },
		{ value: 'stocks', label: 'Solo Acciones' },
		{ value: 'etfs', label: 'Solo ETFs' },
		{ value: 'cryptos', label: 'Criptomonedas' },
		{ value: 'bonds', label: 'Bonos y Renta Fija' },
		{ value: 'diversified', label: 'Portafolio Diverso' },
		{ value: 'forex', label: 'Divisas y Forex' },
		{ value: 'commodities', label: 'Commodities' },
		{ value: 'cash', label: 'Efectivo' }
	];

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

	const capitalPct = $derived(totalValue > 0 ? (totalCost / totalValue) * 100 : 0);
	const gainPct = $derived(totalValue > 0 ? (totalGainLoss / totalValue) * 100 : 0);
	const bestHolding = $derived(
		holdings.length > 0 ? holdings.reduce((a, b) => (a.gainLossPct > b.gainLossPct ? a : b)) : null
	);
	const worstHolding = $derived(
		holdings.length > 0 ? holdings.reduce((a, b) => (a.gainLossPct < b.gainLossPct ? a : b)) : null
	);
	const topConcentration = $derived(
		holdings.length > 0 ? holdings.reduce((a, b) => (a.allocation > b.allocation ? a : b)) : null
	);

	function formatPct(value: number): string {
		return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
	}

	function formatCurrency(value: number): string {
		return privacy.money(
			new Intl.NumberFormat('es-CO', {
				style: 'currency',
				currency: baseCurrency,
				minimumFractionDigits: 2
			}).format(value)
		);
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

	const DONUT_RADIUS = 60;
	const DONUT_CIRCUMFERENCE = 2 * Math.PI * DONUT_RADIUS;
	const DONUT_GAP = 3;

	const donutSegments = $derived.by(() => {
		const gap = typeBreakdown.length > 1 ? DONUT_GAP : 0;
		let acc = 0;
		return typeBreakdown.map((slice) => {
			const sliceLen = (slice.pct / 100) * DONUT_CIRCUMFERENCE;
			const dash = Math.max(sliceLen - gap, 0);
			const segment = {
				...slice,
				dasharray: `${dash} ${DONUT_CIRCUMFERENCE - dash}`,
				dashoffset: -acc
			};
			acc += sliceLen;
			return segment;
		});
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
		<div class="header-actions">
			<button
				onclick={() => (isEditing = !isEditing)}
				class="btn-edit"
				aria-label="Editar portafolio"
			>
				<svg
					width="16"
					height="16"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
					<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
				</svg>
				Editar
			</button>
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
	</div>
</header>

{#if submitSuccess}
	<div class="alert alert-success">Portafolio actualizado correctamente.</div>
{/if}

{#if submitError}
	<div class="alert alert-error">{submitError}</div>
{/if}

{#if isEditing}
	<Card variant="elevated" padding="sm">
		<form
			method="POST"
			action="?/updatePortfolio"
			class="edit-form"
			use:enhance={() => {
				isSubmitting = true;
				submitError = '';
				return async ({ result, update }) => {
					if (result.type === 'success') {
						submitSuccess = true;
						isEditing = false;
						setTimeout(() => (submitSuccess = false), 3000);
					} else if (result.type === 'failure') {
						submitError =
							(result.data as { error?: string })?.error ?? 'Error al actualizar el portafolio.';
					}
					await update({ reset: false });
					isSubmitting = false;
				};
			}}
		>
			<h3 class="edit-title">Editar portafolio</h3>

			<div class="form-group">
				<label for="edit-name">Nombre</label>
				<input
					id="edit-name"
					name="name"
					type="text"
					bind:value={editName}
					required
					minlength="2"
				/>
			</div>

			<div class="form-group">
				<label for="edit-description">Descripción</label>
				<textarea id="edit-description" name="description" rows="2" bind:value={editDescription}
				></textarea>
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="edit-type">Tipo</label>
					<select id="edit-type" name="type" bind:value={editType}>
						{#each portfolioTypes as pt (pt.value)}
							<option value={pt.value}>{pt.label}</option>
						{/each}
					</select>
				</div>

				<div class="form-group">
					<label for="edit-risk">Nivel de riesgo</label>
					<select id="edit-risk" name="riskId" bind:value={editRiskId}>
						{#each risks as risk (risk.id)}
							<option value={risk.id}>{risk.name}</option>
						{/each}
					</select>
				</div>
			</div>

			<div class="form-check">
				<input
					id="edit-default"
					name="isDefault"
					type="checkbox"
					bind:checked={editIsDefault}
					value="true"
				/>
				<label for="edit-default">Portafolio por defecto</label>
			</div>

			<div class="form-actions">
				<button type="button" class="btn-cancel" onclick={() => (isEditing = false)}
					>Cancelar</button
				>
				<button type="submit" class="btn-save" disabled={isSubmitting}>
					{isSubmitting ? 'Guardando…' : 'Guardar cambios'}
				</button>
			</div>
		</form>
	</Card>
{/if}

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

{#if growth}
	<section class="growth-section" aria-label="Crecimiento del portafolio">
		<PortfolioGrowth data={growth.points} summary={growth.summary} />
	</section>
{/if}

<section class="stats-grid">
	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Capital invertido</p>
		<h2 class="hero-value">{formatCurrency(totalCost)}</h2>
		<div class="progress-track">
			<div
				class="progress-fill"
				style="width: {totalValue > 0 ? Math.min((totalCost / totalValue) * 100, 100) : 0}%"
			></div>
		</div>
		<p class="hero-delta">Valor actual: {formatCurrency(totalValue)}</p>
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Composición del portafolio</p>
		{#if totalValue > 0}
			<div class="composition-bar">
				<div
					class="comp-segment comp-capital"
					style="width: {Math.max(capitalPct, 0)}%"
					title="Capital: {capitalPct.toFixed(1)}%"
				></div>
				<div
					class="comp-segment {gainPct >= 0 ? 'comp-gain' : 'comp-loss'}"
					style="width: {Math.abs(gainPct)}%"
					title="Ganancia: {gainPct.toFixed(1)}%"
				></div>
			</div>
			<p class="comp-labels">
				<span class="comp-label-capital">{capitalPct.toFixed(1)}% capital</span>
				<span class="comp-sep">·</span>
				<span class={gainPct >= 0 ? 'comp-label-gain' : 'comp-label-loss'}
					>{gainPct >= 0 ? '+' : ''}{gainPct.toFixed(1)}% {gainPct >= 0
						? 'ganancia'
						: 'pérdida'}</span
				>
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin datos suficientes</p>
		{/if}
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Mejor activo</p>
		{#if bestHolding}
			<h2 class="hero-value">{bestHolding.symbol}</h2>
			<p class="hero-delta {bestHolding.gainLossPct >= 0 ? 'positive' : 'negative'}">
				{formatPct(bestHolding.gainLossPct)} · {bestHolding.name}
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin activos</p>
		{/if}
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Peor activo</p>
		{#if worstHolding}
			<h2 class="hero-value">{worstHolding.symbol}</h2>
			<p class="hero-delta {worstHolding.gainLossPct >= 0 ? 'positive' : 'negative'}">
				{formatPct(worstHolding.gainLossPct)} · {worstHolding.name}
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin activos</p>
		{/if}
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Concentración</p>
		{#if topConcentration}
			<h2 class="hero-value">{topConcentration.allocation.toFixed(1)}%</h2>
			<p class="hero-delta">
				{topConcentration.symbol} · {topConcentration.name}
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin activos</p>
		{/if}
	</Card>

	<Card variant="elevated" padding="sm">
		<p class="eyebrow">Transacción más alta</p>
		{#if topTransaction}
			<h2 class="hero-value">{formatCurrency(parseFloat(topTransaction.value))}</h2>
			<p class="hero-delta">
				{topTransaction.assetTicker} · {topTransaction.type} ·
				{formatCalendarDate(topTransaction.transactionDate, {
					year: 'numeric',
					month: 'short',
					day: 'numeric'
				})}
			</p>
		{:else}
			<h2 class="hero-value">—</h2>
			<p class="hero-delta">Sin transacciones</p>
		{/if}
	</Card>
</section>

{#if typeBreakdown.length > 0}
	<Card variant="elevated" padding="none">
		<div class="distribution">
			<header class="panel-header">
				<h2>Distribución por tipo</h2>
				<span>{typeBreakdown.length} {typeBreakdown.length === 1 ? 'tipo' : 'tipos'}</span>
			</header>

			<div class="distribution-body">
				<div class="donut-wrap">
					<svg
						class="donut"
						viewBox="0 0 160 160"
						role="img"
						aria-label="Distribución por tipo de activo"
					>
						<circle
							cx="80"
							cy="80"
							r={DONUT_RADIUS}
							fill="none"
							stroke="rgba(236, 234, 229, 0.08)"
							stroke-width="22"
						/>
						<g transform="rotate(-90 80 80)">
							{#each donutSegments as slice (slice.type)}
								<circle
									cx="80"
									cy="80"
									r={DONUT_RADIUS}
									fill="none"
									stroke={slice.color}
									stroke-width="22"
									stroke-linecap="round"
									stroke-dasharray={slice.dasharray}
									stroke-dashoffset={slice.dashoffset}
								>
									<title>{slice.label}: {slice.pct.toFixed(1)}%</title>
								</circle>
							{/each}
						</g>
						<text
							x="80"
							y="76"
							text-anchor="middle"
							fill="var(--text)"
							font-size="15"
							font-family="var(--font-mono)"
							font-weight="700">{formatCurrency(totalValue)}</text
						>
						<text
							x="80"
							y="94"
							text-anchor="middle"
							fill="rgba(236, 234, 229, 0.5)"
							font-size="9"
							font-family="var(--font-mono)">Valor total</text
						>
					</svg>
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

	.growth-section {
		margin-bottom: 1.5rem;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.progress-track {
		height: 6px;
		border-radius: 999px;
		background: rgba(236, 234, 229, 0.1);
		overflow: hidden;
		margin: 0.6rem 0 0.5rem;
	}

	.progress-fill {
		height: 100%;
		border-radius: inherit;
		background: var(--amber);
		transition: width 0.4s ease;
	}

	.composition-bar {
		display: flex;
		height: 10px;
		border-radius: 999px;
		overflow: hidden;
		background: rgba(236, 234, 229, 0.1);
		gap: 2px;
		margin: 0.6rem 0 0.5rem;
	}

	.comp-segment {
		height: 100%;
		border-radius: 999px;
		min-width: 2px;
		transition: width 0.4s ease;
	}

	.comp-capital {
		background: var(--amber);
	}

	.comp-gain {
		background: var(--green);
	}

	.comp-loss {
		background: var(--red);
	}

	.comp-labels {
		margin: 0;
		font-size: 0.78rem;
		display: flex;
		align-items: center;
		gap: 0.4rem;
		flex-wrap: wrap;
	}

	.comp-sep {
		color: rgba(236, 234, 229, 0.3);
	}

	.comp-label-capital {
		color: var(--amber);
		font-weight: 600;
	}

	.comp-label-gain {
		color: var(--green);
		font-weight: 600;
	}

	.comp-label-loss {
		color: var(--red);
		font-weight: 600;
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

	.distribution-body {
		display: flex;
		align-items: center;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.donut-wrap {
		flex-shrink: 0;
		width: 160px;
	}

	.donut {
		width: 100%;
		height: auto;
		display: block;
	}

	.donut circle {
		transition:
			stroke-dasharray 0.4s ease,
			opacity 0.2s ease;
	}

	.donut circle:hover {
		opacity: 0.85;
	}

	.type-legend {
		flex: 1;
		min-width: 220px;
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

	.header-actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		flex-wrap: wrap;
	}

	.btn-edit {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.25rem;
		border: 1px solid rgba(212, 145, 42, 0.4);
		border-radius: 10px;
		background: transparent;
		color: var(--amber);
		font-weight: 600;
		font-family: var(--font-body);
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		white-space: nowrap;
	}

	.btn-edit:hover {
		background: rgba(212, 145, 42, 0.12);
		border-color: rgba(212, 145, 42, 0.7);
	}

	.alert {
		margin-bottom: 1.25rem;
		padding: 0.85rem 1.25rem;
		border-radius: 10px;
		font-size: 0.9rem;
	}

	.alert-success {
		background: rgba(34, 197, 94, 0.12);
		border: 1px solid rgba(34, 197, 94, 0.3);
		color: var(--green);
	}

	.alert-error {
		background: rgba(239, 68, 68, 0.12);
		border: 1px solid rgba(239, 68, 68, 0.3);
		color: var(--red);
	}

	.edit-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.edit-title {
		margin: 0 0 0.25rem;
		font-size: 1rem;
		font-weight: 600;
		color: var(--text);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.form-group label {
		font-size: 0.8rem;
		font-weight: 600;
		letter-spacing: 0.4px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.55);
	}

	.form-group input,
	.form-group textarea,
	.form-group select {
		padding: 0.7rem 0.9rem;
		border: 1px solid rgba(212, 145, 42, 0.25);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.04);
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.9rem;
		transition: border-color 0.2s ease;
	}

	.form-group input:focus,
	.form-group textarea:focus,
	.form-group select:focus {
		outline: none;
		border-color: rgba(212, 145, 42, 0.6);
	}

	.form-group textarea {
		resize: vertical;
	}

	.form-group select option {
		background: #1a1209;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.form-check {
		display: flex;
		align-items: center;
		gap: 0.65rem;
	}

	.form-check input[type='checkbox'] {
		width: 16px;
		height: 16px;
		cursor: pointer;
		accent-color: var(--amber);
	}

	.form-check label {
		font-size: 0.9rem;
		color: var(--text);
		cursor: pointer;
	}

	.form-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
		padding-top: 0.25rem;
	}

	.btn-cancel {
		padding: 0.7rem 1.25rem;
		border: 1px solid rgba(236, 234, 229, 0.2);
		border-radius: 8px;
		background: transparent;
		color: rgba(236, 234, 229, 0.7);
		font-family: var(--font-body);
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-cancel:hover {
		border-color: rgba(236, 234, 229, 0.4);
		color: var(--text);
	}

	.btn-save {
		padding: 0.7rem 1.5rem;
		border: none;
		border-radius: 8px;
		background: var(--amber);
		color: #0d0800;
		font-family: var(--font-body);
		font-weight: 700;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-save:hover:not(:disabled) {
		box-shadow: 0 6px 18px rgba(212, 145, 42, 0.3);
	}

	.btn-save:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	@media (max-width: 1024px) {
		.cards-grid {
			grid-template-columns: 1fr;
		}

		.stats-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}

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
