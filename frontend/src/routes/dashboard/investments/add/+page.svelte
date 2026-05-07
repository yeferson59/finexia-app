<script lang="ts">
	import { goto } from '$app/navigation';

	interface FormData {
		name: string;
		description: string;
		type: string;
		riskLevel: string;
		expectedROI: string;
		horizon: string;
		minimumInvestment: string;
		category: string;
		status: string;
	}

	let formData: FormData = $state({
		name: '',
		description: '',
		type: 'Fondos',
		riskLevel: 'Medio',
		expectedROI: '',
		horizon: '',
		minimumInvestment: '',
		category: 'Tecnología',
		status: 'Activo'
	});

	let isSubmitting = $state(false);
	let submitSuccess = $state(false);

	const investmentTypes = ['Fondos', 'Acciones', 'ETF', 'Bonos', 'Criptomonedas', 'Derivados'];
	const riskLevels = ['Bajo', 'Medio', 'Alto', 'Muy Alto'];
	const categories = ['Tecnología', 'Energía Renovable', 'Mercados Emergentes', 'Inmuebles', 'Oro', 'Divisas'];

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!formData.name || !formData.description || !formData.expectedROI || !formData.horizon) {
			alert('Por favor completa todos los campos requeridos');
			return;
		}

		isSubmitting = true;
		try {
			// Simular API call
			await new Promise((resolve) => setTimeout(resolve, 1000));
			submitSuccess = true;
			setTimeout(() => {
				goto('/dashboard/investments');
			}, 1500);
		} catch (error) {
			console.error('Error:', error);
		} finally {
			isSubmitting = false;
		}
	}

	function handleCancel() {
		goto('/dashboard/investments');
	}
</script>

<svelte:head>
	<title>Agregar Producto de Inversión - FINEXIA</title>
	<meta name="description" content="Crea un nuevo producto de inversión" />
</svelte:head>

<header class="page-header">
	<h1 class="page-title">Agregar Producto de Inversión</h1>
	<p class="page-subtitle">Configura los detalles de tu nuevo producto de inversión</p>
</header>

<div class="form-container">
	<form onsubmit={handleSubmit} class="investment-form">
		<!-- Main Information Section -->
		<section class="form-section">
			<h2 class="section-title">Información Básica</h2>

			<div class="form-group">
				<label for="name" class="form-label">Nombre del Producto <span class="required">*</span></label>
				<input
					id="name"
					type="text"
					bind:value={formData.name}
					placeholder="ej: Fondo Crecimiento Global"
					class="form-input"
					required
				/>
			</div>

			<div class="form-group">
				<label for="description" class="form-label">Descripción <span class="required">*</span></label>
				<textarea
					id="description"
					bind:value={formData.description}
					placeholder="Describe los objetivos, estrategia y características del producto..."
					class="form-textarea"
					rows="4"
					required
				></textarea>
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="type" class="form-label">Tipo de Instrumento</label>
					<select id="type" bind:value={formData.type} class="form-select">
						{#each investmentTypes as type}
							<option value={type}>{type}</option>
						{/each}
					</select>
				</div>

				<div class="form-group">
					<label for="category" class="form-label">Categoría</label>
					<select id="category" bind:value={formData.category} class="form-select">
						{#each categories as category}
							<option value={category}>{category}</option>
						{/each}
					</select>
				</div>
			</div>
		</section>

		<!-- Financial Details Section -->
		<section class="form-section">
			<h2 class="section-title">Detalles Financieros</h2>

			<div class="form-row">
				<div class="form-group">
					<label for="roi" class="form-label">ROI Esperado (%) <span class="required">*</span></label>
					<div class="input-addon">
						<input
							id="roi"
							type="number"
							bind:value={formData.expectedROI}
							placeholder="12.5"
							class="form-input"
							min="0"
							max="100"
							step="0.1"
							required
						/>
						<span class="addon-text">%</span>
					</div>
				</div>

				<div class="form-group">
					<label for="horizon" class="form-label">Horizonte de Inversión <span class="required">*</span></label>
					<div class="input-addon">
						<input
							id="horizon"
							type="number"
							bind:value={formData.horizon}
							placeholder="18"
							class="form-input"
							min="1"
							required
						/>
						<span class="addon-text">meses</span>
					</div>
				</div>
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="risk" class="form-label">Nivel de Riesgo</label>
					<select id="risk" bind:value={formData.riskLevel} class="form-select">
						{#each riskLevels as level}
							<option value={level}>{level}</option>
						{/each}
					</select>
				</div>

				<div class="form-group">
					<label for="minimum" class="form-label">Inversión Mínima ($)</label>
					<div class="input-addon">
						<span class="addon-text">$</span>
						<input
							id="minimum"
							type="number"
							bind:value={formData.minimumInvestment}
							placeholder="1000"
							class="form-input"
							min="0"
						/>
					</div>
				</div>
			</div>
		</section>

		<!-- Status Section -->
		<section class="form-section">
			<h2 class="section-title">Configuración</h2>

			<div class="form-group">
				<label for="status" class="form-label">Estado del Producto</label>
				<select id="status" bind:value={formData.status} class="form-select">
					<option value="Activo">Activo</option>
					<option value="Inactivo">Inactivo</option>
					<option value="Pendiente">Pendiente de Aprobación</option>
				</select>
			</div>
		</section>

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
					Crear Producto
				{/if}
			</button>
		</div>
	</form>
</div>

<style>
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
		max-width: 900px;
	}

	.investment-form {
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

	.input-addon {
		position: relative;
		display: flex;
		align-items: center;
	}

	.input-addon .form-input {
		padding-left: 2.5rem;
	}

	.addon-text {
		position: absolute;
		left: 1rem;
		font-size: 0.9rem;
		color: rgba(224, 224, 224, 0.5);
		font-weight: 600;
		pointer-events: none;
	}

	.input-addon:has(.addon-text:last-child) .form-input {
		padding-right: 2.5rem;
		padding-left: 1rem;
	}

	.input-addon:has(.addon-text:last-child) .addon-text {
		right: 1rem;
		left: auto;
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

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
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

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
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
