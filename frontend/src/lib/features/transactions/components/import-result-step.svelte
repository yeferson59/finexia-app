<script lang="ts">
	import type { ImportResult } from '../types';

	let {
		result,
		onRestart,
		onViewTransactions
	}: {
		result: ImportResult;
		onRestart: () => void;
		onViewTransactions: () => void;
	} = $props();
</script>

<div class="result-panel">
	<p class="result-icon" aria-hidden="true">{result.imported > 0 ? '✓' : '!'}</p>
	<h2 class="section-title">
		{result.imported > 0
			? `${result.imported} transacciones importadas`
			: 'No se importó ninguna transacción'}
	</h2>
	<p class="section-hint">
		{result.totalRows} filas procesadas · {result.imported} importadas · {result.skipped} omitidas por
		errores.
	</p>

	{#if result.errors.length > 0}
		<div class="result-errors">
			<h3 class="section-subtitle">Filas omitidas</h3>
			<ul>
				{#each result.errors as err (err.row)}
					<li><span class="mono">Fila {err.row}:</span> {err.message}</li>
				{/each}
			</ul>
		</div>
	{/if}

	<div class="form-actions center">
		<button type="button" class="btn btn-secondary" onclick={onRestart}>
			Importar otro archivo
		</button>
		<button type="button" class="btn btn-primary" onclick={onViewTransactions}>
			Ver mis transacciones
		</button>
	</div>
</div>

<style>
	.section-title {
		font-family: var(--font-display);
		font-size: 1.25rem;
		font-weight: 400;
		color: var(--text);
		margin: 0 0 0.4rem;
	}

	.section-subtitle {
		font-size: 0.95rem;
		font-weight: 700;
		color: var(--text);
		margin: 1.6rem 0 0.3rem;
	}

	.section-hint {
		font-size: 0.85rem;
		color: var(--text-muted);
		margin: 0 0 1rem;
	}

	.mono {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.result-panel {
		text-align: center;
		padding: 1.5rem 0.5rem;
	}

	.result-icon {
		font-size: 2.2rem;
		color: var(--amber);
		margin: 0 0 0.6rem;
	}

	.result-errors {
		text-align: left;
		max-width: 42rem;
		margin: 1.5rem auto 0;
		border: 1px solid var(--border);
		border-radius: 12px;
		padding: 1rem 1.2rem;
		max-height: 16rem;
		overflow-y: auto;
	}

	.result-errors ul {
		margin: 0.5rem 0 0;
		padding-left: 1.1rem;
		font-size: 0.82rem;
		color: var(--text-muted);
		display: grid;
		gap: 0.35rem;
	}

	.form-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 1.8rem;
	}

	.form-actions.center {
		justify-content: center;
	}

	.btn {
		padding: 0.8rem 1.4rem;
		border: none;
		border-radius: 10px;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.92rem;
		cursor: pointer;
		transition: all 0.25s ease;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-primary {
		background: var(--amber);
		color: #0d0800;
	}

	.btn-primary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.btn-secondary {
		background: transparent;
		color: var(--text);
		border: 1.5px solid rgba(212, 145, 42, 0.25);
	}

	.btn-secondary:hover:not(:disabled) {
		border-color: var(--amber);
		color: var(--amber);
	}

	@media (max-width: 768px) {
		.form-actions {
			flex-direction: column-reverse;
		}

		.btn {
			width: 100%;
			justify-content: center;
		}
	}
</style>
