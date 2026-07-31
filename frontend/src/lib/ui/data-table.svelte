<script lang="ts">
	/**
	 * Tabla de datos del dashboard: solo el chrome (scroll horizontal, cabecera
	 * mono en versalitas, filas con separador y hover).
	 *
	 * Las columnas las pone quien la usa; este componente no sabe nada del
	 * dominio. Las reglas de `th`/`td` van con `:global` porque las celdas se
	 * escriben en el componente padre y llevan su scope, no el de aquí.
	 */
	import type { Snippet } from 'svelte';

	interface Props {
		/** `<thead>` y `<tbody>` de la tabla. */
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<div class="table-wrapper">
	<table class="data-table">
		{@render children()}
	</table>
</div>

<style>
	.table-wrapper {
		overflow-x: auto;
	}

	.data-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.data-table :global(th) {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: var(--text-dim);
		padding: 0.875rem 1.25rem;
		text-align: left;
		border-bottom: 1px solid var(--border);
		white-space: nowrap;
	}

	.data-table :global(td) {
		padding: 0.75rem 1.25rem;
		color: var(--text-muted);
		border-bottom: 1px solid var(--border);
		vertical-align: middle;
	}

	.data-table :global(tbody tr:last-child td) {
		border-bottom: none;
	}

	.data-table :global(tbody tr:hover td) {
		background: var(--surface-2);
	}
</style>
