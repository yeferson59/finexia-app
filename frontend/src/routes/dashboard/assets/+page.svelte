<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import {
		AssetConcentration,
		AssetHoldingsTable,
		toAssetHoldingRows
	} from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const rows = $derived(toAssetHoldingRows(data.holdings ?? []));
	const displayCurrency = $derived(data.currency);

	const totalValue = $derived(rows.reduce((sum, row) => sum + row.value, 0));

	// Posiciones que el backend no pudo convertir: siguen sumadas con su importe
	// nativo, así que el total mezcla monedas y hay que decirlo en vez de
	// presentarlo como una cifra limpia.
	const unconverted = $derived(rows.filter((row) => !row.fxConverted).length);

	const PER_PAGE = 15;
	let page = $state(1);
	// La torta reparte la cartera entera; la tabla se pagina. Si la torta leyera
	// la página, sus porciones cambiarían al pasar de hoja.
	const pagedRows = $derived(rows.slice((page - 1) * PER_PAGE, page * PER_PAGE));

	function fmt(value: number): string {
		return formatCurrency(value, displayCurrency);
	}

	function goToPortfolios() {
		goto(resolve('/dashboard/portfolios'));
	}
</script>

<svelte:head>
	<title>Mis Activos - FINEXIA</title>
	<meta
		name="description"
		content="Todo lo que tienes de cada activo, sumado a través de tus portafolios"
	/>
</svelte:head>

<PageHeader
	title="Mis Activos"
	subtitle="Cuánto tienes de cada activo, sumando todos tus portafolios."
/>

<section class="totals">
	<article class="panel total-card">
		<p class="eyebrow">Valor total</p>
		<h2 class="hero-value">{privacy.money(fmt(totalValue))}</h2>
		<p class="hero-note">En {displayCurrency}</p>
	</article>

	<article class="panel total-card">
		<p class="eyebrow">Activos distintos</p>
		<h2 class="hero-value">{rows.length}</h2>
		<p class="hero-note">
			{rows.length === 1 ? 'Una posición consolidada' : 'Posiciones consolidadas'}
		</p>
	</article>
</section>

{#if unconverted > 0}
	<p class="fx-note">
		{unconverted === 1 ? 'Un activo queda' : `${unconverted} activos quedan`} sin convertir a {displayCurrency}:
		no hay tasa para {unconverted === 1 ? 'su moneda' : 'sus monedas'}. {unconverted === 1
			? 'Su importe va'
			: 'Sus importes van'} a valor nominal, así que el total y los pesos mezclan monedas.
	</p>
{/if}

<div class="panel-stack">
	<AssetConcentration {rows} {displayCurrency} formatCurrency={fmt} />

	<AssetHoldingsTable
		rows={pagedRows}
		{displayCurrency}
		formatValue={fmt}
		onGoToPortfolios={goToPortfolios}
	/>

	<Pagination bind:page total={rows.length} perPage={PER_PAGE} label="activos" />
</div>

<style>
	.totals {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.total-card {
		padding: 1.35rem;
	}

	.eyebrow {
		margin: 0 0 0.55rem;
		font-size: 0.72rem;
		letter-spacing: 0.7px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.46);
	}

	.hero-value {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 1.6rem;
		color: var(--text);
	}

	.hero-note {
		margin: 0.4rem 0 0;
		font-size: 0.78rem;
		color: rgba(236, 234, 229, 0.52);
	}

	/* Mismo aviso de «falta una tasa» que las tarjetas de portafolios. */
	.fx-note {
		margin: 0 0 1.5rem;
		padding: 0.5rem 0.7rem;
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: rgba(212, 145, 42, 0.08);
		color: rgba(236, 234, 229, 0.75);
		font-size: 0.78rem;
		line-height: 1.4;
	}

	.panel-stack {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	@media (max-width: 768px) {
		.totals {
			grid-template-columns: 1fr;
		}
	}
</style>
