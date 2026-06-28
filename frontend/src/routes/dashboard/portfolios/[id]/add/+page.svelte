<script lang="ts">
	import type { PageProps } from './$types';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$components/ui/page-header.svelte';
	import DatePicker from '$components/ui/date-picker.svelte';

	const { params, data, form }: PageProps = $props();

	interface FormData {
		platformId: string;
		assetId: string;
		quantity: string;
		purchasePrice: string;
		purchaseDate: string;
		totalValue: number;
		notes: string;
	}

	let formData: FormData = $state({
		platformId: '',
		assetId: '',
		quantity: '',
		purchasePrice: '',
		purchaseDate: new Date().toISOString().split('T')[0],
		totalValue: 0,
		notes: ''
	});

	let isSubmitting = $state(false);
	let submitSuccess = $state(false);
	let submitError = $derived(form?.success === false);

	const platforms = $derived(data?.platforms || []);

	// Asset combobox state — server-side search
	interface AssetSuggestion {
		id: string;
		ticker: string;
		name: string;
		assetType: string;
		exchange: string;
		currency: string;
		currentPrice: { value: string; currency: string } | null;
	}

	let assetSearch = $state('');
	let showSuggestions = $state(false);
	let suggestions = $state<AssetSuggestion[]>([]);
	let selectedAsset = $state<AssetSuggestion | null>(null);
	let isSearching = $state(false);
	let comboboxEl = $state<HTMLDivElement | null>(null);
	let debounceTimer: ReturnType<typeof setTimeout>;

	function selectAsset(asset: AssetSuggestion) {
		selectedAsset = asset;
		formData.assetId = asset.id;
		assetSearch = asset.ticker;
		showSuggestions = false;
	}

	async function fetchSuggestions(q: string) {
		isSearching = true;
		try {
			const url = q.trim()
				? `/api/assets?search=${encodeURIComponent(q.trim())}&limit=10`
				: `/api/assets?limit=10`;
			const res = await fetch(url);
			const json = await res.json();
			suggestions = json.success ? (json.data ?? []) : [];
		} catch {
			suggestions = [];
		} finally {
			isSearching = false;
		}
	}

	function onSearchInput() {
		selectedAsset = null;
		formData.assetId = '';
		showSuggestions = true;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => fetchSuggestions(assetSearch), 300);
	}

	function onSearchFocus() {
		showSuggestions = true;
		if (suggestions.length === 0 && !assetSearch) fetchSuggestions('');
	}

	function clickOutside(node: HTMLElement, handler: (e: MouseEvent) => void) {
		function listener(e: MouseEvent) {
			handler(e);
		}
		document.addEventListener('mousedown', listener);
		return {
			destroy() {
				document.removeEventListener('mousedown', listener);
			}
		};
	}

	$effect(() => {
		const qty = parseFloat(formData.quantity) || 0;
		const price = parseFloat(formData.purchasePrice) || 0;
		formData.totalValue = qty * price;
	});

	function handleCancel() {
		goto(resolve('/dashboard/portfolios/[id]', { id: params.id }));
	}

	function createNewPlatform() {
		goto(resolve('/dashboard/platforms/add'));
	}

	function formatCurrency(value: number): string {
		return new Intl.NumberFormat('es-CO', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 2
		}).format(value);
	}
</script>

<svelte:head>
	<title>Agregar Activo al Portafolio - FINEXIA</title>
	<meta name="description" content="Añade un nuevo activo a tu portafolio de inversiones" />
</svelte:head>

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
		action={`/dashboard/portfolios/${params.id}/add`}
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
		<section class="form-section">
			<h2 class="section-title">Plataforma de Inversión</h2>
			<div class="form-group">
				<label for="platformId" class="form-label"
					>Selecciona una Plataforma <span class="required">*</span></label
				>
				{#if platforms.length > 0}
					<select
						id="platformId"
						bind:value={formData.platformId}
						name="platformId"
						class="form-select"
						required
					>
						<option value="">-- Elige una plataforma --</option>
						{#each platforms as platform (platform.id)}
							<option value={platform.id}>{platform.name}</option>
						{/each}
					</select>
					<p class="field-hint">Selecciona dónde realizarás esta inversión</p>
				{:else}
					<div class="empty-platforms">
						<p class="empty-text">No tienes plataformas registradas</p>
						<button type="button" onclick={createNewPlatform} class="btn-link">
							<svg
								width="16"
								height="16"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<path d="M12 5v14M5 12h14" />
							</svg>
							Crear tu primera plataforma
						</button>
					</div>
				{/if}
			</div>
		</section>

		<!-- Asset Selection -->
		<section class="form-section">
			<h2 class="section-title">Seleccionar Activo</h2>
			<div class="form-group">
				<label for="asset-search" class="form-label"
					>Buscar Activo por Ticker o Nombre <span class="required">*</span></label
				>

				<input type="hidden" name="assetId" value={formData.assetId} />
				<input type="hidden" name="category" value={selectedAsset?.assetType ?? ''} />

				<div
					class="combobox"
					bind:this={comboboxEl}
					use:clickOutside={(e) => {
						if (comboboxEl && !comboboxEl.contains(e.target as Node)) showSuggestions = false;
					}}
				>
					<div class="combobox-input-wrap">
						<input
							id="asset-search"
							type="text"
							class="form-input combobox-input"
							placeholder="Escribe el ticker o nombre, ej: AAPL, Bitcoin…"
							autocomplete="off"
							bind:value={assetSearch}
							oninput={onSearchInput}
							onfocus={onSearchFocus}
						/>
						{#if isSearching}
							<span class="combobox-spinner"></span>
						{:else if selectedAsset}
							<button
								type="button"
								class="combobox-clear"
								aria-label="Limpiar selección"
								onclick={() => {
									selectedAsset = null;
									formData.assetId = '';
									assetSearch = '';
									showSuggestions = false;
								}}>✕</button
							>
						{/if}
					</div>

					{#if showSuggestions && suggestions.length > 0}
						<ul class="combobox-list" role="listbox">
							{#each suggestions as asset (asset.id)}
								<li
									role="option"
									aria-selected={asset.id === formData.assetId}
									class="combobox-option"
									class:selected={asset.id === formData.assetId}
									onmousedown={() => selectAsset(asset)}
								>
									<div class="option-left">
										<span class="option-ticker">{asset.ticker}</span>
										<span class="option-type">{asset.assetType}</span>
									</div>
									<div class="option-right">
										<span class="option-name">{asset.name}</span>
										{#if asset.exchange || asset.currency}
											<span class="option-meta">
												{[asset.exchange, asset.currency].filter(Boolean).join(' · ')}
											</span>
										{/if}
									</div>
									{#if asset.currentPrice}
										<span class="option-price">
											{new Intl.NumberFormat('en-US', {
												style: 'currency',
												currency:
													asset.currency && asset.currency !== 'XXX' ? asset.currency : 'USD',
												minimumFractionDigits: 2,
												maximumFractionDigits: 4
											}).format(parseFloat(asset.currentPrice.value))}
										</span>
									{/if}
								</li>
							{/each}
						</ul>
					{:else if showSuggestions && !isSearching && assetSearch.trim().length > 0}
						<div class="combobox-empty">
							No se encontró ningún activo con "<strong>{assetSearch}</strong>"
						</div>
					{/if}
				</div>

				<p class="field-hint">Búsqueda en tiempo real · escribe el ticker o nombre del activo</p>

				{#if selectedAsset}
					<div class="asset-preview">
						<div class="preview-item">
							<span class="preview-label">Ticker</span>
							<span class="preview-value">{selectedAsset.ticker}</span>
						</div>
						<div class="preview-item">
							<span class="preview-label">Nombre</span>
							<span class="preview-value">{selectedAsset.name}</span>
						</div>
						<div class="preview-item">
							<span class="preview-label">Tipo</span>
							<span class="preview-value">{selectedAsset.assetType}</span>
						</div>
						<div class="preview-item">
							<span class="preview-label">Exchange</span>
							<span class="preview-value">{selectedAsset.exchange}</span>
						</div>
						{#if selectedAsset.currentPrice}
							<div class="preview-item">
								<span class="preview-label">Precio de mercado</span>
								<span class="preview-value"
									>{formatCurrency(parseFloat(selectedAsset.currentPrice.value))}</span
								>
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</section>

		<!-- Purchase Details -->
		<section class="form-section">
			<h2 class="section-title">Detalles de Compra</h2>

			<div class="form-row">
				<div class="form-group">
					<label for="quantity" class="form-label">Cantidad <span class="required">*</span></label>
					<div class="input-addon">
						<input
							id="quantity"
							type="number"
							name="quantity"
							bind:value={formData.quantity}
							placeholder="1000"
							class="form-input"
							min="0"
							step="any"
							required
						/>
					</div>
					<p class="field-hint">Número de unidades</p>
				</div>

				<div class="form-group">
					<label for="purchasePrice" class="form-label"
						>Precio de Compra <span class="required">*</span></label
					>
					<div class="input-addon">
						<span class="addon-text">$</span>
						<input
							id="purchasePrice"
							type="number"
							name="purchasePrice"
							bind:value={formData.purchasePrice}
							placeholder="150.50"
							class="form-input"
							min="0"
							step="0.01"
							required
						/>
					</div>
					<p class="field-hint">Precio por unidad</p>
				</div>
			</div>

			<div class="form-row">
				<div class="form-group">
					<span class="form-label">Fecha de Compra</span>
					<DatePicker name="purchaseDate" bind:value={formData.purchaseDate} required />
				</div>

				<div class="form-group">
					<span class="form-label">Valor Total Invertido</span>
					<div class="value-display">
						<p class="total-value">{formatCurrency(formData.totalValue)}</p>
					</div>
				</div>
			</div>
		</section>

		<!-- Additional Notes -->
		<section class="form-section">
			<h2 class="section-title">Notas y Observaciones</h2>

			<div class="form-group">
				<label for="notes" class="form-label">Notas</label>
				<textarea
					id="notes"
					bind:value={formData.notes}
					name="notes"
					placeholder="Agrega observaciones, estrategia o detalles especiales sobre este activo..."
					class="form-textarea"
					rows="3"></textarea>
				<p class="field-hint">Notas personales sobre este activo (opcional)</p>
			</div>
		</section>

		<!-- Summary Card -->
		{#if selectedAsset && formData.quantity && formData.purchasePrice}
			<section class="summary-card">
				<h3 class="summary-title">Resumen de Inversión</h3>
				<div class="summary-items">
					<div class="summary-item">
						<span class="summary-label">Activo</span>
						<span class="summary-value">{selectedAsset.ticker} - {selectedAsset.name}</span>
					</div>
					<div class="summary-item">
						<span class="summary-label">Cantidad</span>
						<span class="summary-value">{parseFloat(formData.quantity).toLocaleString()}</span>
					</div>
					<div class="summary-item">
						<span class="summary-label">Precio Unitario</span>
						<span class="summary-value">{formatCurrency(parseFloat(formData.purchasePrice))}</span>
					</div>
					<div class="summary-item border-top">
						<span class="summary-label">Inversión Total</span>
						<span class="summary-value highlight">{formatCurrency(formData.totalValue)}</span>
					</div>
				</div>
			</section>
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
				{:else if submitSuccess}
					✓ Guardado
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

	.asset-preview {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		margin-top: 1rem;
		padding: 1rem;
		border-radius: 12px;
		background: var(--surface);
		border: 1px solid var(--border-strong);
	}

	.preview-item {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.preview-label {
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		color: rgba(236, 234, 229, 0.5);
		font-weight: 600;
	}

	.preview-value {
		font-size: 0.95rem;
		color: var(--amber);
		font-weight: 600;
	}

	.empty-text {
		margin: 0;
		font-size: 0.95rem;
		color: rgba(236, 234, 229, 0.7);
		font-weight: 500;
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

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
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

	.form-input,
	.form-select,
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

	.form-input::placeholder,
	.form-textarea::placeholder {
		color: rgba(236, 234, 229, 0.55);
	}

	.form-input:focus,
	.form-select:focus,
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

	.empty-platforms {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		padding: 1.5rem;
		border-radius: 10px;
		background: var(--surface);
		border: 1px dashed rgba(212, 145, 42, 0.2);
		text-align: center;
	}

	.empty-text {
		margin: 0;
		font-size: 0.95rem;
		color: rgba(236, 234, 229, 0.6);
	}

	.btn-link {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.65rem 1.2rem;
		border: none;
		background: var(--border-strong);
		color: var(--amber);
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		font-family: var(--font-body);
	}

	.btn-link:hover {
		background: rgba(212, 145, 42, 0.25);
		transform: translateX(2px);
	}

	.input-addon {
		position: relative;
		display: flex;
		align-items: center;
	}

	.addon-text {
		position: absolute;
		left: 1rem;
		font-size: 0.9rem;
		color: rgba(236, 234, 229, 0.5);
		font-weight: 600;
		pointer-events: none;
	}

	.input-addon .form-input {
		padding-left: 2.5rem;
	}

	.value-display {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		padding: 0.85rem 1rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 44px;
	}

	.total-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		margin: 0;
		font-size: 1.2rem;
		font-weight: 700;
		color: var(--amber);
		font-family: var(--font-body);
	}

	.summary-card {
		border: 2px solid rgba(212, 145, 42, 0.3);
		border-radius: 16px;
		background: rgba(212, 145, 42, 0.08);
		padding: 1.5rem;
		animation: slide-in 0.4s ease-out;
	}

	.summary-title {
		margin: 0 0 1.25rem;
		font-size: 1rem;
		font-weight: 700;
		color: var(--amber);
		font-family: var(--font-body);
	}

	.summary-items {
		display: grid;
		gap: 0.9rem;
	}

	.summary-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 0;
	}

	.summary-item.border-top {
		border-top: 1px solid rgba(212, 145, 42, 0.2);
		padding-top: 1rem;
		margin-top: 0.5rem;
	}

	.summary-label {
		font-size: 0.9rem;
		color: rgba(236, 234, 229, 0.65);
		font-weight: 500;
	}

	.summary-value {
		font-size: 0.95rem;
		color: var(--text);
		font-weight: 600;
	}

	.summary-value.highlight {
		color: var(--amber);
		font-size: 1.1rem;
	}

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

	@media (max-width: 1024px) {
		.asset-preview {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.asset-preview {
			grid-template-columns: 1fr;
		}

		.form-row {
			grid-template-columns: 1fr;
		}

		.form-actions {
			flex-direction: column-reverse;
		}

		.btn {
			width: 100%;
		}
	}

	/* Combobox */
	.combobox {
		position: relative;
	}

	.combobox-input-wrap {
		position: relative;
		display: flex;
		align-items: center;
	}

	.combobox-input {
		width: 100%;
		padding-right: 2.5rem;
	}

	.combobox-clear {
		position: absolute;
		right: 0.75rem;
		background: transparent;
		border: none;
		color: rgba(236, 234, 229, 0.4);
		font-size: 0.85rem;
		cursor: pointer;
		padding: 0.2rem 0.3rem;
		border-radius: 4px;
		transition: color 0.2s ease;
		line-height: 1;
	}

	.combobox-clear:hover {
		color: var(--text);
	}

	.combobox-list {
		position: absolute;
		top: calc(100% + 4px);
		left: 0;
		right: 0;
		z-index: 50;
		list-style: none;
		margin: 0;
		padding: 0.35rem 0;
		background: var(--surface);
		border: 1.5px solid rgba(212, 145, 42, 0.35);
		border-radius: 10px;
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
		max-height: 280px;
		overflow-y: auto;
	}

	.combobox-option {
		display: grid;
		grid-template-columns: 100px 1fr auto;
		align-items: center;
		gap: 0.75rem;
		padding: 0.65rem 1rem;
		cursor: pointer;
		transition: background 0.15s ease;
	}

	.combobox-option:hover,
	.combobox-option.selected {
		background: rgba(212, 145, 42, 0.1);
	}

	.option-left {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		min-width: 0;
	}

	.option-ticker {
		font-family: var(--font-mono);
		font-weight: 700;
		font-size: 0.92rem;
		color: var(--amber);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.option-type {
		font-size: 0.68rem;
		color: rgba(236, 234, 229, 0.4);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.option-right {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		min-width: 0;
	}

	.option-name {
		font-size: 0.88rem;
		color: var(--text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.option-meta {
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.4);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.option-price {
		font-family: var(--font-mono);
		font-size: 0.82rem;
		color: rgba(212, 145, 42, 0.7);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		flex-shrink: 0;
	}

	.combobox-empty {
		padding: 0.75rem 1rem;
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.5);
	}
</style>
