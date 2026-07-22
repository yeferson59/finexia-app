<script lang="ts">
	import type { ImportMapping, ImportPreview, ImportDefaults } from '../types';

	let {
		preview,
		fileName,
		sheet,
		mapping,
		defaults = $bindable(),
		loading,
		importing,
		canImport,
		onChangeSheet,
		onSetMappingColumn,
		onRefreshDefaults,
		onRestart,
		onImport
	}: {
		preview: ImportPreview;
		fileName: string | undefined;
		sheet: string;
		mapping: ImportMapping;
		defaults: ImportDefaults;
		loading: boolean;
		importing: boolean;
		canImport: boolean;
		onChangeSheet: (value: string) => void;
		onSetMappingColumn: (key: keyof ImportMapping, value: string) => void;
		onRefreshDefaults: () => void;
		onRestart: () => void;
		onImport: () => void;
	} = $props();

	const mappingFields: { key: keyof ImportMapping; label: string; required: boolean }[] = [
		{ key: 'date', label: 'Fecha', required: true },
		{ key: 'ticker', label: 'Ticker / Símbolo', required: true },
		{ key: 'quantity', label: 'Cantidad', required: true },
		{ key: 'price', label: 'Precio', required: true },
		{ key: 'type', label: 'Tipo de operación', required: false },
		{ key: 'assetName', label: 'Nombre del activo', required: false },
		{ key: 'fees', label: 'Comisiones', required: false },
		{ key: 'currency', label: 'Moneda', required: false },
		{ key: 'category', label: 'Categoría', required: false },
		{ key: 'notes', label: 'Notas', required: false }
	];

	const txnTypeOptions = [
		{ value: 'buy', label: 'Compra' },
		{ value: 'sell', label: 'Venta' },
		{ value: 'dividend', label: 'Dividendo' },
		{ value: 'interest', label: 'Interés' },
		{ value: 'transfer_in', label: 'Transferencia entrada' },
		{ value: 'transfer_out', label: 'Transferencia salida' },
		{ value: 'fee', label: 'Cargo' },
		{ value: 'split', label: 'División' }
	];

	const categoryOptions = [
		{ value: 'stock', label: 'Acciones' },
		{ value: 'etf', label: 'ETFs' },
		{ value: 'crypto', label: 'Criptomonedas' },
		{ value: 'bond', label: 'Bonos' },
		{ value: 'cash', label: 'Efectivo' },
		{ value: 'real_estate', label: 'Bienes raíces' },
		{ value: 'commodity', label: 'Materias primas' },
		{ value: 'other', label: 'Otros' }
	];

	const typeLabels: Record<string, string> = Object.fromEntries(
		txnTypeOptions.map((t) => [t.value, t.label])
	);

	const fieldLabels: Record<string, string> = {
		date: 'Fecha',
		ticker: 'Ticker',
		quantity: 'Cantidad',
		price: 'Precio'
	};

	function columnLabel(index: number): string {
		let label = '';
		let n = index;
		do {
			label = String.fromCharCode(65 + (n % 26)) + label;
			n = Math.floor(n / 26) - 1;
		} while (n >= 0);
		return label;
	}
</script>

<div class="map-header">
	<div>
		<h2 class="section-title">Asigna tus columnas</h2>
		<p class="section-hint">
			Detectamos <strong>{preview.headers.length}</strong> columnas en
			<strong>{fileName}</strong> (encabezados en la fila {preview.headerRow}). Revisa la asignación
			sugerida y ajústala a tu formato.
		</p>
	</div>
	{#if preview.sheets.length > 1}
		<div class="form-group sheet-select">
			<label class="form-label" for="sheet">Hoja</label>
			<select
				id="sheet"
				class="form-select"
				value={sheet}
				onchange={(e) => onChangeSheet(e.currentTarget.value)}
			>
				{#each preview.sheets as s (s)}
					<option value={s}>{s}</option>
				{/each}
			</select>
		</div>
	{/if}
</div>

{#if preview.missingFields.length > 0}
	<p class="warning-banner" role="alert">
		Faltan columnas obligatorias por asignar:
		<strong>{preview.missingFields.map((f) => fieldLabels[f] ?? f).join(', ')}</strong>.
	</p>
{/if}

<div class="mapping-grid">
	{#each mappingFields as field (field.key)}
		<div class="form-group">
			<label class="form-label" for={`map-${field.key}`}>
				{field.label}
				{#if field.required}<span class="required">*</span>{/if}
			</label>
			<select
				id={`map-${field.key}`}
				class="form-select"
				value={mapping[field.key] === null ? '' : String(mapping[field.key])}
				onchange={(e) => onSetMappingColumn(field.key, e.currentTarget.value)}
			>
				<option value="">— No usar —</option>
				{#each preview.headers as header, i (i)}
					<option value={String(i)}>
						{columnLabel(i)} · {header || '(sin título)'}
					</option>
				{/each}
			</select>
		</div>
	{/each}
</div>

<h3 class="section-subtitle">Valores por defecto</h3>
<p class="section-hint">Se aplican a las filas donde tu archivo no tenga ese dato.</p>
<div class="defaults-grid">
	<div class="form-group">
		<label class="form-label" for="default-type">Tipo de operación</label>
		<select
			id="default-type"
			class="form-select"
			bind:value={defaults.type}
			onchange={onRefreshDefaults}
		>
			{#each txnTypeOptions as t (t.value)}
				<option value={t.value}>{t.label}</option>
			{/each}
		</select>
	</div>
	<div class="form-group">
		<label class="form-label" for="default-currency">Moneda</label>
		<input
			id="default-currency"
			class="form-input"
			type="text"
			maxlength="3"
			bind:value={defaults.currency}
			onchange={onRefreshDefaults}
			placeholder="USD"
		/>
	</div>
	<div class="form-group">
		<label class="form-label" for="default-category">Categoría</label>
		<select
			id="default-category"
			class="form-select"
			bind:value={defaults.category}
			onchange={onRefreshDefaults}
		>
			{#each categoryOptions as c (c.value)}
				<option value={c.value}>{c.label}</option>
			{/each}
		</select>
	</div>
	<div class="form-group">
		<label class="form-label" for="default-dates">Formato de fecha</label>
		<select
			id="default-dates"
			class="form-select"
			bind:value={defaults.dateFormat}
			onchange={onRefreshDefaults}
		>
			<option value="auto">Detectar automáticamente</option>
			<option value="dmy">Día/Mes/Año</option>
			<option value="mdy">Mes/Día/Año</option>
		</select>
	</div>
</div>

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
					<td>{typeLabels[row.type] ?? row.type ?? '—'}</td>
					<td class="mono">{row.ticker || '—'}</td>
					<td class="mono">{row.quantity || '—'}</td>
					<td class="mono">{row.price || '—'}</td>
					<td class="mono">{row.fees || '—'}</td>
					<td class="mono">{row.currency || '—'}</td>
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

<div class="form-actions">
	<button type="button" class="btn btn-secondary" onclick={onRestart} disabled={importing}>
		Elegir otro archivo
	</button>
	<button type="button" class="btn btn-primary" onclick={onImport} disabled={!canImport}>
		{#if importing}
			<span class="spinner dark"></span> Importando…
		{:else}
			Importar {preview.validRows} transacciones
		{/if}
	</button>
</div>

<style>
	.map-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1.5rem;
		flex-wrap: wrap;
	}

	.sheet-select {
		min-width: 180px;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.form-label {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text);
		letter-spacing: 0.3px;
	}

	.required {
		color: var(--red, #e05a5a);
	}

	.form-input,
	.form-select {
		padding: 0.7rem 0.9rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		color: var(--text);
		font-size: 0.9rem;
		font-family: var(--font-body);
		transition: border-color 0.2s ease;
	}

	.form-input:focus,
	.form-select:focus {
		outline: none;
		border-color: var(--amber);
		box-shadow: 0 0 0 3px var(--border);
	}

	.form-select option {
		background: #1a1611;
		color: var(--text);
	}

	.warning-banner {
		border-radius: 10px;
		padding: 0.8rem 1rem;
		font-size: 0.85rem;
		margin-bottom: 1.2rem;
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.4);
		color: var(--amber);
	}

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

	.mapping-grid,
	.defaults-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 1rem;
		margin-top: 1rem;
	}

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

	.form-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 1.8rem;
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

	.btn-primary:disabled {
		opacity: 0.55;
		cursor: not-allowed;
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

	.spinner {
		display: inline-block;
		width: 14px;
		height: 14px;
		border: 2px solid rgba(212, 145, 42, 0.3);
		border-top-color: var(--amber);
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	.spinner.dark {
		border-color: rgba(13, 8, 0, 0.25);
		border-top-color: #0d0800;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
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
