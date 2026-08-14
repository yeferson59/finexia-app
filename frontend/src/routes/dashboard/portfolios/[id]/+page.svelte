<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { PortfolioGrowth } from '$lib/features/dashboard';
	import { privacy } from '$lib/shared/privacy.svelte';
	import {
		PortfolioEditForm,
		PortfolioSummaryCards,
		PortfolioStatsCards,
		AllocationDonut,
		HoldingsTable,
		PortfolioDetailHeader,
		groupHoldings,
		computeTypeBreakdown
	} from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { params, data }: PageProps = $props();

	const portfolio = $derived(data.portfolio);
	const risks = $derived(data.risks);
	const topTransaction = $derived(data.topTransaction);
	const growth = $derived(data.growth);

	let isEditing = $state(false);
	let submitSuccess = $state(false);
	let submitError = $state('');

	// Group entries by ticker so the same asset held in multiple platforms
	// appears as a single row with aggregated quantity and cost basis.
	const holdings = $derived(groupHoldings(portfolio?.holdings ?? []));

	const totalValue = $derived(holdings.reduce((sum, h) => sum + h.value, 0));
	const totalCost = $derived(holdings.reduce((sum, h) => sum + h.costBasis, 0));
	const totalGainLoss = $derived(totalValue - totalCost);
	const totalGainLossPct = $derived(totalCost > 0 ? (totalGainLoss / totalCost) * 100 : 0);
	const baseCurrency = $derived(portfolio?.baseCurrency?.trim() || 'USD');

	// Posiciones que el backend no pudo convertir por falta de tasa: sus
	// importes están en su moneda nativa, así que los totales de arriba mezclan
	// monedas y hay que decirlo en vez de presentarlos como comparables.
	const unconverted = $derived(holdings.filter((h) => !h.fxConverted));

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

<PortfolioDetailHeader
	name={portfolio?.name ?? 'Portafolio'}
	description={portfolio?.description}
	onBack={goBack}
	onEdit={startEditing}
	onAddAsset={addAsset}
/>

{#if submitSuccess}
	<div class="alert alert-success">Portafolio actualizado correctamente.</div>
{/if}

{#if submitError}
	<div class="alert alert-error">{submitError}</div>
{/if}

{#if isEditing && portfolio}
	<PortfolioEditForm
		{portfolio}
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

{#if unconverted.length > 0}
	<p class="alert alert-warning">
		Sin tasa de cambio para {unconverted.map((h) => `${h.symbol} (${h.currency})`).join(', ')}: esos
		importes van sin convertir a {baseCurrency}, así que los totales de abajo mezclan monedas.
	</p>
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

	.alert-warning {
		margin-top: 0;
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.3);
		color: rgba(236, 234, 229, 0.8);
		line-height: 1.5;
	}

	.growth-section {
		margin-bottom: 1.5rem;
	}
</style>
