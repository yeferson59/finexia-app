<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';

	interface FormData {
		name: string;
		description: string;
		type: string;
		status: string;
		investments: number;
		totalValue: string;
	}

	let formData: FormData = $state({
		name: 'Interactive Brokers',
		description: 'Plataforma de trading internacional con acceso a múltiples mercados',
		type: 'Bróker',
		status: 'Activo',
		investments: 12,
		totalValue: '$45,230.50'
	});

	let isEditing = $state(false);
	let isSubmitting = $state(false);
	let submitSuccess = $state(false);

	const platformTypes = [
		'Bróker',
		'Banco de Inversión',
		'Plataforma de Trading',
		'DeFi',
		'Billetera Cripto',
		'Fondos Mutuos',
		'Casa de Bolsa',
		'Otro'
	];

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!formData.name || !formData.type) {
			alert('Por favor completa los campos requeridos');
			return;
		}

		isSubmitting = true;
		try {
			await new Promise((resolve) => setTimeout(resolve, 1000));
			submitSuccess = true;
			isEditing = false;
			setTimeout(() => {
				submitSuccess = false;
			}, 2000);
		} catch (error) {
			console.error('Error:', error);
		} finally {
			isSubmitting = false;
		}
	}

	function toggleEdit() {
		isEditing = !isEditing;
		submitSuccess = false;
	}

	function handleCancel() {
		if (isEditing) {
			isEditing = false;
		} else {
			goto('/dashboard/platforms');
		}
	}

	function goBack() {
		goto('/dashboard/platforms');
	}
</script>

<svelte:head>
	<title>{formData.name} - FINEXIA</title>
	<meta name="description" content={`Detalles de ${formData.name}`} />
</svelte:head>

<div class="page-container">
	<header class="page-header">
		<div class="header-top">
			<button onclick={goBack} class="btn-back" aria-label="Volver">
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
			</button>
			<div class="header-content">
				<h1 class="page-title">{formData.name}</h1>
				<p class="page-subtitle">{formData.type}</p>
			</div>
			<div class="header-actions">
				{#if !isEditing}
					<button onclick={toggleEdit} class="btn-edit">
						<svg
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
							<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
						</svg>
						Editar
					</button>
				{/if}
			</div>
		</div>
	</header>

	<div class="content-grid">
		<!-- Main Content -->
		<main class="main-content">
			{#if !isEditing}
				<!-- View Mode -->
				<section class="panel info-section">
					<h2 class="section-title">Información de la Plataforma</h2>

					<div class="info-group">
						<div class="info-item">
							<span class="info-label">Estado</span>
							<div
								class="status-badge"
								style="--status-color: {formData.status === 'Activo'
									? 'var(--green)'
									: 'var(--red)'}"
							>
								{formData.status}
							</div>
						</div>
						<div class="info-item">
							<span class="info-label">Tipo de Plataforma</span>
							<span class="info-value">{formData.type}</span>
						</div>
					</div>

					{#if formData.description}
						<div class="info-description">
							<h3>Descripción</h3>
							<p>{formData.description}</p>
						</div>
					{/if}
				</section>

				<!-- Statistics Panel -->
				<section class="panel stats-section">
					<h2 class="section-title">Resumen de Inversiones</h2>
					<div class="stats-grid">
						<div class="stat-card">
							<span class="stat-icon">📊</span>
							<div class="stat-content">
								<span class="stat-label">Número de Inversiones</span>
								<span class="stat-value">{formData.investments}</span>
							</div>
						</div>
						<div class="stat-card">
							<span class="stat-icon">💰</span>
							<div class="stat-content">
								<span class="stat-label">Valor Total Invertido</span>
								<span class="stat-value">{formData.totalValue}</span>
							</div>
						</div>
					</div>
				</section>
			{:else}
				<!-- Edit Mode -->
				<section class="panel edit-section">
					<h2 class="section-title">Editar Plataforma</h2>

					<form onsubmit={handleSubmit} class="platform-form">
						<div class="form-group">
							<label for="name" class="form-label"
								>Nombre de la Plataforma <span class="required">*</span></label
							>
							<input id="name" type="text" bind:value={formData.name} class="form-input" required />
						</div>

						<div class="form-group">
							<label for="description" class="form-label">Descripción</label>
							<textarea
								id="description"
								bind:value={formData.description}
								class="form-textarea"
								rows="4"
							></textarea>
						</div>

						<div class="form-row">
							<div class="form-group">
								<label for="type" class="form-label"
									>Tipo de Plataforma <span class="required">*</span></label
								>
								<select id="type" bind:value={formData.type} class="form-select" required>
									{#each platformTypes as type}
										<option value={type}>{type}</option>
									{/each}
								</select>
							</div>

							<div class="form-group">
								<label for="status" class="form-label">Estado</label>
								<select id="status" bind:value={formData.status} class="form-select">
									<option value="Activo">Activo</option>
									<option value="Inactivo">Inactivo</option>
								</select>
							</div>
						</div>

						<div class="form-actions">
							<button type="button" onclick={handleCancel} class="btn btn-secondary">
								Cancelar
							</button>
							<button type="submit" disabled={isSubmitting} class="btn btn-primary">
								{#if isSubmitting}
									<span class="spinner"></span>
									Guardando...
								{:else if submitSuccess}
									✓ Guardado
								{:else}
									Guardar Cambios
								{/if}
							</button>
						</div>
					</form>
				</section>
			{/if}
		</main>
	</div>
</div>

<style>
	.page-container {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.page-header {
		margin-bottom: 1rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.header-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
		flex-wrap: wrap;
	}

	.btn-back {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border: none;
		border-radius: 8px;
		background: transparent;
		color: var(--amber);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-back:hover {
		background: var(--border);
		transform: translateX(-2px);
	}

	.header-content {
		flex: 1;
		min-width: 200px;
	}

	.page-title {
		margin: 0 0 0.3rem;
		font-size: 1.85rem;
		font-weight: 300;
		color: var(--text);
		font-family: var(--font-display);
		letter-spacing: -0.02em;
	}

	.page-subtitle {
		margin: 0;
		color: rgba(236, 234, 229, 0.6);
		font-size: 0.95rem;
	}

	.header-actions {
		display: flex;
		gap: 0.75rem;
	}

	.btn-edit {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.2rem;
		border: 1.5px solid rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: transparent;
		color: var(--amber);
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-edit:hover {
		border-color: var(--amber);
		background: var(--border);
		transform: translateY(-2px);
	}

	.content-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 2rem;
	}

	.main-content {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.panel {
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

	.info-group {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.info-label {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.3px;
		font-weight: 600;
	}

	.info-value {
		color: var(--text);
		font-size: 0.95rem;
	}

	.status-badge {
		width: fit-content;
		padding: 0.4rem 0.8rem;
		border-radius: 6px;
		background: var(--status-color, var(--amber));
		color: #0d0800;
		font-size: 0.8rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.info-description {
		margin-top: 1.5rem;
		padding-top: 1.5rem;
		border-top: 1px solid var(--border);
	}

	.info-description h3 {
		margin: 0 0 0.75rem;
		color: var(--text);
		font-size: 0.95rem;
		font-weight: 600;
	}

	.info-description p {
		margin: 0;
		color: rgba(236, 234, 229, 0.75);
		line-height: 1.6;
		font-size: 0.9rem;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
	}

	.stat-card {
		display: flex;
		gap: 1rem;
		padding: 1.25rem;
		border-radius: 12px;
		background: var(--border);
		border: 1px solid var(--border);
	}

	.stat-icon {
		font-size: 1.75rem;
		display: flex;
		align-items: center;
	}

	.stat-content {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.stat-label {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.6);
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.stat-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--amber);
		font-size: 1.25rem;
		font-weight: 700;
		font-family: var(--font-mono);
	}

	.edit-section {
		animation: fade-in 0.3s ease-out;
	}

	.platform-form {
		display: flex;
		flex-direction: column;
		gap: 1.35rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
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

	.form-input::placeholder {
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

	.form-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 1rem;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.9rem;
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

	@media (max-width: 1024px) {
		.content-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.5rem;
		}

		.header-top {
			flex-direction: column;
			align-items: flex-start;
		}

		.header-actions {
			width: 100%;
		}

		.btn-edit {
			width: 100%;
			justify-content: center;
		}

		.info-group {
			grid-template-columns: 1fr;
		}

		.stats-grid {
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
