<script lang="ts">
	/**
	 * Import masivo desde CSV/Excel, común al catálogo de activos y a las tasas
	 * de cambio: solo cambian el título, la action y las columnas que se piden.
	 *
	 * El import es parcial por diseño, así que el resultado se muestra siempre
	 * que llegue: cuenta lo que entró y detalla fila a fila lo que no.
	 */
	import type { Snippet } from 'svelte';
	import { enhance } from '$app/forms';
	import Button from '$lib/ui/button.svelte';
	import AdminFormFields from './admin-form-fields.svelte';
	import { summarizeImport, type ImportResult } from '../admin';

	interface Props {
		/** Nombre de la form action, sin el `?/`. */
		action: string;
		/** Descripción de las columnas que espera el archivo. */
		hint: Snippet;
		error?: string;
		result?: ImportResult | null;
		/**
		 * Se llama cuando el envío sale bien. La página cierra el panel desde aquí
		 * y no desde el `form` común, que también cambia con el resto de actions.
		 */
		onSuccess?: () => void;
		/** Cierra el modal sin enviar. */
		onCancel?: () => void;
	}

	let { action, hint, error = '', result = null, onSuccess, onCancel }: Props = $props();

	let importing = $state(false);
</script>

<AdminFormFields>
	<p class="import-hint">{@render hint()}</p>
	<form
		method="POST"
		action="?/{action}"
		enctype="multipart/form-data"
		use:enhance={() => {
			importing = true;
			return async ({ result, update }) => {
				importing = false;
				await update();
				if (result.type === 'success') onSuccess?.();
			};
		}}
	>
		<div class="import-row">
			<input type="file" name="file" accept=".csv,.xlsx,.xls" class="field-input" required />
			{#if onCancel}
				<Button type="button" variant="ghost" onclick={onCancel}>Cancelar</Button>
			{/if}
			<Button type="submit" loading={importing}>Importar</Button>
		</div>
		{#if error}
			<p class="form-error">{error}</p>
		{/if}
	</form>
	{#if result}
		<div class="import-result">
			<p class="import-summary">{summarizeImport(result)}</p>
			{#if result.errors.length > 0}
				<ul class="import-errors">
					{#each result.errors as e (e.row)}
						<li>Fila {e.row}: {e.message}</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/if}
</AdminFormFields>

<style>
	.import-hint {
		font-size: 0.82rem;
		color: var(--text-muted);
		margin: 0 0 1rem 0;
	}

	.import-hint :global(code) {
		font-family: var(--font-mono);
		font-size: 0.78rem;
		color: var(--amber-light);
	}

	.import-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.import-row :global(.field-input) {
		flex: 1;
	}

	.import-result {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
	}

	.import-summary {
		font-size: 0.85rem;
		color: var(--text);
		margin: 0;
	}

	.import-errors {
		margin: 0.6rem 0 0 0;
		padding-left: 1.1rem;
		font-size: 0.78rem;
		color: var(--red);
		max-height: 200px;
		overflow-y: auto;
	}

	.import-errors li {
		margin-bottom: 0.25rem;
	}
</style>
