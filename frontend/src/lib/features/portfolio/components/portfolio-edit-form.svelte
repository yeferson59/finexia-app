<script lang="ts">
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import Card from '$lib/ui/card.svelte';
	import { PORTFOLIO_TYPES } from '../portfolio';

	interface EditablePortfolio {
		name: string;
		description?: string | null;
		type: string;
		riskId?: string;
		isDefault: boolean;
	}

	let {
		portfolio,
		risks,
		onCancel,
		onSaved,
		onError
	}: {
		portfolio: EditablePortfolio;
		risks: { id: string; name: string }[];
		onCancel: () => void;
		onSaved: () => void;
		onError: (message: string) => void;
	} = $props();

	let isSubmitting = $state(false);

	let editName = $state('');
	let editDescription = $state('');
	let editType = $state('');
	let editRiskId = $state('');
	let editIsDefault = $state(false);

	$effect(() => {
		if (portfolio) {
			untrack(() => {
				editName = portfolio.name;
				editDescription = portfolio.description ?? '';
				editType = portfolio.type;
				editRiskId = portfolio.riskId ?? '';
				editIsDefault = portfolio.isDefault;
			});
		}
	});
</script>

<Card variant="elevated" padding="sm">
	<form
		method="POST"
		action="?/updatePortfolio"
		class="edit-form"
		use:enhance={() => {
			isSubmitting = true;
			return async ({ result, update }) => {
				if (result.type === 'success') {
					onSaved();
				} else if (result.type === 'failure') {
					onError(
						(result.data as { error?: string })?.error ?? 'Error al actualizar el portafolio.'
					);
				}
				await update({ reset: false });
				isSubmitting = false;
			};
		}}
	>
		<h3 class="edit-title">Editar portafolio</h3>

		<div class="form-group">
			<label for="edit-name">Nombre</label>
			<input id="edit-name" name="name" type="text" bind:value={editName} required minlength="2" />
		</div>

		<div class="form-group">
			<label for="edit-description">Descripción</label>
			<textarea id="edit-description" name="description" rows="2" bind:value={editDescription}
			></textarea>
		</div>

		<div class="form-row">
			<div class="form-group">
				<label for="edit-type">Tipo</label>
				<select id="edit-type" name="type" bind:value={editType}>
					{#each PORTFOLIO_TYPES as pt (pt.value)}
						<option value={pt.value}>{pt.label}</option>
					{/each}
				</select>
			</div>

			<div class="form-group">
				<label for="edit-risk">Nivel de riesgo</label>
				<select id="edit-risk" name="riskId" bind:value={editRiskId}>
					{#each risks as risk (risk.id)}
						<option value={risk.id}>{risk.name}</option>
					{/each}
				</select>
			</div>
		</div>

		<div class="form-check">
			<input
				id="edit-default"
				name="isDefault"
				type="checkbox"
				bind:checked={editIsDefault}
				value="true"
			/>
			<label for="edit-default">Portafolio por defecto</label>
		</div>

		<div class="form-actions">
			<button type="button" class="btn-cancel" onclick={onCancel}>Cancelar</button>
			<button type="submit" class="btn-save" disabled={isSubmitting}>
				{isSubmitting ? 'Guardando…' : 'Guardar cambios'}
			</button>
		</div>
	</form>
</Card>

<style>
	.edit-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.edit-title {
		margin: 0 0 0.25rem;
		font-size: 1rem;
		font-weight: 600;
		color: var(--text);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.form-group label {
		font-size: 0.8rem;
		font-weight: 600;
		letter-spacing: 0.4px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.55);
	}

	.form-group input,
	.form-group textarea,
	.form-group select {
		padding: 0.7rem 0.9rem;
		border: 1px solid rgba(212, 145, 42, 0.25);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.04);
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.9rem;
		transition: border-color 0.2s ease;
	}

	.form-group input:focus,
	.form-group textarea:focus,
	.form-group select:focus {
		outline: none;
		border-color: rgba(212, 145, 42, 0.6);
	}

	.form-group textarea {
		resize: vertical;
	}

	.form-group select option {
		background: #1a1209;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.form-check {
		display: flex;
		align-items: center;
		gap: 0.65rem;
	}

	.form-check input[type='checkbox'] {
		width: 16px;
		height: 16px;
		cursor: pointer;
		accent-color: var(--amber);
	}

	.form-check label {
		font-size: 0.9rem;
		color: var(--text);
		cursor: pointer;
	}

	.form-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
		padding-top: 0.25rem;
	}

	.btn-cancel {
		padding: 0.7rem 1.25rem;
		border: 1px solid rgba(236, 234, 229, 0.2);
		border-radius: 8px;
		background: transparent;
		color: rgba(236, 234, 229, 0.7);
		font-family: var(--font-body);
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-cancel:hover {
		border-color: rgba(236, 234, 229, 0.4);
		color: var(--text);
	}

	.btn-save {
		padding: 0.7rem 1.5rem;
		border: none;
		border-radius: 8px;
		background: var(--amber);
		color: #0d0800;
		font-family: var(--font-body);
		font-weight: 700;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.btn-save:hover:not(:disabled) {
		box-shadow: 0 6px 18px rgba(212, 145, 42, 0.3);
	}

	.btn-save:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
</style>
