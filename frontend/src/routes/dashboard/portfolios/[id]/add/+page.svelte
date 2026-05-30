<script lang="ts">
	import type { PageProps } from './$types';
	import { goto } from '$app/navigation';

	interface Asset {
		id: string;
		ticker: string;
		name: string;
		assetType: string;
		exchange: string;
		currency: string;
		createdAt: string;
		updatedAt: string;
	}

	interface Platform {
		id: string;
		name: string;
	}

	interface PageData {
		platforms?: Platform[];
		assets?: Asset[];
	}

	const { params, data }: PageProps & { data: PageData } = $props();

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

	const platforms = $derived(data?.platforms || []);
	const assets = $derived(data?.assets || []);
	const selectedAsset = $derived(assets.find((a) => a.id === formData.assetId));
	const assetTypeIcon: Record<string, string> = {
		stock: '📈',
		crypto: '₿',
		bond: '📊',
		etf: '🎯',
		commodity: '⛏️',
		fund: '💼',
		option: '⚙️',
		forex: '💱'
	};

	$effect(() => {
		const qty = parseFloat(formData.quantity) || 0;
		const price = parseFloat(formData.purchasePrice) || 0;
		formData.totalValue = qty * price;
	});

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!formData.assetId || !formData.quantity || !formData.purchasePrice) {
			alert('Por favor completa todos los campos requeridos');
			return;
		}

		if (!formData.platformId) {
			alert('Por favor selecciona una plataforma');
			return;
		}

		isSubmitting = true;
		try {
			await new Promise((resolve) => setTimeout(resolve, 1000));
			submitSuccess = true;
			setTimeout(() => {
				goto(`/dashboard/portafolios/${params.id}`);
			}, 1500);
		} catch (error) {
			console.error('Error:', error);
		} finally {
			isSubmitting = false;
		}
	}

	function handleCancel() {
		goto(`/dashboard/portafolios/${params.id}`);
	}

	function createNewPlatform() {
		goto('/dashboard/platforms/add?redirect=/dashboard/portafolios/' + params.id + '/add');
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

<header class="page-header">
	<h1 class="page-title">Agregar Activo al Portafolio</h1>
	<p class="page-subtitle">Registra un nuevo activo en tu cartera de inversiones</p>
</header>

<div class="form-container">
<form onsubmit={handleSubmit} class="portfolio-form">
		<!-- Platform Selection -->
		<section class="form-section">
			<h2 class="section-title">Plataforma de Inversión</h2>
			<div class="form-group">
				<label for="platformId" class="form-label"
					>Selecciona una Plataforma <span class="required">*</span></label
				>
				{#if platforms.length > 0}
					<select id="platformId" bind:value={formData.platformId} class="form-select" required>
						<option value="">-- Elige una plataforma --</option>
						{#each platforms as platform}
							<option value={platform.id}>{platform.name}</option>
						{/each}
					</select>
					<p class="field-hint">Selecciona dónde realizarás esta inversión</p>
				{:else}
					<div class="empty-platforms">
						<p class="empty-text">No tienes plataformas registradas</p>
						<button type="button" onclick={createNewPlatform} class="btn-link">
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
				<label for="assetId" class="form-label"
					>Elige un Activo <span class="required">*</span></label
				>
				{#if assets.length > 0}
					<select id="assetId" bind:value={formData.assetId} class="form-select" required>
						<option value="">-- Selecciona un activo --</option>
						{#each assets as asset}
							<option value={asset.id}>
								{asset.ticker} - {asset.name} ({asset.assetType})
							</option>
						{/each}
					</select>
					<p class="field-hint">Selecciona de la lista de activos disponibles</p>

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
						</div>
					{/if}
				{:else}
					<div class="empty-assets">
						<p class="empty-text">No hay activos disponibles en el sistema</p>
						<p class="empty-hint">Contacta al administrador para agregar nuevos activos</p>
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
							bind:value={formData.quantity}
							placeholder="1000"
							class="form-input"
							min="0"
							step="0.01"
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
					<label for="purchaseDate" class="form-label">Fecha de Compra</label>
					<input
						id="purchaseDate"
						type="date"
						bind:value={formData.purchaseDate}
						class="form-input"
					/>
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
					placeholder="Agrega observaciones, estrategia o detalles especiales sobre este activo..."
					class="form-textarea"
					rows="3"
				></textarea>
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
		border: 1.5px solid rgba(212, 175, 55, 0.25);
		border-radius: 8px;
		color: #d4af37;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.back-button:hover {
		background: rgba(212, 175, 55, 0.1);
		border-color: #d4af37;
	}

	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		color: #d4af37;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.page-subtitle {
		margin: 0;
		color: rgba(224, 224, 224, 0.62);
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
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		padding: 1.75rem;
	}

	.section-title {
		margin: 0 0 1.5rem;
		font-size: 1.15rem;
		font-weight: 700;
		color: #e0e0e0;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.asset-preview {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		margin-top: 1rem;
		padding: 1rem;
		border-radius: 12px;
		background: rgba(212, 175, 55, 0.05);
		border: 1px solid rgba(212, 175, 55, 0.15);
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
		color: rgba(224, 224, 224, 0.5);
		font-weight: 600;
	}

	.preview-value {
		font-size: 0.95rem;
		color: #d4af37;
		font-weight: 600;
	}

	.empty-assets {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
		padding: 2rem 1.5rem;
		border-radius: 10px;
		background: rgba(231, 76, 60, 0.05);
		border: 1px dashed rgba(231, 76, 60, 0.2);
		text-align: center;
	}

	.empty-text {
		margin: 0;
		font-size: 0.95rem;
		color: rgba(224, 224, 224, 0.7);
		font-weight: 500;
	}

	.empty-hint {
		margin: 0;
		font-size: 0.85rem;
		color: rgba(224, 224, 224, 0.5);
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
		color: #e0e0e0;
		letter-spacing: 0.3px;
	}

	.required {
		color: #e74c3c;
	}

	.form-input,
	.form-select,
	.form-textarea {
		padding: 0.85rem 1rem;
		border: 1.5px solid rgba(212, 175, 55, 0.25);
		border-radius: 10px;
		background: rgba(15, 20, 25, 0.5);
		color: #e0e0e0;
		font-size: 0.95rem;
		font-family: 'Poppins', system-ui, sans-serif;
		transition: all 0.3s ease;
	}

	.form-input::placeholder,
	.form-textarea::placeholder {
		color: rgba(224, 224, 224, 0.4);
	}

	.form-input:focus,
	.form-select:focus,
	.form-textarea:focus {
		outline: none;
		border-color: #d4af37;
		background: rgba(15, 20, 25, 0.7);
		box-shadow: 0 0 0 3px rgba(212, 175, 55, 0.1);
	}

	.form-textarea {
		resize: vertical;
		min-height: 100px;
	}

	.field-hint {
		margin: 0.4rem 0 0;
		font-size: 0.8rem;
		color: rgba(224, 224, 224, 0.4);
		font-style: italic;
	}

	.empty-platforms {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		padding: 1.5rem;
		border-radius: 10px;
		background: rgba(212, 175, 55, 0.05);
		border: 1px dashed rgba(212, 175, 55, 0.2);
		text-align: center;
	}

	.empty-text {
		margin: 0;
		font-size: 0.95rem;
		color: rgba(224, 224, 224, 0.6);
	}

	.btn-link {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.65rem 1.2rem;
		border: none;
		background: rgba(212, 175, 55, 0.15);
		color: #d4af37;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.btn-link:hover {
		background: rgba(212, 175, 55, 0.25);
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
		color: rgba(224, 224, 224, 0.5);
		font-weight: 600;
		pointer-events: none;
	}

	.input-addon .form-input {
		padding-left: 2.5rem;
	}

	.value-display {
		padding: 0.85rem 1rem;
		border: 1.5px solid rgba(212, 175, 55, 0.25);
		border-radius: 10px;
		background: rgba(15, 20, 25, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 44px;
	}

	.total-value {
		margin: 0;
		font-size: 1.2rem;
		font-weight: 700;
		color: #d4af37;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.summary-card {
		border: 2px solid rgba(212, 175, 55, 0.3);
		border-radius: 16px;
		background: linear-gradient(135deg, rgba(212, 175, 55, 0.08) 0%, rgba(46, 204, 113, 0.05) 100%);
		padding: 1.5rem;
		animation: slide-in 0.4s ease-out;
	}

	.summary-title {
		margin: 0 0 1.25rem;
		font-size: 1rem;
		font-weight: 700;
		color: #d4af37;
		font-family: 'Poppins', system-ui, sans-serif;
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
		border-top: 1px solid rgba(212, 175, 55, 0.2);
		padding-top: 1rem;
		margin-top: 0.5rem;
	}

	.summary-label {
		font-size: 0.9rem;
		color: rgba(224, 224, 224, 0.65);
		font-weight: 500;
	}

	.summary-value {
		font-size: 0.95rem;
		color: #e0e0e0;
		font-weight: 600;
	}

	.summary-value.highlight {
		color: #d4af37;
		font-size: 1.1rem;
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
		font-family: 'Poppins', system-ui, sans-serif;
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
		background: linear-gradient(135deg, #d4af37, #e8c547);
		color: #0f1419;
		font-weight: 700;
	}

	.btn-primary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 175, 55, 0.25);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-secondary {
		background: transparent;
		color: #e0e0e0;
		border: 1.5px solid rgba(212, 175, 55, 0.25);
	}

	.btn-secondary:hover {
		border-color: #d4af37;
		background: rgba(212, 175, 55, 0.1);
		color: #d4af37;
	}

	.spinner {
		display: inline-block;
		width: 14px;
		height: 14px;
		border: 2px solid rgba(15, 20, 25, 0.3);
		border-top-color: #0f1419;
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
		.page-title {
			font-size: 1.85rem;
		}

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
</style>
