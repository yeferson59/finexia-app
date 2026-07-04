<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	interface ImportMapping {
		date: number | null;
		type: number | null;
		ticker: number | null;
		assetName: number | null;
		quantity: number | null;
		price: number | null;
		fees: number | null;
		currency: number | null;
		category: number | null;
		notes: number | null;
	}

	interface ImportRow {
		rowNumber: number;
		raw: string[];
		date: string;
		type: string;
		ticker: string;
		assetName: string;
		quantity: string;
		price: string;
		fees: string;
		currency: string;
		category: string;
		notes: string;
		valid: boolean;
		errors: string[];
	}

	interface ImportPreview {
		sheets: string[];
		sheet: string;
		headerRow: number;
		headers: string[];
		suggestedMapping: ImportMapping;
		missingFields: string[];
		totalRows: number;
		validRows: number;
		invalidRows: number;
		rows: ImportRow[];
	}

	interface ImportResult {
		totalRows: number;
		imported: number;
		skipped: number;
		errors: { row: number; message: string }[];
	}

	type Step = 'upload' | 'map' | 'done';

	const emptyMapping: ImportMapping = {
		date: null,
		type: null,
		ticker: null,
		assetName: null,
		quantity: null,
		price: null,
		fees: null,
		currency: null,
		category: null,
		notes: null
	};

	let step: Step = $state('upload');
	let file: File | null = $state(null);
	// The load data only seeds the initial selection; the user owns it afterwards.
	// svelte-ignore state_referenced_locally
	let portfolioId = $state(
		data.portfolios.find((p) => p.isDefault)?.id ?? data.portfolios[0]?.id ?? ''
	);
	// svelte-ignore state_referenced_locally
	let sourceId = $state(data.platforms[0]?.id ?? '');
	let sheet = $state('');
	let preview: ImportPreview | null = $state(null);
	let result: ImportResult | null = $state(null);
	let mapping: ImportMapping = $state({ ...emptyMapping });
	let defaults = $state({ type: 'buy', currency: 'USD', category: 'stock', dateFormat: 'auto' });
	let loading = $state(false);
	let importing = $state(false);
	let errorMsg = $state('');
	let dragOver = $state(false);
	let fileInput: HTMLInputElement | null = $state(null);

	const canImport = $derived.by(() => {
		if (!preview || loading || importing) return false;
		return (
			preview.validRows > 0 && preview.missingFields.length === 0 && !!portfolioId && !!sourceId
		);
	});

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

	function selectFile(candidate: File | undefined | null) {
		errorMsg = '';
		if (!candidate) return;
		const name = candidate.name.toLowerCase();
		if (!name.endsWith('.xlsx') && !name.endsWith('.csv')) {
			errorMsg = 'Formato no soportado. Sube un archivo .xlsx o .csv.';
			return;
		}
		if (candidate.size > 8 * 1024 * 1024) {
			errorMsg = 'El archivo supera el tamaño máximo de 8 MB.';
			return;
		}
		file = candidate;
		sheet = '';
		void requestPreview(false);
	}

	async function requestPreview(withMapping: boolean) {
		if (!file) return;
		loading = true;
		errorMsg = '';
		try {
			const form = new FormData();
			form.append('file', file);
			if (sheet) form.append('sheet', sheet);
			if (withMapping) form.append('mapping', JSON.stringify(mapping));
			form.append('defaults', JSON.stringify(defaults));

			const res = await fetch(resolve('/dashboard/transactions/import/preview'), {
				method: 'POST',
				body: form
			});
			const body = await res.json();
			if (!res.ok || !body.success) {
				errorMsg = body?.details || body?.message || 'No se pudo leer el archivo.';
				if (step === 'upload') file = null;
				return;
			}
			preview = body.data as ImportPreview;
			sheet = preview.sheet;
			if (!withMapping) {
				mapping = { ...preview.suggestedMapping };
			}
			step = 'map';
		} catch {
			errorMsg = 'Error de conexión al procesar el archivo.';
			if (step === 'upload') file = null;
		} finally {
			loading = false;
		}
	}

	function setMappingColumn(key: keyof ImportMapping, value: string) {
		mapping[key] = value === '' ? null : Number(value);
		void requestPreview(true);
	}

	function changeSheet(value: string) {
		sheet = value;
		// A different sheet means different columns: let the backend re-suggest.
		void requestPreview(false);
	}

	function refreshWithDefaults() {
		void requestPreview(true);
	}

	async function doImport() {
		if (!file || !canImport) return;
		importing = true;
		errorMsg = '';
		try {
			const form = new FormData();
			form.append('file', file);
			form.append('portfolioId', portfolioId);
			form.append('sourceId', sourceId);
			if (sheet) form.append('sheet', sheet);
			form.append('mapping', JSON.stringify(mapping));
			form.append('defaults', JSON.stringify(defaults));

			const res = await fetch(resolve('/dashboard/transactions/import/commit'), {
				method: 'POST',
				body: form
			});
			const body = await res.json();
			if (!res.ok || !body.success) {
				errorMsg = body?.details || body?.message || 'No se pudieron importar las transacciones.';
				return;
			}
			result = body.data as ImportResult;
			step = 'done';
		} catch {
			errorMsg = 'Error de conexión al importar las transacciones.';
		} finally {
			importing = false;
		}
	}

	function restart() {
		step = 'upload';
		file = null;
		preview = null;
		result = null;
		sheet = '';
		mapping = { ...emptyMapping };
		errorMsg = '';
	}

	function onDrop(event: DragEvent) {
		event.preventDefault();
		dragOver = false;
		selectFile(event.dataTransfer?.files?.[0]);
	}
</script>

<svelte:head>
	<title>Importar transacciones - FINEXIA</title>
	<meta
		name="description"
		content="Sube tu Excel de inversiones y registra todas tus transacciones"
	/>
</svelte:head>

<PageHeader
	title="Importar transacciones"
	subtitle="Sube el Excel donde llevas tu registro de inversiones: detectamos tus columnas y las adaptamos automáticamente."
/>

<div class="steps" aria-hidden="true">
	<span class="step-chip" class:active={step === 'upload'}>1 · Archivo</span>
	<span class="step-chip" class:active={step === 'map'}>2 · Columnas y vista previa</span>
	<span class="step-chip" class:active={step === 'done'}>3 · Resultado</span>
</div>

{#if errorMsg}
	<p class="error-banner" role="alert">{errorMsg}</p>
{/if}

{#if step === 'upload'}
	<Card variant="elevated" padding="md">
		<div class="upload-grid">
			<div class="form-group">
				<label class="form-label" for="portfolio">Portafolio destino</label>
				<select id="portfolio" class="form-select" bind:value={portfolioId}>
					{#each data.portfolios as p (p.id)}
						<option value={p.id}>{p.name} ({p.baseCurrency})</option>
					{/each}
				</select>
				{#if data.portfolios.length === 0}
					<span class="field-hint">
						Primero crea un portafolio en
						<a href={resolve('/dashboard/portfolios/add')}>Portafolios</a>.
					</span>
				{/if}
			</div>

			<div class="form-group">
				<label class="form-label" for="platform">Plataforma / broker</label>
				<select id="platform" class="form-select" bind:value={sourceId}>
					{#each data.platforms as p (p.id)}
						<option value={p.id}>{p.name}</option>
					{/each}
				</select>
				{#if data.platforms.length === 0}
					<span class="field-hint">
						Primero registra una plataforma en
						<a href={resolve('/dashboard/platforms/add')}>Plataformas</a>.
					</span>
				{/if}
			</div>
		</div>

		<div
			class="dropzone"
			class:drag-over={dragOver}
			role="button"
			tabindex="0"
			aria-label="Subir archivo de transacciones (.xlsx o .csv)"
			onclick={() => fileInput?.click()}
			onkeydown={(e) => {
				if (e.key === 'Enter' || e.key === ' ') {
					e.preventDefault();
					fileInput?.click();
				}
			}}
			ondragover={(e) => {
				e.preventDefault();
				dragOver = true;
			}}
			ondragleave={() => (dragOver = false)}
			ondrop={onDrop}
		>
			<input
				bind:this={fileInput}
				type="file"
				accept=".xlsx,.csv"
				class="sr-only"
				onchange={(e) => selectFile(e.currentTarget.files?.[0])}
			/>
			{#if loading}
				<span class="spinner"></span>
				<p class="dropzone-title">Analizando tu archivo…</p>
			{:else}
				<p class="dropzone-title">Arrastra tu Excel aquí o haz clic para buscarlo</p>
				<p class="dropzone-hint">
					Formatos .xlsx y .csv · máx. 8 MB. No importa cómo se llamen tus columnas: podrás
					asignarlas en el siguiente paso.
				</p>
			{/if}
		</div>
	</Card>
{:else if step === 'map' && preview}
	<Card variant="elevated" padding="md">
		<div class="map-header">
			<div>
				<h2 class="section-title">Asigna tus columnas</h2>
				<p class="section-hint">
					Detectamos <strong>{preview.headers.length}</strong> columnas en
					<strong>{file?.name}</strong> (encabezados en la fila {preview.headerRow}). Revisa la
					asignación sugerida y ajústala a tu formato.
				</p>
			</div>
			{#if preview.sheets.length > 1}
				<div class="form-group sheet-select">
					<label class="form-label" for="sheet">Hoja</label>
					<select
						id="sheet"
						class="form-select"
						value={sheet}
						onchange={(e) => changeSheet(e.currentTarget.value)}
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
						onchange={(e) => setMappingColumn(field.key, e.currentTarget.value)}
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
					onchange={refreshWithDefaults}
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
					onchange={refreshWithDefaults}
					placeholder="USD"
				/>
			</div>
			<div class="form-group">
				<label class="form-label" for="default-category">Categoría</label>
				<select
					id="default-category"
					class="form-select"
					bind:value={defaults.category}
					onchange={refreshWithDefaults}
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
					onchange={refreshWithDefaults}
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
					Mostrando las primeras {preview.rows.length} filas de {preview.totalRows}. El total se
					valida completo al importar.
				</p>
			{/if}
		</div>

		<div class="form-actions">
			<button type="button" class="btn btn-secondary" onclick={restart} disabled={importing}>
				Elegir otro archivo
			</button>
			<button type="button" class="btn btn-primary" onclick={doImport} disabled={!canImport}>
				{#if importing}
					<span class="spinner dark"></span> Importando…
				{:else}
					Importar {preview.validRows} transacciones
				{/if}
			</button>
		</div>
	</Card>
{:else if step === 'done' && result}
	<Card variant="elevated" padding="md">
		<div class="result-panel">
			<p class="result-icon" aria-hidden="true">{result.imported > 0 ? '✓' : '!'}</p>
			<h2 class="section-title">
				{result.imported > 0
					? `${result.imported} transacciones importadas`
					: 'No se importó ninguna transacción'}
			</h2>
			<p class="section-hint">
				{result.totalRows} filas procesadas · {result.imported} importadas · {result.skipped} omitidas
				por errores.
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
				<button type="button" class="btn btn-secondary" onclick={restart}>
					Importar otro archivo
				</button>
				<button
					type="button"
					class="btn btn-primary"
					onclick={() => goto(resolve('/dashboard/transactions'))}
				>
					Ver mis transacciones
				</button>
			</div>
		</div>
	</Card>
{/if}

<style>
	.steps {
		display: flex;
		gap: 0.6rem;
		flex-wrap: wrap;
		margin-bottom: 1.5rem;
	}

	.step-chip {
		font-family: var(--font-mono);
		font-size: 0.72rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		padding: 0.4rem 0.8rem;
		border-radius: 999px;
		border: 1px solid var(--border);
		color: var(--text-muted);
	}

	.step-chip.active {
		border-color: var(--amber);
		color: var(--amber);
		background: rgba(212, 145, 42, 0.1);
	}

	.error-banner,
	.warning-banner {
		border-radius: 10px;
		padding: 0.8rem 1rem;
		font-size: 0.85rem;
		margin-bottom: 1.2rem;
	}

	.error-banner {
		background: rgba(224, 90, 90, 0.12);
		border: 1px solid rgba(224, 90, 90, 0.4);
		color: #e05a5a;
	}

	.warning-banner {
		background: rgba(212, 145, 42, 0.1);
		border: 1px solid rgba(212, 145, 42, 0.4);
		color: var(--amber);
	}

	.upload-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
		margin-bottom: 1.5rem;
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

	.field-hint {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.field-hint a {
		color: var(--amber);
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

	.dropzone {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.6rem;
		text-align: center;
		padding: 3rem 1.5rem;
		border: 2px dashed rgba(212, 145, 42, 0.35);
		border-radius: 14px;
		background: rgba(255, 255, 255, 0.015);
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			background 0.2s ease;
	}

	.dropzone:hover,
	.dropzone:focus-visible,
	.dropzone.drag-over {
		border-color: var(--amber);
		background: rgba(212, 145, 42, 0.06);
		outline: none;
	}

	.dropzone-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text);
		margin: 0;
	}

	.dropzone-hint {
		font-size: 0.82rem;
		color: var(--text-muted);
		margin: 0;
		max-width: 30rem;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border: 0;
	}

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

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 768px) {
		.upload-grid {
			grid-template-columns: 1fr;
		}

		.form-actions {
			flex-direction: column-reverse;
		}

		.btn {
			width: 100%;
			justify-content: center;
		}
	}
</style>
