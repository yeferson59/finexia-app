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
	import { cn } from '$lib/shared/css';

	interface Props {
		/**
		 * Qué contiene la tabla, para el lector de pantalla. Se pinta como
		 * `<caption>` oculto a la vista: una tabla sin título obliga a deducirlo
		 * de las cabeceras.
		 */
		caption?: string;
		/** Muestra el `caption` también a la vista. */
		showCaption?: boolean;
		/**
		 * Mantiene la cabecera fija al hacer scroll vertical. Para listas largas
		 * dentro de un contenedor con alto máximo.
		 */
		stickyHeader?: boolean;
		class?: string;
		/** `<thead>` y `<tbody>` de la tabla. */
		children: Snippet;
	}

	let {
		caption = '',
		showCaption = false,
		stickyHeader = false,
		class: className = '',
		children
	}: Props = $props();
</script>

<div class={cn('table-wrapper', className)}>
	<table class={cn('data-table', { 'sticky-header': stickyHeader })}>
		{#if caption}
			<caption class={showCaption ? 'caption' : 'caption sr-only'}>{caption}</caption>
		{/if}
		{@render children()}
	</table>
</div>

<style>
	.table-wrapper {
		overflow-x: auto;
		/*
		 * Sombra al borde solo mientras haya contenido cortado: el degradado va
		 * pegado al scroll y el "tapón" del mismo color lo cubre en los extremos,
		 * así que aparece y desaparece sin JavaScript.
		 */
		background:
			linear-gradient(to right, var(--bg) 30%, rgba(8, 9, 10, 0)) left / 28px 100% no-repeat,
			linear-gradient(to left, var(--bg) 30%, rgba(8, 9, 10, 0)) right / 28px 100% no-repeat,
			radial-gradient(farthest-side at 0 50%, rgba(0, 0, 0, 0.45), transparent) left / 12px 100%
				no-repeat,
			radial-gradient(farthest-side at 100% 50%, rgba(0, 0, 0, 0.45), transparent) right / 12px 100%
				no-repeat;
		background-attachment: local, local, scroll, scroll;
	}

	.data-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.caption {
		text-align: left;
		font-size: 0.75rem;
		color: var(--text-dim);
		padding: 0 1.25rem 0.75rem;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
		border: 0;
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

	.sticky-header :global(thead th) {
		position: sticky;
		top: 0;
		z-index: 1;
		background: #0d0e10;
	}

	.data-table :global(td) {
		padding: 0.75rem 1.25rem;
		color: var(--text-muted);
		border-bottom: 1px solid var(--border);
		vertical-align: middle;
	}

	/* Importes alineados a la derecha y con cifras de ancho fijo, para poder
	   comparar columnas de un vistazo. */
	.data-table :global(th.num),
	.data-table :global(td.num) {
		text-align: right;
		font-variant-numeric: tabular-nums;
	}

	.data-table :global(td.num) {
		font-family: var(--font-mono);
	}

	.data-table :global(tbody tr:nth-child(even) td) {
		background: rgba(255, 255, 255, 0.014);
	}

	.data-table :global(tbody tr:last-child td) {
		border-bottom: none;
	}

	.data-table :global(tbody tr:hover td) {
		background: var(--surface-2);
	}
</style>
