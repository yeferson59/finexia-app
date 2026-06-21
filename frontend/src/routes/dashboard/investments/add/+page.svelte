<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$components/ui/page-header.svelte';
	import { investmentStore } from '$lib/stores/investments.svelte';

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
	let errors: Partial<Record<keyof FormData, string>> = $state({});

	const investmentTypes = ['Fondos', 'Acciones', 'ETF', 'Bonos', 'Criptomonedas', 'Derivados'];
	const riskLevels = ['Bajo', 'Medio', 'Alto', 'Muy Alto'];
	const categories = [
		'Tecnología',
		'Energía Renovable',
		'Mercados Emergentes',
		'Inmuebles',
		'Oro',
		'Divisas'
	];

	function validate(): boolean {
		const nextErrors: Partial<Record<keyof FormData, string>> = {};
		if (!formData.name.trim()) nextErrors.name = 'El nombre es obligatorio';
		if (!formData.description.trim()) nextErrors.description = 'La descripción es obligatoria';
		if (!formData.expectedROI) nextErrors.expectedROI = 'Indica el ROI esperado';
		if (!formData.horizon) nextErrors.horizon = 'Indica el horizonte de inversión';
		errors = nextErrors;
		return Object.keys(nextErrors).length === 0;
	}

	function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!validate()) return;

		isSubmitting = true;
		try {
			investmentStore.addInvestment({
				name: formData.name.trim(),
				description: formData.description.trim(),
				type: formData.type,
				category: formData.category,
				riskLevel: formData.riskLevel,
				expectedROI: Number(formData.expectedROI),
				horizon: Number(formData.horizon),
				minimumInvestment: Number(formData.minimumInvestment) || 0,
				status: formData.status
			});
			submitSuccess = true;
			setTimeout(() => {
				goto(resolve('/dashboard/investments'));
			}, 1500);
		} catch (error) {
			console.error('Error:', error);
			isSubmitting = false;
		}
	}

	function handleCancel() {
		goto(resolve('/dashboard/investments'));
	}
</script>

<svelte:head>
	<title>Agregar Producto de Inversión - FINEXIA</title>
	<meta name="description" content="Crea un nuevo producto de inversión" />
</svelte:head>

<PageHeader
	title="Agregar Producto de Inversión"
	subtitle="Configura los detalles de tu nuevo producto de inversión"
/>

<div class="form-container">
	<form onsubmit={handleSubmit} class="investment-form">
		<!-- Main Information Section -->
		<section class="form-section">
			<h2 class="section-title">Información Básica</h2>

			<div class="form-group">
				<label for="name" class="form-label"
					>Nombre del Producto <span class="required">*</span></label
				>
				<input
					id="name"
					type="text"
					bind:value={formData.name}
					placeholder="ej: Fondo Crecimiento Global"
					class="form-input"
					required
				/>
				{#if errors.name}<span class="field-error">{errors.name}</span>{/if}
			</div>

			<div class="form-group">
				<label for="description" class="form-label"
					>Descripción <span class="required">*</span></label
				>
				<textarea
					id="description"
					bind:value={formData.description}
					placeholder="Describe los objetivos, estrategia y características del producto..."
					class="form-textarea"
					rows="4"
					required></textarea>
				{#if errors.description}<span class="field-error">{errors.description}</span>{/if}
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="type" class="form-label">Tipo de Instrumento</label>
					<select id="type" bind:value={formData.type} class="form-select">
						{#each investmentTypes as type (type)}
							<option value={type}>{type}</option>
						{/each}
					</select>
				</div>

				<div class="form-group">
					<label for="category" class="form-label">Categoría</label>
					<select id="category" bind:value={formData.category} class="form-select">
						{#each categories as category (category)}
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
					<label for="roi" class="form-label"
						>ROI Esperado (%) <span class="required">*</span></label
					>
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
					{#if errors.expectedROI}<span class="field-error">{errors.expectedROI}</span>{/if}
				</div>

				<div class="form-group">
					<label for="horizon" class="form-label"
						>Horizonte de Inversión <span class="required">*</span></label
					>
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
					{#if errors.horizon}<span class="field-error">{errors.horizon}</span>{/if}
				</div>
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="risk" class="form-label">Nivel de Riesgo</label>
					<select id="risk" bind:value={formData.riskLevel} class="form-select">
						{#each riskLevels as level (level)}
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
	.form-container {
		max-width: 900px;
	}

	.investment-form {
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

	.field-error {
		font-size: 0.8rem;
		color: var(--red);
		letter-spacing: 0.2px;
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
		color: rgba(236, 234, 229, 0.4);
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
		color: rgba(236, 234, 229, 0.5);
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
