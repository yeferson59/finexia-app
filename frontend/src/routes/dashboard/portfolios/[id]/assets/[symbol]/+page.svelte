<script lang="ts">
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency as formatMoney } from '$lib/shared/format/money';
	import {
		AssetPositionHeader,
		AssetPositionHeadline,
		AssetTransactionHistory,
		AssetDeletePosition,
		computePosition
	} from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { params, data, form }: PageProps = $props();

	const entries = $derived(data.entries);
	const transactions = $derived(data.transactions ?? []);
	const txnMeta = $derived(data.txnMeta);
	const portfolioName = $derived(data.portfolioName ?? '');

	// Métricas de la posición a partir de los agregados de las entradas
	// (exactas con independencia de la paginación de transacciones).
	const position = $derived(
		computePosition(entries, data.portfolioTotalValue, data.baseCurrency ?? 'USD')
	);

	/**
	 * El formulario de alta lo abre la cabecera, que es donde vive la acción
	 * principal de la ficha; el historial es quien lo aloja.
	 */
	let showAddForm = $state(false);

	/**
	 * En esta ficha conviven tres monedas —la de coste, la de cotización del
	 * activo y la base del portafolio— y cada cifra tiene que llevar la suya:
	 * pintar un precio en EUR con el símbolo del dólar era el origen de la
	 * confusión que este cambio corrige.
	 *
	 * El formateador es el compartido, que ya elige el locale de cada moneda.
	 * El de aquí estaba escrito a mano con `es-CO` para todas, así que un
	 * importe en dólares salía «US$ 45.035,10» mientras el resto de la
	 * aplicación decía «$45,035.10» del mismo número.
	 */
	function formatAmount(value: number, currency: string): string {
		const code = currency || 'USD';
		// Un precio por unidad por debajo del céntimo —el interés de una cuenta
		// a 0,0021 por dólar— se escribía «$0.00» en la misma fila que su total
		// de $19.95, y la fila dejaba de cuadrar.
		const maxDigits = value !== 0 && Math.abs(value) < 0.01 ? 6 : undefined;

		return privacy.money(formatMoney(value, code, maxDigits));
	}

	/** Importes por unidad de la posición: van en la moneda de coste. */
	function formatCurrency(value: number): string {
		return formatAmount(value, position?.costCurrency || 'USD');
	}
</script>

<svelte:head>
	<title>{params.symbol} - Portfolio - FINEXIA</title>
	<meta name="description" content="Detalles de la posición {params.symbol}" />
</svelte:head>

{#if !position}
	<p class="missing">
		No hay ninguna posición de <strong>{params.symbol}</strong> en este portafolio. Puede que la hayas
		eliminado, o que esté en otro.
	</p>
{:else}
	<AssetPositionHeader
		{position}
		portfolioId={params.id}
		{portfolioName}
		onAddTransaction={() => (showAddForm = true)}
	/>

	{#if !position.fxConverted}
		<!-- Filete y prosa, como el resto de avisos del panel: era una caja roja
		     que pesaba más que la cifra que matizaba. -->
		<p class="notice fx">
			Sin tasa {position.currency} → {position.baseCurrency}: los totales van sin convertir, así que
			no son comparables con el resto del portafolio.
		</p>
	{/if}

	<AssetPositionHeadline {position} {portfolioName} />

	<AssetTransactionHistory
		portfolioId={params.id}
		symbol={params.symbol}
		bind:showAddForm
		{entries}
		{transactions}
		{txnMeta}
		marketPrice={position.marketPrice}
		{form}
		{formatCurrency}
		{formatAmount}
	/>

	<AssetDeletePosition
		portfolioId={params.id}
		{entries}
		transactionsCount={txnMeta.total}
		{formatAmount}
	/>
{/if}

<style>
	.missing {
		max-width: 56ch;
		margin: 0;
		padding: 3rem 0;
		font-size: 0.9rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	.notice {
		max-width: 68ch;
		margin: 0 0 1.5rem;
		padding-left: 0.75rem;
		border-left: 2px solid;
		font-size: 0.85rem;
		line-height: 1.5;
	}

	.notice.fx {
		border-color: rgba(212, 145, 42, 0.45);
		color: var(--text-muted);
	}
</style>
