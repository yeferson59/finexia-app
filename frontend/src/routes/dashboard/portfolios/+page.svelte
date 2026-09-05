<script lang="ts">
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
	import {
		PortfolioList,
		portfolioBarScale,
		portfolioTotals,
		toPortfolioRows
	} from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	/** Moneda de la cuenta: en ella pide el layout los resúmenes ya convertidos. */
	const displayCurrency = $derived(data.currency ?? FALLBACK_CURRENCY);
	const rows = $derived(toPortfolioRows(data.portfolios ?? [], displayCurrency));

	const totals = $derived(portfolioTotals(rows));
	// Escala y totales salen de la lista entera, no de la hoja: si se calcularan
	// por hoja, la primera barra de cada una llegaría al final del carril y el
	// pie diría un total distinto en cada página.
	const scale = $derived(portfolioBarScale(rows));

	const PER_PAGE = 12;
	let page = $state(1);
	const pagedRows = $derived(rows.slice((page - 1) * PER_PAGE, page * PER_PAGE));
</script>

<svelte:head>
	<title>Portafolios - FINEXIA</title>
	<meta name="description" content="Gestión de múltiples portafolios de inversión" />
</svelte:head>

<PageHeader
	title="Portafolios"
	subtitle="Cómo tienes agrupado tu dinero, y cómo le va a cada grupo."
>
	{#snippet actions()}
		<a class="create" href={resolve('/dashboard/portfolios/add')}>Crear portafolio</a>
	{/snippet}
</PageHeader>

{#if data.success === false}
	<!-- El listado vacío y «no pudimos traerlo» se veían igual: una página en
	     blanco. Quien no tiene portafolios necesita una invitación a crear uno,
	     y quien sí los tiene necesita saber que el fallo no es suyo. -->
	<p class="failure">
		No pudimos cargar tus portafolios. Vuelve a intentarlo en un momento; tus datos siguen ahí.
	</p>
{:else}
	<PortfolioList rows={pagedRows} {totals} {scale} {displayCurrency} />

	{#if totals.excluded > 0}
		<p class="fx">
			{totals.excluded === 1 ? 'Un portafolio queda' : `${totals.excluded} portafolios quedan`} fuera
			del total: no hay tasa para pasarlo{totals.excluded === 1 ? '' : 's'} a {displayCurrency}. Su
			fila enseña el importe en su propia moneda.
		</p>
	{/if}

	<Pagination bind:page total={rows.length} perPage={PER_PAGE} label="portafolios" />
{/if}

<style>
	/*
	 * La acción principal, en el tono de la marca pero sin el salto y el halo
	 * que tenía: es un enlace a un formulario, no un acontecimiento. Y es un
	 * enlace de verdad, así que se puede abrir en otra pestaña.
	 */
	.create {
		display: inline-flex;
		align-items: center;
		padding: 0.6rem 1.15rem;
		border-radius: 9px;
		background: var(--amber);
		color: #0d0800;
		font-size: 0.88rem;
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
		transition: background 0.2s ease;
	}

	.create:hover {
		background: var(--amber-light);
	}

	.failure {
		max-width: 56ch;
		margin: 0;
		padding: 2.5rem 0;
		font-size: 0.9rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	/* Mismo aviso de «falta una tasa» que el panel y la vista de activos:
	   filete ámbar y prosa, no una caja de alerta. */
	.fx {
		max-width: 62ch;
		margin: 1.25rem 0 0;
		padding-left: 0.75rem;
		border-left: 2px solid rgba(212, 145, 42, 0.45);
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	@media (prefers-reduced-motion: reduce) {
		.create {
			transition: none;
		}
	}
</style>
