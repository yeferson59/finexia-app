<script lang="ts">
	import { goto } from '$app/navigation';

	interface FormData {
		assetType: string;
		symbol: string;
		name: string;
		quantity: string;
		purchasePrice: string;
		purchaseDate: string;
		totalValue: number;
		notes: string;
		brokerName: string;
	}

	let formData: FormData = $state({
		assetType: 'Acciones',
		symbol: '',
		name: '',
		quantity: '',
		purchasePrice: '',
		purchaseDate: new Date().toISOString().split('T')[0],
		totalValue: 0,
		notes: '',
		brokerName: ''
	});

	let isSubmitting = $state(false);
	let submitSuccess = $state(false);

	const assetTypes = [
		{ value: 'Acciones', label: 'Acciones', icon: '📈' },
		{ value: 'Criptomonedas', label: 'Criptomonedas', icon: '₿' },
		{ value: 'Bonos', label: 'Bonos', icon: '📊' },
		{ value: 'ETF', label: 'ETF', icon: '🎯' },
		{ value: 'Commodities', label: 'Commodities', icon: '⛏️' },
		{ value: 'Fondos', label: 'Fondos Mutuos', icon: '💼' },
		{ value: 'Opciones', label: 'Opciones', icon: '⚙️' },
		{ value: 'Divisas', label: 'Divisas', icon: '💱' }
	];

	const brokers = [
		'Interactive Brokers',
		'Charles Schwab',
		'E-Trade',
		'Fidelity',
		'Coinbase',
		'Kraken',
		'Binance',
		'TD Ameritrade',
		'Otro'
	];

	$effect(() => {
		const qty = parseFloat(formData.quantity) || 0;
		const price = parseFloat(formData.purchasePrice) || 0;
		formData.totalValue = qty * price;
	});

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!formData.symbol || !formData.name || !formData.quantity || !formData.purchasePrice) {
			alert('Por favor completa todos los campos requeridos');
			return;
		}

		isSubmitting = true;
		try {
			await new Promise((resolve) => setTimeout(resolve, 1000));
			submitSuccess = true;
			setTimeout(() => {
				goto('/dashboard/portfolio');
			}, 1500);
		} catch (error) {
			console.error('Error:', error);
		} finally {
			isSubmitting = false;
		}
	}

	function handleCancel() {
		goto('/dashboard/portfolio');
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
	<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
		<!-- Asset Type Selection -->
		<section class="form-section">
			<h2 class="section-title">Tipo de Activo</h2>
			<div class="asset-type-grid">
				{#each assetTypes as type}
					<label class="asset-type-card" class:active={formData.assetType === type.value}>
						<input
							type="radio"
							name="assetType"
							value={type.value}
							bind:group={formData.assetType}
							class="radio-input"
						/>
						<span class="type-icon">{type.icon}</span>
						<span class="type-label">{type.label}</span>
					</label>
				{/each}
			</div>
		</section>

		<!-- Asset Information -->
		<section class="form-section">
			<h2 class="section-title">Información del Activo</h2>

			<div class="form-row">
				<div class="form-group">
					<label for="symbol" class="form-label">Símbolo/Ticker <span class="required">*</span></label>
					<input
						id="symbol"
						type="text"
						bind:value={formData.symbol}
						placeholder="ej: AAPL, BTC, TSLA"
						class="form-input"
						required
						maxlength="10"
					/>
					<p class="field-hint">Símbolo del mercado (ej: AAPL para Apple)</p>
				</div>

				<div class="form-group">
					<label for="name" class="form-label">Nombre Completo <span class="required">*</span></label>
					<input
						id="name"
						type="text"
						bind:value={formData.name}
						placeholder="ej: Apple Inc."
						class="form-input"
						required
					/>
					<p class="field-hint">Nombre completo del activo</p>
				</div>
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
					<label for="purchasePrice" class="form-label">Precio de Compra <span class="required">*</span></label>
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
					<label class="form-label">Valor Total Invertido</label>
					<div class="value-display">
						<p class="total-value">{formatCurrency(formData.totalValue)}</p>
					</div>
				</div>
			</div>
		</section>

		<!-- Broker & Additional Info -->
		<section class="form-section">
			<h2 class="section-title">Información Adicional</h2>

			<div class="form-group">
				<label for="brokerName" class="form-label">Broker/Proveedor</label>
				<select id="brokerName" bind:value={formData.brokerName} class="form-select">
					<option value="">Selecciona un broker</option>
					{#each brokers as broker}
						<option value={broker}>{broker}</option>
					{/each}
				</select>
				<p class="field-hint">Dónde compraste este activo</p>
			</div>

			<div class="form-group">
				<label for="notes" class="form-label">Notas</label>
				<textarea
					id="notes"
					bind:value={formData.notes}
					placeholder="Agrega observaciones, estrategia o detalles especiales..."
					class="form-textarea"
					rows="3"
				></textarea>
				<p class="field-hint">Notas personales sobre este activo</p>
			</div>
		</section>

		<!-- Summary Card -->
		{#if formData.quantity && formData.purchasePrice}
			<section class="summary-card">
				<h3 class="summary-title">Resumen de Inversión</h3>
				<div class="summary-items">
					<div class="summary-item">
						<span class="summary-label">Activo</span>
						<span class="summary-value">{formData.symbol} - {formData.name}</span>
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

	.asset-type-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 1rem;
	}

	.asset-type-card {
		position: relative;
		padding: 1.25rem;
		border: 2px solid rgba(212, 175, 55, 0.15);
		border-radius: 12px;
		background: rgba(15, 20, 25, 0.5);
		cursor: pointer;
		transition: all 0.3s ease;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		text-align: center;
	}

	.asset-type-card:hover {
		border-color: rgba(212, 175, 55, 0.4);
		background: rgba(15, 20, 25, 0.7);
	}

	.asset-type-card.active {
		border-color: #d4af37;
		background: rgba(212, 175, 55, 0.15);
		box-shadow: 0 0 0 3px rgba(212, 175, 55, 0.1);
	}

	.radio-input {
		position: absolute;
		opacity: 0;
		cursor: pointer;
		width: 0;
		height: 0;
	}

	.type-icon {
		font-size: 1.8rem;
		display: block;
	}

	.type-label {
		font-size: 0.85rem;
		font-weight: 600;
		color: #e0e0e0;
		letter-spacing: 0.2px;
	}

	.asset-type-card.active .type-label {
		color: #d4af37;
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
		.asset-type-grid {
			grid-template-columns: repeat(3, 1fr);
		}
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}

		.asset-type-grid {
			grid-template-columns: repeat(2, 1fr);
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
