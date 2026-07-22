<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PortfolioGrowth from '$components/dashboard/portfolio-growth.svelte';
	import { privacy } from '$lib/stores/privacy.svelte';
	import {
		PortfolioEditForm,
		PortfolioSummaryCards,
		PortfolioStatsCards,
		AllocationDonut,
		HoldingsTable,
		groupHoldings,
		computeTypeBreakdown,
		type RawHolding,
		type TopTransactionData,
		type GrowthDataPoint,
		type GrowthSummary
	} from '$lib/features/portfolio';
	import type { PageProps } from './$types';

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
	let submitSuccess = $state(false);
	let submitError = $state('');

	// Group entries by ticker so the same asset held in multiple platforms
	// appears as a single row with aggregated quantity and cost basis.
	const holdings = $derived(groupHoldings((portfolio?.holdings ?? []) as RawHolding[]));

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

	function formatCurrency(value: number): string {
		return privacy.money(
			new Intl.NumberFormat('es-CO', {
				style: 'currency',
				currency: baseCurrency,
				minimumFractionDigits: 2
			}).format(value)
		);
	}

	const typeBreakdown = $derived(computeTypeBreakdown(holdings));

	function goBack() {
		goto(resolve('/dashboard/portfolios'));
	}

	function addAsset() {
		goto(resolve('/dashboard/portfolios/[id]/add', { id: params.id }));
	}

	function viewAssetDetails(symbol: string) {
		goto(resolve('/dashboard/portfolios/[id]/assets/[symbol]', { id: params.id, symbol }));
	}

	function startEditing() {
		submitError = '';
		isEditing = true;
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
			<button onclick={startEditing} class="btn-edit" aria-label="Editar portafolio">
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
	<PortfolioEditForm
		portfolio={portfolio as unknown as {
			name: string;
			description?: string | null;
			type: string;
			riskId?: string;
			isDefault: boolean;
		}}
		{risks}
		onCancel={() => (isEditing = false)}
		onSaved={() => {
			submitSuccess = true;
			isEditing = false;
			setTimeout(() => (submitSuccess = false), 3000);
		}}
		onError={(msg) => (submitError = msg)}
	/>
{/if}

<PortfolioSummaryCards
	{totalValue}
	{totalCost}
	{baseCurrency}
	{totalGainLoss}
	{totalGainLossPct}
	riskName={portfolio?.riskName}
	holdingsCount={holdings.length}
	{formatCurrency}
/>

{#if growth}
	<section class="growth-section" aria-label="Crecimiento del portafolio">
		<PortfolioGrowth data={growth.points} summary={growth.summary} />
	</section>
{/if}

<PortfolioStatsCards
	{totalValue}
	{totalCost}
	{capitalPct}
	{gainPct}
	{bestHolding}
	{worstHolding}
	{topConcentration}
	{topTransaction}
	{formatCurrency}
/>

{#if typeBreakdown.length > 0}
	<AllocationDonut {typeBreakdown} {totalValue} {formatCurrency} />
{/if}

<HoldingsTable {holdings} {formatCurrency} onViewAsset={viewAssetDetails} onAddAsset={addAsset} />

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

	.header-top > div {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
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

	.growth-section {
		margin-bottom: 1.5rem;
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
	}
</style>
