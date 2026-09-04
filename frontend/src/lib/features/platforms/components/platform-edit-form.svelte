<script lang="ts">
	import { enhance } from '$app/forms';
	import { PLATFORM_TYPES, type Platform } from '../platforms';

	let {
		platform,
		onCancel,
		onSaved
	}: {
		platform: Platform;
		onCancel: () => void;
		onSaved: () => void;
	} = $props();

	let isSubmitting = $state(false);
</script>

<form
	method="POST"
	action="?/update"
	class="platform-form"
	use:enhance={() => {
		isSubmitting = true;
		return async ({ result, update }) => {
			await update({ reset: false });
			isSubmitting = false;
			if (result.type === 'success' && result.data?.success) {
				onSaved();
			}
		};
	}}
>
	<div class="form-group">
		<label for="name" class="form-label">Nombre <span class="required">*</span></label>
		<input id="name" name="name" type="text" value={platform.name} class="form-input" required />
	</div>

	<div class="form-group">
		<label for="description" class="form-label">Descripción</label>
		<textarea id="description" name="description" class="form-textarea" rows="3"
			>{platform.description}</textarea
		>
	</div>

	<div class="form-row">
		<div class="form-group">
			<label for="type" class="form-label">Tipo <span class="required">*</span></label>
			<select id="type" name="type" class="form-select" required>
				{#each PLATFORM_TYPES.entries() as [key, label] (key)}
					<option value={key} selected={key === platform.sourceType}>{label}</option>
				{/each}
			</select>
		</div>

		<div class="form-group">
			<label for="isActive" class="form-label">Estado</label>
			<select id="isActive" name="isActive" class="form-select">
				<option value="true" selected={platform.isActive}>Activo</option>
				<option value="false" selected={!platform.isActive}>Inactivo</option>
			</select>
		</div>
	</div>

	<div class="form-actions">
		<button type="button" onclick={onCancel} class="btn btn-secondary"> Cancelar </button>
		<button type="submit" disabled={isSubmitting} class="btn btn-primary">
			{#if isSubmitting}
				<span class="spinner"></span>
				Guardando...
			{:else}
				Guardar Cambios
			{/if}
		</button>
	</div>
</form>

<style>
	.platform-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
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

	.form-input:focus,
	.form-select:focus,
	.form-textarea:focus {
		outline: none;
		border-color: var(--amber);
		box-shadow: 0 0 0 3px var(--border);
	}

	.form-textarea {
		resize: vertical;
		min-height: 90px;
	}

	.form-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 0.5rem;
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
		border: 2px solid rgba(13, 8, 0, 0.3);
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

		.btn {
			width: 100%;
		}
	}
</style>
