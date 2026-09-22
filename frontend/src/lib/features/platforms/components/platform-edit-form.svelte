<script lang="ts">
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import { optimisticSubmit } from '$lib/shared/optimistic.svelte';
	import { PLATFORM_TYPES, type Platform } from '../platforms';

	let {
		platform,
		onApply,
		onRejected,
		onCancel,
		onSaved
	}: {
		platform: Platform;
		/**
		 * Pinta los cambios en la ficha al pulsar, sin esperar al servidor, y
		 * oculta el diálogo sin desmontarlo. Devuelve cómo deshacerlo.
		 */
		onApply: (changes: Partial<Platform>) => () => void;
		/** Rechazado: el diálogo vuelve con lo escrito y el motivo. */
		onRejected: () => void;
		onCancel: () => void;
		onSaved: () => void;
	} = $props();

	let isSubmitting = $state(false);
	// Antes la action devolvía su motivo y nadie lo leía: un cambio rechazado
	// cerraba nada y no decía nada.
	let error = $state('');

	const submit = optimisticSubmit({
		fallbackError: 'Error al actualizar la plataforma',
		apply: (formData) => {
			isSubmitting = true;
			error = '';
			return onApply({
				name: String(formData.get('name') ?? '').trim(),
				description: String(formData.get('description') ?? '').trim(),
				sourceType: String(formData.get('type') ?? platform.sourceType),
				isActive: formData.get('isActive') !== 'false'
			});
		},
		onError: (message) => {
			isSubmitting = false;
			error = message;
			onRejected();
		},
		onSuccess: () => {
			isSubmitting = false;
			onSaved();
		}
	});
</script>

<form method="POST" action="?/update" class="platform-form" use:enhance={submit}>
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

	{#if error}
		<p class="feedback error" role="alert">{error}</p>
	{/if}

	<div class="modal-actions">
		<Button type="button" variant="ghost" onclick={onCancel} disabled={isSubmitting}
			>Cancelar</Button
		>
		<Button type="submit" loading={isSubmitting}>Guardar cambios</Button>
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

	@media (max-width: 768px) {
		.form-row {
			grid-template-columns: 1fr;
		}
	}
</style>
