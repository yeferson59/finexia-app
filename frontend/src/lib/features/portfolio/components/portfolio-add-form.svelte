<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$lib/ui/page-header.svelte';
	import { PORTFOLIO_TYPES } from '../portfolio';

	let { risks }: { risks: { id: string; name: string; description: string }[] } = $props();

	interface FormData {
		name: string;
		description: string;
		type: string;
		riskLevel: string;
		currency: string;
		targetAmount: string;
		isDefault: boolean;
	}

	let formData: FormData = $state({
		name: '',
		description: '',
		type: 'stocks_etfs',
		riskLevel: '',
		currency: 'USD',
		targetAmount: '',
		isDefault: false
	});

	let isSubmitting = $state(false);
	let submitSuccess = $state(false);
	let errors: Record<string, string> = $state({});

	const currencies = ['USD', 'COP', 'EUR', 'MXN', 'ARS'];

	function handleCancel() {
		goto(resolve('/dashboard/portfolios'));
	}
</script>

<button class="back-button" onclick={handleCancel} aria-label="Volver a portafolios">
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

<main class="form-container">
	<PageHeader
		title="Crear Nuevo Portafolio"
		subtitle="Configura un nuevo portafolio para gestionar tus inversiones"
	/>

	{#if submitSuccess}
		<div class="success-message">
			<svg
				width="24"
				height="24"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<polyline points="20 6 9 17 4 12"></polyline>
			</svg>
			<span>Portafolio creado exitosamente</span>
		</div>
	{/if}

	<form
		method="POST"
		action="/dashboard/portfolios/add"
		class="form"
		use:enhance={() => {
			isSubmitting = true;
			return async ({ update }) => {
				await update();
				isSubmitting = false;
			};
		}}
	>
		<fieldset class="form-section">
			<legend class="section-title">Información Básica</legend>

			<div class="form-group">
				<label for="name" class="label">Nombre del Portafolio *</label>
				<input
					type="text"
					id="name"
					name="name"
					bind:value={formData.name}
					placeholder="Ej: Mi Portafolio Principal"
					class="input"
					class:error={errors.name}
					disabled={isSubmitting}
				/>
				{#if errors.name}
					<span class="error-message">{errors.name}</span>
				{/if}
			</div>

			<div class="form-group">
				<label for="description" class="label">Descripción (opcional)</label>
				<textarea
					id="description"
					name="description"
					bind:value={formData.description}
					placeholder="Describe el propósito de este portafolio"
					class="textarea"
					disabled={isSubmitting}
					rows="3"></textarea>
			</div>
		</fieldset>

		<fieldset class="form-section">
			<legend class="section-title">Características del Portafolio</legend>

			<div class="form-row">
				<div class="form-group">
					<label for="type" class="label">Tipo de Portafolio *</label>
					<select
						id="type"
						bind:value={formData.type}
						name="type"
						class="select"
						disabled={isSubmitting}
					>
						{#each PORTFOLIO_TYPES as type (type.value)}
							<option value={type.value}>{type.label}</option>
						{/each}
					</select>
				</div>

				<div class="form-group">
					<label for="currency" class="label">Moneda *</label>
					<select
						id="currency"
						bind:value={formData.currency}
						class="select"
						name="currency"
						disabled={isSubmitting}
					>
						{#each currencies as curr (curr)}
							<option value={curr}>{curr}</option>
						{/each}
					</select>
				</div>
			</div>

			<div class="form-group">
				<label class="label" for="risk">Nivel de Riesgo *</label>
				<fieldset class="risk-options">
					{#each risks as risk (risk.id)}
						<label class="radio-label">
							<input
								type="radio"
								name="riskId"
								value={risk.id}
								bind:group={formData.riskLevel}
								disabled={isSubmitting}
								required
							/>
							<span class="radio-content">
								<span class="radio-title">{risk.name}</span>
								<span class="radio-description">{risk.description}</span>
							</span>
						</label>
					{/each}
				</fieldset>
			</div>
		</fieldset>

		<fieldset class="form-section">
			<legend class="section-title">Objetivo Financiero</legend>

			<div class="form-group">
				<label for="targetAmount" class="label">Monto Objetivo (opcional)</label>
				<div class="input-with-prefix">
					<span class="prefix">{formData.currency}</span>
					<input
						type="number"
						id="targetAmount"
						bind:value={formData.targetAmount}
						placeholder="0.00"
						class="input"
						name="priceValue"
						class:error={errors.targetAmount}
						disabled={isSubmitting}
						step="0.01"
						min="0"
					/>
				</div>
				{#if errors.targetAmount}
					<span class="error-message">{errors.targetAmount}</span>
				{/if}
				<p class="help-text">Define el monto que deseas alcanzar en este portafolio</p>
			</div>

			<div class="form-group">
				<label class="checkbox-label" for="isDefault">
					<input
						type="checkbox"
						id="isDefault"
						name="isDefault"
						bind:checked={formData.isDefault}
						disabled={isSubmitting}
					/>
					<span class="checkbox-content">
						<span class="checkbox-title">Marcar como portafolio por defecto</span>
						<span class="checkbox-description">
							Este portafolio se usará como selección predeterminada
						</span>
					</span>
				</label>
			</div>
		</fieldset>

		<div class="form-actions">
			<button type="button" onclick={handleCancel} class="btn-cancel" disabled={isSubmitting}>
				Cancelar
			</button>
			<button type="submit" class="btn-submit" disabled={isSubmitting}>
				{#if isSubmitting}
					<span class="spinner"></span>
					Creando...
				{:else}
					Crear Portafolio
				{/if}
			</button>
		</div>
	</form>
</main>

<style>
	.back-button {
		display: inline-flex;
		align-items: center;
		gap: 0.6rem;
		margin-bottom: 2rem;
		padding: 0.7rem 1.2rem;
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: var(--surface);
		color: var(--amber);
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.back-button:hover {
		background: var(--border-strong);
		border-color: rgba(212, 145, 42, 0.5);
		transform: translateX(-2px);
	}

	.form-container {
		max-width: 800px;
		margin: 0 auto;
	}

	.success-message {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 2rem;
		padding: 1rem 1.5rem;
		border-radius: 12px;
		background: rgba(34, 201, 126, 0.1);
		border: 1px solid rgba(34, 201, 126, 0.3);
		color: var(--green);
		font-weight: 600;
	}

	.form {
		display: grid;
		gap: 2rem;
	}

	.form-section {
		display: grid;
		gap: 1.5rem;
		padding: 1.5rem;
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		backdrop-filter: blur(16px);
	}

	.section-title {
		margin: 0 0 0.5rem;
		font-size: 1.15rem;
		font-weight: 400;
		color: var(--amber-light);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
	}

	.form-group {
		display: grid;
		gap: 0.6rem;
	}

	.label {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
	}

	.input,
	.select,
	.textarea {
		padding: 0.85rem;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.95rem;
		transition: all 0.3s ease;
	}

	.input:focus,
	.select:focus,
	.textarea:focus {
		outline: none;
		border-color: var(--amber);
		background: rgba(255, 255, 255, 0.022);
		box-shadow: 0 0 0 3px var(--border);
	}

	.input::placeholder,
	.textarea::placeholder {
		color: rgba(236, 234, 229, 0.4);
	}

	.input:disabled,
	.select:disabled,
	.textarea:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.input.error,
	.input.error:focus {
		box-shadow: 0 0 0 3px rgba(224, 90, 90, 0.1);
	}

	.error-message {
		font-size: 0.8rem;
		color: var(--red);
	}

	.help-text {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
	}

	.input-with-prefix {
		display: flex;
		align-items: center;
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		overflow: hidden;
		transition: all 0.3s ease;
	}

	.input-with-prefix:focus-within {
		border-color: var(--amber);
		background: rgba(255, 255, 255, 0.022);
		box-shadow: 0 0 0 3px var(--border);
	}

	.prefix {
		padding: 0.85rem;
		color: var(--amber);
		font-weight: 600;
		border-right: 1px solid rgba(212, 145, 42, 0.2);
		background: var(--surface);
	}

	.input-with-prefix .input {
		flex: 1;
		padding: 0.85rem;
		border: none;
		background: transparent;
	}

	.input-with-prefix .input:focus {
		box-shadow: none;
	}

	.risk-options {
		display: grid;
		gap: 1rem;
		border: none;
		padding: 0;
		margin: 0;
	}

	.radio-label {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid var(--border-strong);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.radio-label:hover {
		background: var(--border);
		border-color: rgba(212, 145, 42, 0.3);
	}

	.radio-label input[type='radio'] {
		margin-top: 0.25rem;
		cursor: pointer;
		accent-color: var(--amber);
		width: 18px;
		height: 18px;
	}

	.radio-label input[type='radio']:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.radio-content {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.radio-title {
		font-weight: 600;
		color: var(--text);
	}

	.radio-description {
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.5);
	}

	.checkbox-label {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		padding: 1rem;
		border: 1px solid var(--border-strong);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.checkbox-label:hover {
		background: var(--border);
		border-color: rgba(212, 145, 42, 0.3);
	}

	.checkbox-label input[type='checkbox'] {
		margin-top: 0.2rem;
		cursor: pointer;
		accent-color: var(--amber);
		width: 18px;
		height: 18px;
	}

	.checkbox-label input[type='checkbox']:disabled {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.checkbox-content {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.checkbox-title {
		font-weight: 600;
		color: var(--text);
	}

	.checkbox-description {
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.5);
	}

	.form-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 1rem;
	}

	.btn-cancel,
	.btn-submit {
		padding: 0.85rem 1.8rem;
		border: none;
		border-radius: 10px;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
	}

	.btn-cancel {
		background: transparent;
		border: 1px solid rgba(212, 145, 42, 0.3);
		color: var(--amber);
	}

	.btn-cancel:hover:not(:disabled) {
		background: var(--border);
		border-color: rgba(212, 145, 42, 0.5);
	}

	.btn-submit {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		background: var(--amber);
		color: #0d0800;
	}

	.btn-submit:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.btn-cancel:disabled,
	.btn-submit:disabled {
		opacity: 0.6;
		cursor: not-allowed;
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

	@media (max-width: 768px) {
		.form-row {
			grid-template-columns: 1fr;
		}

		.form-actions {
			flex-direction: column-reverse;
		}

		.btn-cancel,
		.btn-submit {
			width: 100%;
		}
	}
</style>
