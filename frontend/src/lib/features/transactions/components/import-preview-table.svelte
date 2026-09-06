<script lang="ts">
	/*
	 * Las primeras filas del archivo, ya interpretadas.
	 *
	 * Tres cosas de la versión anterior:
	 *
	 *  - Doce columnas fijas, y cuatro de ellas —comisión, moneda, tasa y moneda
	 *    de la cuenta— salían llenas de «—» en la importación normal, que es un
	 *    extracto en una sola moneda y sin comisión por línea. Ahora una columna
	 *    solo aparece si alguna fila trae ese dato.
	 *  - Los errores iban en una celda al final, o sea fuera de la pantalla,
	 *    detrás de un scroll horizontal, que es donde no sirven. Ahora van debajo
	 *    de su fila, a todo lo ancho.
	 *  - La cabecera era una barra oscura en versalitas; es la misma tabla de un
	 *    pelo que el libro de movimientos.
	 */
	import type { ImportPreview } from '../types';
	import { TXN_TYPE_LABELS } from '../transactions';

	let { preview, loading }: { preview: ImportPreview; loading: boolean } = $props();

	const OPTIONAL_COLUMNS = [
		{ key: 'fees', label: 'Comisión' },
		{ key: 'currency', label: 'Moneda' },
		{ key: 'fxRate', label: 'Tasa' },
		{ key: 'costCurrency', label: 'Cuenta' }
	] as const;

	const shownColumns = $derived(
		OPTIONAL_COLUMNS.filter((column) => preview.rows.some((row) => row[column.key]))
	);

	/* Fila, fecha, tipo, ticker, cantidad y precio van siempre. */
	const columnCount = $derived(6 + shownColumns.length);
</script>

<p class="summary" aria-live="polite">
	{#if loading}
		<span class="spinner"></span> Actualizando la vista previa…
	{:else}
		{preview.totalRows} filas: {preview.validRows} listas para importar{#if preview.invalidRows > 0}
			y {preview.invalidRows} con errores, que se quedan fuera{/if}.
	{/if}
</p>

<div class="scroll">
	<table>
		<caption class="sr-only">
			Las primeras filas de tu archivo tal como quedarán registradas, con el motivo por el que se
			omitirán las que no se puedan interpretar
		</caption>
		<thead>
			<tr>
				<th scope="col" class="num">Fila</th>
				<th scope="col">Fecha</th>
				<th scope="col">Movimiento</th>
				<th scope="col">Ticker</th>
				<th scope="col" class="num">Cantidad</th>
				<th scope="col" class="num">Precio</th>
				{#each shownColumns as column (column.key)}
					<th scope="col" class="num">{column.label}</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each preview.rows as row (row.rowNumber)}
				<tr class:invalid={!row.valid}>
					<th scope="row" class="num figure line">{row.rowNumber}</th>
					<td class="figure">{row.date || '—'}</td>
					<td class="kind">{TXN_TYPE_LABELS[row.type] ?? row.type ?? '—'}</td>
					<td class="figure ticker">{row.ticker || '—'}</td>
					<td class="num figure">{row.quantity || '—'}</td>
					<td class="num figure">{row.price || '—'}</td>
					{#each shownColumns as column (column.key)}
						<!-- La tasa y la moneda de la cuenta sin destacar cuando no hubo
						     conversión: en un extracto de una sola moneda son 1 y la misma
						     de la fila en cada línea, y llamar la atención sobre ellas
						     sería ruido en el caso normal. -->
						<td
							class="num figure"
							class:muted={(column.key === 'fxRate' && row.fxRate === '1') ||
								(column.key === 'costCurrency' && row.costCurrency === row.currency)}
						>
							{row[column.key] || '—'}
						</td>
					{/each}
				</tr>
				{#if !row.valid && row.errors.length > 0}
					<tr class="why">
						<td colspan={columnCount}>{row.errors.join('. ')}</td>
					</tr>
				{/if}
			{/each}
		</tbody>
	</table>
</div>

{#if preview.totalRows > preview.rows.length}
	<p class="note">
		Son las primeras {preview.rows.length} de {preview.totalRows} filas. Al importar se revisan todas.
	</p>
{/if}

<style>
	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	.summary {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin: 0 0 1.25rem;
		font-size: 0.85rem;
		color: var(--text-muted);
	}

	.scroll {
		overflow-x: auto;
		overscroll-behavior-x: contain;
	}

	table {
		width: 100%;
		min-width: 42rem;
		border-collapse: collapse;
	}

	thead th {
		padding: 0 0.75rem 0.6rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
		text-align: left;
		white-space: nowrap;
	}

	tbody th,
	tbody td {
		padding: 0.7rem 0.75rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.82rem;
		font-weight: 400;
		color: var(--text);
		text-align: left;
		white-space: nowrap;
	}

	thead th:first-child,
	tbody th:first-child {
		padding-left: 0;
	}

	thead th:last-child,
	tbody td:last-child {
		padding-right: 0;
	}

	.num {
		text-align: right;
	}

	/* El número de fila es la referencia con la que el usuario vuelve a su hoja:
	   en tono bajo, pero legible. */
	.line {
		color: var(--text-dim);
	}

	.kind {
		white-space: nowrap;
	}

	.ticker {
		letter-spacing: 0.03em;
	}

	.muted {
		color: var(--text-dim);
	}

	/* Lo que no entra se ve de un vistazo, y el motivo va justo debajo y entero,
	   no en una celda al final de un scroll horizontal. */
	tbody tr.invalid th,
	tbody tr.invalid td {
		background: rgba(224, 90, 90, 0.045);
		border-bottom-color: transparent;
	}

	tbody tr.invalid .line {
		color: var(--red);
	}

	.why td {
		padding-top: 0;
		padding-bottom: 0.7rem;
		background: rgba(224, 90, 90, 0.045);
		font-size: 0.79rem;
		line-height: 1.45;
		color: var(--red);
		white-space: normal;
	}

	.note {
		margin: 0.9rem 0 0;
		font-size: 0.8rem;
		color: var(--text-dim);
	}
</style>
