<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';

	interface FormData {
		name: string;
		description: string;
		type: string;
		status: string;
	}

	let formData: FormData = $state({
		name: '',
		description: '',
		type: 'Bróker',
		status: 'Activo'
	});

	let isSubmitting = $state(false);
	let submitSuccess = $state(false);

	const platformTypes = new Map<string, string>([
		['broker', 'Bróker'],
		['investment_bank', 'Banco de Inversión'],
		['trading_platform', 'Plataforma de Trading'],
		['neobank', 'NeoBank'],
		['de_fi', 'DeFi'],
		['crypto_wallet', 'Billetera Cripto'],
		['mutual_funds', 'Fondos Mutuos'],
		['brokerage_house', 'Casa de Bolsa'],
		['other', 'Otro']
	]);

	function handleCancel() {
		goto(resolve('/dashboard/platforms'));
	}
</script>

<svelte:head>
	<title>Agregar Plataforma - FINEXIA</title>
	<meta name="description" content="Registra una nueva plataforma de inversión" />
</svelte:head>

<PageHeader
	title="Agregar Plataforma de Inversión"
	subtitle="Completa los datos de tu nueva plataforma para comenzar a rastrear tus inversiones"
/>

<div class="form-container">
	<form
		method="POST"
		action="/dashboard/platforms/add"
		class="platform-form"
		use:enhance={() => {
			isSubmitting = true;
			return async ({ update }) => {
				await update();
				isSubmitting = false;
			};
		}}
	>
		<!-- Basic Information Section -->
		<section class="form-section">
			<h2 class="section-title">Información Básica</h2>

			<div class="form-group">
				<label for="name" class="form-label"
					>Nombre de la Plataforma <span class="required">*</span></label
				>
				<input
					id="name"
					name="name"
					type="text"
					bind:value={formData.name}
					placeholder="ej: Interactive Brokers"
					class="form-input"
					required
				/>
			</div>

			<div class="form-group">
				<label for="description" class="form-label">Descripción</label>
				<textarea
					id="description"
					name="description"
					bind:value={formData.description}
					placeholder="Describe qué tipo de inversiones realizas en esta plataforma..."
					class="form-textarea"
					rows="4"></textarea>
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="type" class="form-label"
						>Tipo de Plataforma <span class="required">*</span></label
					>
					<select id="type" name="type" bind:value={formData.type} class="form-select" required>
						{#each platformTypes.entries() as [key, type] (key)}
							<option value={key}>{type}</option>
						{/each}
					</select>
				</div>
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
					Crear Plataforma
				{/if}
			</button>
		</div>
	</form>
</div>

<style>
	.form-container {
		max-width: 900px;
	}

	.platform-form {
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
