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

<form
	class="rail-fields"
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
	<div class="field">
		<label for="import-file">Archivo</label>
		<input id="import-file" type="file" name="file" accept=".csv,.xlsx,.xls" required />
		<p class="hint">{@render hint()}</p>
	</div>

	{#if error}
		<p class="feedback error" role="alert">{error}</p>
	{/if}

	<div class="actions">
		{#if onCancel}
			<Button type="button" variant="ghost" onclick={onCancel}>Cancelar</Button>
		{/if}
		<Button type="submit" loading={importing}>Importar</Button>
	</div>
</form>

{#if result}
	<div class="result">
		<p class="summary">{summarizeImport(result)}</p>
		{#if result.errors.length > 0}
			<!-- Fila a fila: un import parcial solo sirve si se sabe qué falta
			     arreglar antes de volver a subir el archivo. -->
			<ul class="errors">
				{#each result.errors as e (e.row)}
					<li><span class="row-number">Fila {e.row}</span>{e.message}</li>
				{/each}
			</ul>
		{/if}
	</div>
{/if}

<style>
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}

	/* El control de archivo nativo se pinta con la tipografía del sistema; solo
	   el botón necesita alinearse con el resto del formulario. */
	.field input[type='file'] {
		font-family: var(--font-body);
		font-size: 0.9rem;
		color: var(--text-muted);
	}

	.field input[type='file']::file-selector-button {
		margin-right: 0.9rem;
		padding: 0.55rem 0.9rem;
		border: 1px solid var(--border-strong);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.03);
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.85rem;
		cursor: pointer;
	}

	.field input[type='file']::file-selector-button:hover {
		border-color: rgba(212, 145, 42, 0.5);
	}

	.result {
		margin-top: 1.5rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--border);
	}

	.summary {
		margin: 0;
		font-size: 0.9rem;
		color: var(--text);
	}

	.errors {
		max-height: 12rem;
		margin: 0.85rem 0 0;
		padding: 0;
		overflow-y: auto;
		list-style: none;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.errors li {
		display: flex;
		gap: 0.75rem;
		padding: 0.3rem 0;
		border-top: 1px solid var(--border);
	}

	/* El número de fila es lo que se busca en la hoja: en mono y a un ancho
	   fijo, la columna se puede recorrer con la vista. */
	.row-number {
		flex-shrink: 0;
		width: 4.5rem;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		color: var(--stale);
	}
</style>
