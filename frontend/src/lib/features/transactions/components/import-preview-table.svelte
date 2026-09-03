<script lang="ts">
	import type { ImportPreview } from '../types';
	import { TXN_TYPE_LABELS } from '../transactions';

	let { preview, loading }: { preview: ImportPreview; loading: boolean } = $props();
</script>

<div class="preview-summary" aria-live="polite">
	{#if loading}
		<span class="spinner"></span> Actualizando vista previa…
	{:else}
		<span class="count total">{preview.totalRows} filas</span>
		<span class="count ok">{preview.validRows} listas para importar</span>
		{#if preview.invalidRows > 0}
			<span class="count bad">{preview.invalidRows} con errores (se omitirán)</span>
		{/if}
	{/if}
</div>

<div class="preview-table-wrap">
	<table class="preview-table">
		<thead>
			<tr>
				<th>Fila</th>
				<th>Estado</th>
				<th>Fecha</th>
				<th>Tipo</th>
				<th>Ticker</th>
				<th>Cantidad</th>
				<th>Precio</th>
				<th>Comisión</th>
				<th>Moneda</th>
				<th>Tasa</th>
				<th>Cuenta</th>
				<th>Detalle</th>
			</tr>
		</thead>
		<tbody>
			{#each preview.rows as row (row.rowNumber)}
				<tr class:invalid={!row.valid}>
					<td class="mono">{row.rowNumber}</td>
					<td>
						{#if row.valid}
							<span class="status ok">✓</span>
						{:else}
							<span class="status bad">✗</span>
						{/if}
					</td>
					<td class="mono">{row.date || '—'}</td>
					<td>{TXN_TYPE_LABELS[row.type] ?? row.type ?? '—'}</td>
					<td class="mono">{row.ticker || '—'}</td>
					<td class="mono">{row.quantity || '—'}</td>
					<td class="mono">{row.price || '—'}</td>
					<td class="mono">{row.fees || '—'}</td>
					<td class="mono">{row.currency || '—'}</td>
					<!-- La tasa y la moneda de la cuenta sin destacar cuando no hubo
					     conversión: en un extracto de una sola moneda son 1 y la misma
					     de la fila en cada línea, y llamar la atención sobre ellas
					     sería ruido en el caso normal. -->
					<td class="mono" class:muted={row.fxRate === '1'}>{row.fxRate || '—'}</td>
					<td class="mono" class:muted={row.costCurrency === row.currency}>
						{row.costCurrency || '—'}
					</td>
					<td class="errors-cell">{row.errors.join('; ')}</td>
				</tr>
			{/each}
		</tbody>
	</table>
	{#if preview.totalRows > preview.rows.length}
		<p class="table-note">
			Mostrando las primeras {preview.rows.length} filas de {preview.totalRows}. El total se valida
			completo al importar.
		</p>
	{/if}
</div>

<style>
	.preview-summary {
		display: flex;
		align-items: center;
		gap: 0.8rem;
		flex-wrap: wrap;
		margin: 1.6rem 0 0.8rem;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.count {
		padding: 0.3rem 0.7rem;
		border-radius: 999px;
		font-weight: 600;
		font-size: 0.8rem;
	}

	.count.total {
		background: rgba(255, 255, 255, 0.06);
		color: var(--text);
	}

	.count.ok {
		background: rgba(34, 201, 126, 0.12);
		color: #22c97e;
	}

	.count.bad {
		background: rgba(224, 90, 90, 0.12);
		color: #e05a5a;
	}

	.preview-table-wrap {
		overflow-x: auto;
		max-height: 26rem;
		overflow-y: auto;
		border: 1px solid var(--border);
		border-radius: 12px;
	}

	.preview-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.82rem;
	}

	.preview-table th {
		position: sticky;
		top: 0;
		background: #1f1a12;
		color: rgba(236, 234, 229, 0.75);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-size: 0.7rem;
		text-align: left;
		padding: 0.7rem 0.8rem;
		z-index: 1;
	}

	.preview-table td {
		padding: 0.55rem 0.8rem;
		border-top: 1px solid rgba(255, 255, 255, 0.05);
		color: var(--text);
		white-space: nowrap;
	}

	.preview-table tr.invalid td {
		background: rgba(224, 90, 90, 0.05);
	}

	.muted {
		color: rgba(236, 234, 229, 0.35);
	}

	.errors-cell {
		color: #e05a5a;
		font-size: 0.78rem;
		max-width: 26rem;
		white-space: normal;
	}

	.status.ok {
		color: #22c97e;
		font-weight: 700;
	}

	.status.bad {
		color: #e05a5a;
		font-weight: 700;
	}

	.table-note {
		font-size: 0.78rem;
		color: var(--text-muted);
		padding: 0.6rem 0.8rem;
		margin: 0;
	}

	.mono {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.spinner {
		display: inline-block;
		width: 14px;
		height: 14px;
		border: 2px solid rgba(212, 145, 42, 0.3);
		border-top-color: var(--amber);
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
