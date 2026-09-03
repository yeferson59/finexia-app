<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import { todayLocalDateString } from '$lib/shared/format/date';
	import { FALLBACK_CURRENCY } from '$lib/shared/currency';
	import type { Asset, Platform } from '$lib/api/types';
	import AssetCombobox from './asset-combobox.svelte';
	import PortfolioEntryPlatformField from './portfolio-entry-platform-field.svelte';
	import PortfolioEntryPurchaseFields from './portfolio-entry-purchase-fields.svelte';
	import AssetPreview from './asset-preview.svelte';
	import PortfolioEntrySummary from './portfolio-entry-summary.svelte';

	let {
		portfolioId,
		platforms,
		submitError = false
	}: {
		portfolioId: string;
		platforms: Platform[];
		submitError?: boolean;
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
	 * Viven aquí porque el resumen y el total invertido también necesitan saber
	 * en qué moneda acabó el coste.
	 */
	let costCurrency = $state('');
	let fxRate = $state('');

	let isSubmitting = $state(false);

	// El total es lo que la cuenta pagó, no lo que la operación cotizó: son el
	// mismo número salvo que el bróker haya convertido, y cuando convierte es el
	// segundo el que no se puede comparar con nada más de la pantalla.
	const totalValue = $derived(
		(parseFloat(quantity) || 0) * (parseFloat(purchasePrice) || 0) * (parseFloat(fxRate) || 1)
	);

	function handleCancel() {
		goto(resolve('/dashboard/portfolios/[id]', { id: portfolioId }));
	}

	function formatCurrency(value: number, code: string = costCurrency || currency): string {
		return new Intl.NumberFormat('es-CO', {
			style: 'currency',
			currency: code,
			minimumFractionDigits: 2
		}).format(value);
	}
</script>

<button class="back-button" onclick={handleCancel} aria-label="Volver al portafolio">
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
	Volver
</button>

<PageHeader
	title="Agregar Activo al Portafolio"
	subtitle="Registra un nuevo activo en tu cartera de inversiones"
/>

<div class="form-container">
	<form
		method="POST"
		action={`/dashboard/portfolios/${portfolioId}/add`}
		class="portfolio-form"
		use:enhance={() => {
			isSubmitting = true;
			return async ({ update }) => {
				await update();
				isSubmitting = false;
			};
		}}
	>
		<!-- Platform Selection -->
		<PortfolioEntryPlatformField {platforms} bind:selected={platformId} />

		<!-- Asset Selection -->
		<section class="form-section">
			<h2 class="section-title">Seleccionar Activo</h2>
			<div class="form-group">
				<label for="asset-search" class="form-label"
					>Buscar Activo por Ticker o Nombre <span class="required">*</span></label
				>

				<input type="hidden" name="assetId" value={selectedAsset?.id ?? ''} />

				<AssetCombobox bind:selected={selectedAsset} bind:search={assetSearch} />

				<p class="field-hint">Búsqueda en tiempo real · escribe el ticker o nombre del activo</p>

				{#if selectedAsset}
					<AssetPreview asset={selectedAsset} {formatCurrency} />
				{/if}
			</div>
		</section>

		<!-- Purchase Details -->
		<PortfolioEntryPurchaseFields
			asset={selectedAsset}
			bind:quantity
			bind:purchasePrice
			bind:purchaseDate
			bind:currency
			bind:costCurrency
			bind:fxRate
			{formatCurrency}
		/>

		<!-- Additional Notes -->
		<section class="form-section">
			<h2 class="section-title">Notas y Observaciones</h2>

			<div class="form-group">
				<label for="notes" class="form-label">Notas</label>
				<textarea
					id="notes"
					bind:value={notes}
					name="notes"
					placeholder="Agrega observaciones, estrategia o detalles especiales sobre este activo..."
					class="form-textarea"
					rows="3"></textarea>
				<p class="field-hint">Notas personales sobre este activo (opcional)</p>
			</div>
		</section>

		<!-- Summary Card -->
		{#if selectedAsset && quantity && purchasePrice}
			<PortfolioEntrySummary
				asset={selectedAsset}
				{quantity}
				{purchasePrice}
				{currency}
				costCurrency={costCurrency || currency}
				{totalValue}
				{formatCurrency}
			/>
		{/if}

		<!-- Error feedback -->
		{#if submitError}
			<div class="form-error">
				No se pudo registrar el activo. Verifica que todos los campos sean correctos e intenta de
				nuevo.
			</div>
		{/if}

		<!-- Action Buttons -->
		<div class="form-actions">
			<button type="button" onclick={handleCancel} class="btn btn-secondary">Cancelar</button>
			<button type="submit" disabled={isSubmitting} class="btn btn-primary">
				{#if isSubmitting}
					<span class="spinner"></span>
					Guardando...
				{:else}
					Agregar Activo
				{/if}
			</button>
		</div>
	</form>
</div>

<style>
	.back-button {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 1.5rem;
		padding: 0.65rem 1rem;
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

	.back-button:hover {
		background: var(--border);
		border-color: var(--amber);
	}

	.form-container {
		max-width: 1000px;
	}

	.portfolio-form {
		display: grid;
		gap: 2rem;
		animation: fade-in 0.4s ease-out;
	}

	.form-section {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		padding: 1.75rem;
	}

	.section-title {
		margin: 0 0 1.5rem;
		font-size: 1.15rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
		margin-bottom: 1.35rem;
	}

	.form-group:last-child {
		margin-bottom: 0;
	}

	.form-label {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text);
		letter-spacing: 0.3px;
	}

	.required {
		color: var(--red);
	}

	.form-textarea {
		padding: 0.85rem 1rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		color: var(--text);
		font-size: 0.95rem;
		font-family: var(--font-body);
		transition: all 0.3s ease;
	}

	.form-textarea::placeholder {
		color: rgba(236, 234, 229, 0.55);
	}

	.form-textarea:focus {
		outline: none;
		border-color: var(--amber);
		background: rgba(255, 255, 255, 0.022);
		box-shadow: 0 0 0 3px var(--border);
	}

	.form-textarea {
		resize: vertical;
		min-height: 100px;
	}

	.field-hint {
		margin: 0.4rem 0 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.4);
		font-style: italic;
	}

	/* La página original declaraba `.empty-text` dos veces; se fusionan aquí con
	   el resultado efectivo de la cascada (color de la segunda, peso de la primera). */

	.form-error {
		padding: 1rem 1.25rem;
		border-radius: 10px;
		background: rgba(224, 90, 90, 0.1);
		border: 1px solid rgba(224, 90, 90, 0.3);
		color: rgba(224, 90, 90, 0.9);
		font-size: 0.9rem;
		font-weight: 500;
	}

	.form-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 2rem;
	}

	.btn {
		padding: 0.85rem 1.5rem;
		border: none;
		border-radius: 10px;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		letter-spacing: 0.3px;
	}

	.btn-primary {
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
	}

	.btn-primary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-secondary {
		background: transparent;
		color: var(--text);
		border: 1.5px solid rgba(212, 145, 42, 0.25);
	}

	.btn-secondary:hover {
		border-color: var(--amber);
		background: var(--border);
		color: var(--amber);
	}

	.spinner {
		display: inline-block;
		width: 14px;
		height: 14px;
		border: 2px solid rgba(255, 255, 255, 0.022);
		border-top-color: #0d0800;
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes fade-in {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes slide-in {
		from {
			opacity: 0;
			transform: translateX(-10px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 768px) {
		.form-actions {
			flex-direction: column-reverse;
		}

		.btn {
			width: 100%;
		}
	}
</style>
