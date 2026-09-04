<script lang="ts">
	/**
	 * Estilo de los campos que comparten el alta y el import de las pantallas de
	 * catálogo (activos y tasas), con `:global` acotado a este envoltorio.
	 *
	 * Interna de la feature. Aportaba también el título y la tarjeta; los pone
	 * ahora el `Modal` que aloja el formulario, que es quien tiene el encabezado.
	 */
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<div class="create-form-wrap">
	{@render children()}
</div>

<style>
	.create-form-wrap :global(.form-grid) {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.create-form-wrap :global(.form-field) {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.create-form-wrap :global(.field-label) {
		font-family: var(--font-mono);
		font-size: 0.65rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--text-dim);
	}

	.create-form-wrap :global(.required) {
		color: var(--red);
	}

	.create-form-wrap :global(.optional) {
		color: var(--text-dim);
		font-weight: 400;
	}

	.create-form-wrap :global(.field-input) {
		padding: 0.55rem 0.75rem;
		background: var(--surface-2);
		border: 1px solid var(--border-strong);
		border-radius: 6px;
		color: var(--text);
		font-size: 0.875rem;
		font-family: var(--font-body);
		transition: border-color 0.2s ease;
	}

	.create-form-wrap :global(.field-input:focus) {
		outline: none;
		border-color: var(--amber);
	}

	.create-form-wrap :global(.field-select) {
		cursor: pointer;
		appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23888' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.6rem center;
		padding-right: 2rem;
	}

	.create-form-wrap :global(.field-select option) {
		background: var(--bg);
		color: var(--text);
	}

	.create-form-wrap :global(.form-error) {
		font-size: 0.82rem;
		color: var(--red);
		margin: 0 0 0.75rem 0;
	}

	.create-form-wrap :global(.form-actions) {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
	}
</style>
