<script lang="ts">
	import type { PageProps } from './$types';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { privacy } from '$lib/shared/privacy.svelte';
	import {
		AssetPositionHeader,
		AssetPositionSummary,
		AssetInfoPanel,
		AssetTransactionHistory,
		computePosition
	} from '$lib/features/portfolio';

	const { params, data, form }: PageProps = $props();

	const entries = $derived(data.entries);
	const transactions = $derived(data.transactions ?? []);
	const txnMeta = $derived(data.txnMeta);

	// Métricas de la posición a partir de los agregados de las entradas
	// (exactas con independencia de la paginación de transacciones).
	const position = $derived(computePosition(entries, data.portfolioTotalValue));

	function formatCurrency(value: number, decimals = 2): string {
		return privacy.money(
			new Intl.NumberFormat('es-CO', {
				style: 'currency',
				currency: position?.costCurrency || 'USD',
				currencyDisplay: 'narrowSymbol',
				minimumFractionDigits: decimals,
				maximumFractionDigits: decimals
			}).format(value)
		);
	}

	function goBack() {
		goto(resolve('/dashboard/portfolios/[id]', { id: params.id }));
	}
</script>

<svelte:head>
	<title>{params.symbol} - Portfolio - FINEXIA</title>
	<meta name="description" content="Detalles de la posición {params.symbol}" />
</svelte:head>

<div class="container">
	<button class="btn-back" onclick={goBack}>
		<svg
			width="20"
			height="20"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
		>
			<path d="M15 19l-7-7 7-7" />
		</svg>
		Volver
	</button>

	{#if !position}
		<div class="empty-state">
			<p>No se encontraron entradas para <strong>{params.symbol}</strong> en este portafolio.</p>
		</div>
	{:else}
		<AssetPositionHeader {position} {formatCurrency} />

		<AssetPositionSummary {position} {formatCurrency} />

		<AssetInfoPanel {position} transactionsCount={txnMeta.total} />

		<AssetTransactionHistory
			portfolioId={params.id}
			symbol={params.symbol}
			{entries}
			{transactions}
			{txnMeta}
			marketPrice={position.marketPrice}
			{form}
			{formatCurrency}
		/>
	{/if}
</div>

<style>
	.container {
		max-width: 1400px;
		margin: 0 auto;
	}

	.btn-back {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.65rem 1rem;
		margin-bottom: 1.5rem;
		background: transparent;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 8px;
		color: var(--amber);
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		font-family: var(--font-body);
	}

	.btn-back:hover {
		background: var(--border);
		border-color: var(--amber);
	}

	.empty-state {
		padding: 3rem;
		text-align: center;
		color: rgba(236, 234, 229, 0.5);
		border: 1px dashed var(--border-strong);
		border-radius: 12px;
	}
</style>
