<script lang="ts">
	import Modal from '$lib/ui/modal.svelte';
	import { PortfolioGrowth } from '$lib/features/dashboard';
	import { flash } from '$lib/shared/flash.svelte';
	import {
		PortfolioEditForm,
		PortfolioDetailHeader,
		PortfolioHeadline,
		PortfolioPositions,
		groupHoldings,
		computeTypeBreakdown
	} from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { params, data }: PageProps = $props();

	const portfolio = $derived(data.portfolio);
	const risks = $derived(data.risks);
	const growth = $derived(data.growth);

	let isEditing = $state(false);
	let submitError = $state('');
	const saved = flash(3000);

	// Group entries by ticker so the same asset held in multiple platforms
	// appears as a single row with aggregated quantity and cost basis.
	const holdings = $derived(groupHoldings(portfolio?.holdings ?? []));

	const totalValue = $derived(holdings.reduce((sum, h) => sum + h.value, 0));
	const totalCost = $derived(holdings.reduce((sum, h) => sum + h.costBasis, 0));
	const baseCurrency = $derived(portfolio?.baseCurrency?.trim() || 'USD');

	// Posiciones que el backend no pudo convertir por falta de tasa: sus
	// importes están en su moneda nativa, así que los totales de arriba mezclan
	// monedas y hay que decirlo en vez de presentarlos como comparables.
	const unconverted = $derived(holdings.filter((h) => !h.fxConverted));

	const typeBreakdown = $derived(computeTypeBreakdown(holdings));

	function startEditing() {
		submitError = '';
		isEditing = true;
	}
</script>

<svelte:head>
	<title>{portfolio?.name ?? 'Portafolio'} - FINEXIA</title>
	<meta name="description" content="Detalle de posiciones y asignación de portafolio" />
</svelte:head>

<PortfolioDetailHeader
	name={portfolio?.name ?? 'Portafolio'}
	description={portfolio?.description}
	riskName={portfolio?.riskName}
	holdingsCount={holdings.length}
	portfolioId={params.id}
	onEdit={startEditing}
/>

{#if saved.text}
	<p class="notice ok">{saved.text}</p>
{/if}

{#if submitError}
	<p class="notice bad">{submitError}</p>
{/if}

<Modal
	open={isEditing && !!portfolio}
	title="Editar portafolio"
	onClose={() => (isEditing = false)}
	size="lg"
>
	{#if portfolio}
		<PortfolioEditForm
			{portfolio}
			{risks}
			onCancel={() => (isEditing = false)}
			onSaved={() => {
				// El error de un intento anterior tiene que irse con el acuse nuevo:
				// si no, la pantalla mostraba las dos alertas a la vez.
				submitError = '';
				isEditing = false;
				saved.show('Portafolio actualizado correctamente.');
			}}
			onError={(msg) => (submitError = msg)}
		/>
	{/if}
</Modal>

{#if unconverted.length > 0}
	<p class="notice fx">
		Sin tasa de cambio para {unconverted.map((h) => `${h.symbol} (${h.currency})`).join(', ')}: esos
		importes van sin convertir a {baseCurrency}, así que los totales de abajo mezclan monedas.
	</p>
{/if}

<PortfolioHeadline value={totalValue} cost={totalCost} {baseCurrency} />

{#if growth}
	<section class="growth" aria-label="Crecimiento del portafolio">
		<!-- `bare`, como en el panel: ya no hay tarjetas alrededor con las que
		     tuviera que competir. Y sin formateador propio, para que escriba los
		     importes igual que el resto de la aplicación. -->
		<PortfolioGrowth bare data={growth.points} summary={growth.summary} />
	</section>
{/if}

<PortfolioPositions
	{holdings}
	{typeBreakdown}
	topTransaction={data.topTransaction}
	portfolioId={params.id}
	{baseCurrency}
/>

<style>
	/*
	 * Los avisos: filete de color y prosa, como en el resto del panel. Eran
	 * cajas con borde y fondo tintado que competían con la cifra de al lado.
	 */
	.notice {
		max-width: 68ch;
		margin: 0 0 1.5rem;
		padding-left: 0.75rem;
		border-left: 2px solid;
		font-size: 0.85rem;
		line-height: 1.5;
	}

	.notice.ok {
		border-color: var(--green);
		color: var(--green);
	}

	.notice.bad {
		border-color: var(--red);
		color: var(--red);
	}

	.notice.fx {
		border-color: rgba(212, 145, 42, 0.45);
		color: var(--text-muted);
	}

	.growth {
		padding: 2rem 0;
		border-bottom: 1px solid var(--border);
	}
</style>
