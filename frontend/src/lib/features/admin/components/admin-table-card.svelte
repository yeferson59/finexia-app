<script lang="ts">
	/**
	 * Tarjeta de las tablas de catálogo (activos y tasas de cambio).
	 *
	 * Interna de la feature. Ambas tablas comparten el mismo tipo de celdas y el
	 * mismo editor en línea (campo numérico + «OK» por fila), así que su estilo
	 * vive aquí una sola vez, con `:global` acotado a la tarjeta.
	 */
	import type { Snippet } from 'svelte';
	import Card from '$lib/ui/card.svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<div class="table-card">
	<Card padding="none">
		{@render children()}
	</Card>
</div>

<style>
	.table-card :global(.cell-ticker) {
		font-family: var(--font-mono);
		font-weight: 600;
		color: var(--amber-light) !important;
		font-size: 0.85rem;
	}

	.table-card :global(.cell-name) {
		color: var(--text) !important;
		font-weight: 500;
	}

	.table-card :global(.cell-price) {
		font-family: var(--font-mono);
		font-weight: 600;
		color: var(--text) !important;
		white-space: nowrap;
	}

	.table-card :global(.cell-date) {
		font-family: var(--font-mono);
		font-size: 0.78rem;
		white-space: nowrap;
	}

	.table-card :global(.cell-update) {
		min-width: 160px;
	}

	.table-card :global(.update-row) {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.table-card :global(.price-input) {
		width: 90px;
		padding: 0.4rem 0.6rem;
		background: var(--surface-2);
		border: 1px solid var(--border-strong);
		border-radius: 6px;
		color: var(--text);
		font-family: var(--font-mono);
		font-size: 0.82rem;
		transition: border-color 0.2s ease;
	}

	.table-card :global(.price-input:focus) {
		outline: none;
		border-color: var(--amber);
	}

	.table-card :global(.price-input.input-error) {
		border-color: var(--red);
	}

	.table-card :global(.row-success td) {
		background: rgba(76, 175, 80, 0.05) !important;
	}

	.table-card :global(.row-error) {
		font-size: 0.75rem;
		color: var(--red);
		margin: 0.25rem 0 0 0;
	}

	.table-card :global(.empty-state) {
		text-align: center;
		padding: 3rem;
		color: var(--text-dim);
		font-size: 0.9rem;
	}
</style>
