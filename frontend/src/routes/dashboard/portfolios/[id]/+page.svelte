<script lang="ts">
	import Modal from '$lib/ui/modal.svelte';
	import { PeriodReturns, PortfolioGrowth, hasTrailingReturns } from '$lib/features/dashboard';
	import { flash } from '$lib/shared/flash.svelte';
	import {
		PortfolioEditForm,
		PortfolioDetailHeader,
		PortfolioHeadline,
		PortfolioPositions,
		groupHoldings,
		splitClosedHoldings,
		computeTypeBreakdown
	} from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { params, data }: PageProps = $props();

	/* Lo editado se ve en la cabecera al pulsar; el refresco de fondo trae lo
	   guardado. Por versión, no por identidad: `$state` guarda un proxy. */
	let changes = $state<Record<string, string | undefined> | null>(null);
	let editHidden = $state(false);
	let version = 0;
	const portfolio = $derived(
		data.portfolio && changes ? { ...data.portfolio, ...changes } : data.portfolio
	);

	function applyChanges(next: Record<string, string | undefined>) {
		const mine = ++version;
		changes = next;
		editHidden = true;
		return () => {
			if (version === mine) changes = null;
		};
	}

	function closeEdit() {
		isEditing = false;
		editHidden = false;
	}
	const risks = $derived(data.risks);
	const growth = $derived(data.growth);

	let isEditing = $state(false);
	let submitError = $state('');
	const saved = flash(3000);

	// Group entries by ticker so the same asset held in multiple platforms
	// appears as a single row with aggregated quantity and cost basis.
	// Las vendidas enteras van aparte: siguen siendo del portafolio, pero ni
	// cuentan como activo ni entran en los totales ni en el reparto por clase.
	const positions = $derived(splitClosedHoldings(groupHoldings(portfolio?.holdings ?? [])));
	const holdings = $derived(positions.open);

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
	open={isEditing && !!data.portfolio}
	hidden={editHidden}
	title="Editar portafolio"
	onClose={closeEdit}
	size="lg"
>
	{#if data.portfolio}
		<PortfolioEditForm
			portfolio={data.portfolio}
			{risks}
			onApply={applyChanges}
			onCancel={closeEdit}
			onSaved={() => {
				// El error de un intento anterior tiene que irse con el acuse nuevo:
				// si no, la pantalla mostraba las dos alertas a la vez.
				submitError = '';
				closeEdit();
				saved.show('Portafolio actualizado correctamente.');
			}}
			onError={(msg) => {
				submitError = msg;
				editHidden = false;
			}}
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

{#if growth && hasTrailingReturns(growth.returns)}
	<!-- Lo mismo que la cabecera del panel, para este portafolio: su serie ya
	     viene en su moneda base, así que no hay conversión que avisar. -->
	<div class="returns">
		<PeriodReturns returns={growth.returns} currency={baseCurrency} />
	</div>
{/if}

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
	closed={positions.closed}
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

	.returns {
		padding: 1.75rem 0;
		border-bottom: 1px solid var(--border);
	}

	.growth {
		padding: 2rem 0;
		border-bottom: 1px solid var(--border);
	}
</style>
