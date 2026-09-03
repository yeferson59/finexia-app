<script lang="ts">
	import type { ImportMapping, ImportPreview, ImportDefaults } from '../types';
	import ImportPreviewTable from './import-preview-table.svelte';
	import { CATEGORY_OPTIONS, TXN_TYPE_OPTIONS } from '../transactions';

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
		{ key: 'fxRate', label: 'Tasa de cambio', required: false },
		{ key: 'category', label: 'Categoría', required: false },
		{ key: 'notes', label: 'Notas', required: false }
	];

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
			{#each TXN_TYPE_OPTIONS as t (t.value)}
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
		<label class="form-label" for="default-cost-currency">Moneda de la cuenta</label>
		<input
			id="default-cost-currency"
			class="form-input"
			type="text"
			maxlength="3"
			bind:value={defaults.costCurrency}
			onchange={onRefreshDefaults}
			placeholder="igual que la fila"
		/>
		<p class="field-hint">
			En la que tu bróker debitó. Déjala vacía si el extracto no convirtió nada; si la rellenas,
			cada fila en otra moneda necesita su tasa.
		</p>
	</div>
	<div class="form-group">
		<label class="form-label" for="default-category">Categoría</label>
		<select
			id="default-category"
			class="form-select"
			bind:value={defaults.category}
			onchange={onRefreshDefaults}
		>
			{#each CATEGORY_OPTIONS as c (c.value)}
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

<ImportPreviewTable {preview} {loading} />

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

	.field-hint {
		margin: 0.35rem 0 0;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.45);
		font-style: italic;
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
