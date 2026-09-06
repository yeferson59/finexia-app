<script lang="ts">
	/*
	 * Añadir un activo a un portafolio.
	 *
	 * Eran cuatro tarjetas con sombra —plataforma, activo, compra, notas— y,
	 * en cuanto elegías algo, dos paneles más: una rejilla con los cinco datos
	 * del activo y una tarjeta ámbar de «Resumen de Inversión». Seis superficies
	 * para una tarea, y el total escrito dos veces. Ahora son tres bloques en el
	 * carril de configuración y un único total al cierre del segundo.
	 *
	 * El orden también cambia. Antes empezaba por la plataforma, que es un
	 * desplegable de dos o tres cosas; ahora empieza por el activo, que es lo que
	 * identifica la operación y además siembra la moneda del precio.
	 */
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Button from '$lib/ui/button.svelte';
	import PageHeader from '$lib/ui/page-header.svelte';
	import { todayLocalDateString } from '$lib/shared/format/date';
	import { formatCurrency as formatMoney } from '$lib/shared/format/money';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
	import type { Asset, Platform } from '$lib/api/types';
	import AssetCombobox from './asset-combobox.svelte';
	import AssetPreview from './asset-preview.svelte';
	import PortfolioFormSection from './portfolio-form-section.svelte';
	import PortfolioEntryPlatformField from './portfolio-entry-platform-field.svelte';
	import PortfolioEntryPurchaseFields from './portfolio-entry-purchase-fields.svelte';
	import PortfolioEntryTotal from './portfolio-entry-total.svelte';

	let {
		portfolioId,
		platforms,
		submitError = false,
		submitErrorDetail = ''
	}: {
		portfolioId: string;
		platforms: Platform[];
		submitError?: boolean;
		/**
		 * Lo que dijo el backend, cuando dijo algo.
		 *
		 * Un 400 de este endpoint nombra el campo que no puede ser cierto —«USD
		 * no se convierte en sí misma a 1.0638»— y el mensaje genérico de abajo
		 * lo tapaba, dejando al usuario reintentando contra el mismo rechazo.
		 */
		submitErrorDetail?: string;
	} = $props();

	let platformId = $state('');
	let quantity = $state('');
	let purchasePrice = $state('');
	let purchaseDate = $state(todayLocalDateString());
	let notes = $state('');

	let assetSearch = $state('');
	let selectedAsset = $state<Asset | null>(null);

	/**
	 * Moneda en la que se pagó la compra. El precio que el usuario copia de su
	 * bróker viene en la moneda de cotización del activo, así que esa es la
	 * semilla; antes se enviaba USD fijo y una acción danesa entraba con su
	 * precio en DKK etiquetado como dólares, lo que inflaba la pérdida al
	 * comparar un coste sin convertir contra un valor de mercado convertido.
	 *
	 * Sigue siendo editable porque las dos monedas pueden diferir de verdad:
	 * un bróker puede liquidar en USD una compra cotizada en DKK. Derivado
	 * reasignable: se resiembra al cambiar de activo y, mientras tanto, manda lo
	 * que haya elegido el usuario en el selector.
	 *
	 * Se exige el código ISO de tres letras: `Intl.NumberFormat` lanza con
	 * cualquier otra cosa, y `assets.currency` es un CHAR(3) que un activo
	 * aportado por un usuario puede traer relleno de espacios.
	 */
	const assetCurrency = $derived(selectedAsset?.currency?.trim().toUpperCase() ?? '');
	let currency: string = $derived(
		/^[A-Z]{3}$/.test(assetCurrency) ? assetCurrency : FALLBACK_CURRENCY
	);

	/**
	 * Moneda en la que liquidó la cuenta y tasa a la que el bróker convirtió.
	 *
	 * Las gestiona el campo de compra: mientras las dos monedas coincidan
	 * mantiene `costCurrency` igual a `currency` y la tasa vacía, que es tasa 1.
	 * Viven aquí porque el total también necesita saber en qué moneda acabó el
	 * coste.
	 */
	let costCurrency = $state('');
	let fxRate = $state('');

	let isSubmitting = $state(false);

	const units = $derived(parseFloat(quantity) || 0);
	const unitPrice = $derived(parseFloat(purchasePrice) || 0);
	/*
	 * Hubo conversión cuando las dos monedas difieren de verdad, no cuando está
	 * marcado el interruptor: con el interruptor puesto y las dos monedas
	 * iguales, el formulario manda tasa 1 —el backend rechaza cualquier otra—,
	 * así que el total tiene que enseñar la misma cuenta que se va a guardar.
	 */
	const converted = $derived(!!costCurrency && costCurrency !== currency);
	const rate = $derived(converted ? parseFloat(fxRate) || 0 : 1);

	/*
	 * El mismo formateador que el resto del panel. Esta pantalla componía los
	 * importes con un `Intl.NumberFormat` en `es-CO` para toda moneda, así que un
	 * precio en dólares salía «US$ 150,50» donde las demás dicen «$150.50».
	 *
	 * Y con el escape para el precio por unidad por debajo del céntimo: una
	 * cripto a 0,00004182 se escribía «$0.00» en el mismo renglón que su total.
	 */
	function formatCurrency(value: number, code: string = costCurrency || currency): string {
		const maxDigits = value !== 0 && Math.abs(value) < 0.01 ? 8 : undefined;

		return formatMoney(value, code || FALLBACK_CURRENCY, maxDigits);
	}
</script>

<!-- Sin filete: lo pone el primer bloque. Y sin el botón «Volver» que flotaba
     encima sin alinear con nada: al portafolio se vuelve por «Cancelar». -->
<PageHeader
	title="Añadir activo"
	subtitle="Anota un activo que ya tienes, con lo que te costó."
	divider={false}
/>

<form
	method="POST"
	action={`/dashboard/portfolios/${portfolioId}/add`}
	use:enhance={() => {
		isSubmitting = true;
		return async ({ update }) => {
			await update();
			isSubmitting = false;
		};
	}}
>
	<PortfolioFormSection
		title="Qué compraste"
		description="Busca por el ticker o por el nombre. Si el activo no está en el catálogo, el propio buscador te deja crearlo y seguir sin salir de aquí."
	>
		<div class="field">
			<label for="asset-search">Activo</label>

			<input type="hidden" name="assetId" value={selectedAsset?.id ?? ''} />

			<AssetCombobox bind:selected={selectedAsset} bind:search={assetSearch} />

			{#if selectedAsset}
				<AssetPreview asset={selectedAsset} {formatCurrency} />
			{/if}
		</div>
	</PortfolioFormSection>

	<PortfolioFormSection
		title="Cuánto pagaste"
		description="Copia las cifras de la confirmación de tu bróker, no las de hoy: lo que te costó la posición es lo que pagaste el día que compraste."
	>
		<PortfolioEntryPurchaseFields
			asset={selectedAsset}
			bind:quantity
			bind:purchasePrice
			bind:purchaseDate
			bind:currency
			bind:costCurrency
			bind:fxRate
		/>

		<PortfolioEntryTotal
			{units}
			{unitPrice}
			{currency}
			{costCurrency}
			{rate}
			{converted}
			{formatCurrency}
		/>
	</PortfolioFormSection>

	<PortfolioFormSection
		title="Dónde lo tienes"
		description="Hay una posición por plataforma: el mismo ticker comprado en dos brókers son dos posiciones, cada una con su propio coste."
	>
		<PortfolioEntryPlatformField {platforms} bind:selected={platformId} />

		<div class="field">
			<label for="notes">Notas <span class="optional">(opcional)</span></label>
			<textarea
				id="notes"
				name="notes"
				bind:value={notes}
				placeholder="Por qué compraste, o cualquier cosa que quieras recordar de esta posición."
				rows="3"></textarea>
		</div>
	</PortfolioFormSection>

	<div class="close">
		{#if submitError}
			<p class="feedback error">
				{#if submitErrorDetail}
					No pudimos añadir el activo: {submitErrorDetail}
				{:else}
					No pudimos añadir el activo. Revisa las cifras y vuelve a intentarlo.
				{/if}
			</p>
		{/if}

		<div class="actions">
			<Button type="submit" loading={isSubmitting}>
				{isSubmitting ? 'Añadiendo…' : 'Añadir activo'}
			</Button>
			<a class="cancel" href={resolve('/dashboard/portfolios/[id]', { id: portfolioId })}>
				Cancelar
			</a>
		</div>
	</div>
</form>

<style>
	.close {
		padding-top: 2.25rem;
		border-top: 1px solid var(--border-strong);
	}

	/* Prosa con un filete rojo, no una caja de alerta: el idioma que ya hablan
	   configuración, notificaciones y las altas de plataforma y portafolio. */
	.feedback {
		max-width: 62ch;
		margin: 0 0 1.25rem;
		padding-left: 0.75rem;
		border-left: 2px solid;
		font-size: 0.83rem;
		line-height: 1.5;
	}

	.feedback.error {
		border-color: var(--red);
		color: var(--red);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 1.25rem;
	}

	/* Sin el halo ámbar de `ui/button`: el ámbar de esta pantalla es el total. */
	.actions :global(.btn-primary) {
		box-shadow: none;
	}

	.cancel {
		font-size: 0.85rem;
		color: var(--text-muted);
		text-decoration: none;
		transition: color 0.2s ease;
	}

	.cancel:hover {
		color: var(--text);
	}

	@media (prefers-reduced-motion: reduce) {
		.cancel {
			transition: none;
		}
	}
</style>
