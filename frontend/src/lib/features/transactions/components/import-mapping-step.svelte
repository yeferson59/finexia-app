<script lang="ts">
	/*
	 * Paso dos: decir qué es cada columna del archivo.
	 *
	 * Eran once desplegables idénticos, uno por campo, cuyas opciones decían
	 * «C · Ticker» sin enseñar en ningún momento qué hay dentro de la C. Asignar
	 * columnas es emparejar, y la pantalla daba dos listas que no se tocaban: el
	 * usuario tenía que bajar a la tabla de vista previa, comprobar, subir y
	 * corregir.
	 *
	 * Ahora cada campo es una fila que termina en los datos reales de la columna
	 * elegida, en mono y sacados de las primeras filas del archivo. Si eliges la
	 * columna equivocada, al lado de «Cantidad» aparecen fechas.
	 */
	import Button from '$lib/ui/button.svelte';
	import RailSection from '$lib/ui/rail-section.svelte';
	import ImportPreviewTable from './import-preview-table.svelte';
	import type { ImportMapping, ImportPreview, ImportDefaults } from '../types';
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

	type Field = { key: keyof ImportMapping; label: string };

	/* Las cuatro sin las que no hay transacción. */
	const REQUIRED: Field[] = [
		{ key: 'date', label: 'Fecha' },
		{ key: 'ticker', label: 'Ticker' },
		{ key: 'quantity', label: 'Cantidad' },
		{ key: 'price', label: 'Precio' }
	];

	const OPTIONAL: Field[] = [
		{ key: 'type', label: 'Tipo de operación' },
		{ key: 'assetName', label: 'Nombre del activo' },
		{ key: 'fees', label: 'Comisiones' },
		{ key: 'currency', label: 'Moneda' },
		{ key: 'fxRate', label: 'Tasa de cambio' },
		{ key: 'category', label: 'Categoría' },
		{ key: 'notes', label: 'Notas' }
	];

	const FIELD_LABELS: Record<string, string> = Object.fromEntries(
		[...REQUIRED, ...OPTIONAL].map((f) => [f.key, f.label])
	);

	/** A, B, … Z, AA: cómo nombra las columnas la propia hoja de cálculo. */
	function columnLetter(index: number): string {
		let label = '';
		let n = index;
		do {
			label = String.fromCharCode(65 + (n % 26)) + label;
			n = Math.floor(n / 26) - 1;
		} while (n >= 0);
		return label;
	}

	function optionLabel(header: string, index: number): string {
		const letter = columnLetter(index);
		return header ? `${header} (${letter})` : `Columna ${letter}`;
	}

	/** Los primeros valores con contenido de una columna, tal cual venían. */
	function samples(column: number | null): string[] {
		if (column === null) return [];
		const found: string[] = [];
		for (const row of preview.rows) {
			const value = row.raw?.[column]?.trim();
			if (value) found.push(value);
			if (found.length === 3) break;
		}
		return found;
	}

	const missing = $derived(preview.missingFields.map((f) => FIELD_LABELS[f] ?? f));

	const importLabel = $derived(
		preview.validRows === 0
			? 'No hay filas que importar'
			: preview.validRows === 1
				? 'Importar 1 transacción'
				: `Importar ${preview.validRows} transacciones`
	);
</script>

{#snippet columnRow(field: Field, required: boolean)}
	<!--
		`?? null` y no `=== null`: una asignación sugerida que el backend no manda
		llega como `undefined`, y con `String(undefined)` el desplegable se queda
		en blanco, sin coincidir siquiera con «Sin asignar».
	-->
	{@const column = mapping[field.key] ?? null}
	{@const values = samples(column)}
	<li class="row" class:pending={required && column === null}>
		<div class="field">
			<label for={`map-${field.key}`}>
				{field.label}{#if !required}<span class="optional">&nbsp;(opcional)</span>{/if}
			</label>
			<select
				id={`map-${field.key}`}
				value={column === null ? '' : String(column)}
				onchange={(e) => onSetMappingColumn(field.key, e.currentTarget.value)}
			>
				<option value="">Sin asignar</option>
				{#each preview.headers as header, i (i)}
					<option value={String(i)}>{optionLabel(header, i)}</option>
				{/each}
			</select>
		</div>

		<!-- Sin columna asignada no se dice nada: el propio desplegable ya pone
		     «Sin asignar», y repetirlo en siete filas seguidas era ruido. -->
		{#if values.length > 0}
			<p class="samples figure">
				{#each values as value, i (i)}
					<span class="cell">{value}</span>
				{/each}
			</p>
		{:else if required}
			<p class="samples pending-note">Elige la columna que trae este dato.</p>
		{/if}
	</li>
{/snippet}

<p class="lead">
	{preview.headers.length} columnas en <span class="figure">{fileName}</span>, con los encabezados
	en la fila {preview.headerRow}.
</p>

<RailSection
	title="Qué es cada columna"
	description="A la derecha de cada campo van los primeros valores de la columna que elijas. Si no son los que esperabas, la columna es otra."
	contentMax="none"
	fields
>
	{#if preview.sheets.length > 1}
		<div class="field sheet">
			<label for="sheet">Hoja del archivo</label>
			<select
				id="sheet"
				class="sheet-select"
				value={sheet}
				onchange={(e) => onChangeSheet(e.currentTarget.value)}
			>
				{#each preview.sheets as s (s)}
					<option value={s}>{s}</option>
				{/each}
			</select>
		</div>
	{/if}

	{#if missing.length > 0}
		<p class="missing" role="alert">
			Falta decir de qué columna sale {missing.join(', ')}. Sin eso no se puede importar.
		</p>
	{/if}

	<ul class="rows">
		{#each REQUIRED as field (field.key)}
			{@render columnRow(field, true)}
		{/each}
	</ul>

	<p class="group-lead">Y estas otras, si tu archivo las trae.</p>

	<ul class="rows">
		{#each OPTIONAL as field (field.key)}
			{@render columnRow(field, false)}
		{/each}
	</ul>
</RailSection>

<RailSection
	title="Lo que el archivo no diga"
	description="Los valores con los que se rellenan las filas a las que les falte el dato. No pisan lo que sí venga en una columna."
	fields
>
	<div class="defaults">
		<div class="field">
			<label for="default-type">Tipo de operación</label>
			<select id="default-type" bind:value={defaults.type} onchange={onRefreshDefaults}>
				{#each TXN_TYPE_OPTIONS as t (t.value)}
					<option value={t.value}>{t.label}</option>
				{/each}
			</select>
		</div>

		<div class="field">
			<label for="default-category">Categoría</label>
			<select id="default-category" bind:value={defaults.category} onchange={onRefreshDefaults}>
				{#each CATEGORY_OPTIONS as c (c.value)}
					<option value={c.value}>{c.label}</option>
				{/each}
			</select>
		</div>

		<div class="field">
			<label for="default-currency">Moneda</label>
			<input
				id="default-currency"
				type="text"
				maxlength="3"
				bind:value={defaults.currency}
				onchange={onRefreshDefaults}
				placeholder="USD"
			/>
		</div>

		<div class="field">
			<label for="default-dates">Formato de fecha</label>
			<select id="default-dates" bind:value={defaults.dateFormat} onchange={onRefreshDefaults}>
				<option value="auto">Detectar automáticamente</option>
				<option value="dmy">Día/Mes/Año</option>
				<option value="mdy">Mes/Día/Año</option>
			</select>
		</div>

		<div class="field wide">
			<label for="default-cost-currency">Moneda de la cuenta</label>
			<input
				id="default-cost-currency"
				type="text"
				maxlength="3"
				bind:value={defaults.costCurrency}
				onchange={onRefreshDefaults}
				placeholder="la misma de cada fila"
			/>
			<p class="hint">
				En la que tu bróker debitó. Déjala vacía si el extracto no convirtió nada; si la rellenas,
				cada fila en otra moneda necesita su tasa.
			</p>
		</div>
	</div>
</RailSection>

<RailSection
	title="Antes de importar"
	description="Las primeras filas tal como quedarán registradas. Las que no se puedan interpretar se quedan fuera y aquí dicen por qué."
	contentMax="none"
>
	<ImportPreviewTable {preview} {loading} />
</RailSection>

<div class="actions">
	<Button
		type="button"
		variant="primary"
		onclick={onImport}
		disabled={!canImport}
		loading={importing}
	>
		{importing ? 'Importando…' : importLabel}
	</Button>
	<button type="button" class="quiet-action" onclick={onRestart} disabled={importing}>
		Elegir otro archivo
	</button>
</div>

<style>
	.lead {
		max-width: 62ch;
		margin: 1.75rem 0 0.5rem;
		font-size: 0.88rem;
		line-height: 1.6;
		color: var(--text-muted);
	}

	/* El filete llega hasta el final del carril; el desplegable, no: los nombres
	   de hoja son cortos. */
	.sheet {
		align-items: flex-start;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border-strong);
	}

	/* Con su propia clase, y no `.sheet select`: esa empata en especificidad con
	   el `select { width: 100% }` que aporta el asistente y perdía por orden. */
	.field.sheet .sheet-select {
		width: 14rem;
		max-width: 100%;
	}

	.missing {
		max-width: 62ch;
		margin: 0;
		padding-left: 0.75rem;
		border-left: 2px solid var(--red);
		font-size: 0.83rem;
		line-height: 1.5;
		color: var(--red);
	}

	.rows {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/*
	 * Campo, desplegable y datos reales en la misma línea. El tercer carril es el
	 * que hace el trabajo: es donde se ve si la columna elegida es la correcta.
	 */
	.row {
		display: grid;
		grid-template-columns: minmax(0, 18rem) minmax(0, 1fr);
		align-items: end;
		gap: 0.5rem 2rem;
		padding: 1rem 0;
		border-bottom: 1px solid var(--border);
	}

	.rows .row:last-child {
		border-bottom: none;
	}

	.samples {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem 1.5rem;
		min-width: 0;
		margin: 0 0 0.7rem;
		font-size: 0.82rem;
	}

	.cell {
		max-width: 22ch;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--text);
	}

	/* El segundo y el tercer valor bajan de tono: el primero es el que se lee. */
	.cell:nth-child(2) {
		color: var(--text-muted);
	}

	.cell:nth-child(3) {
		color: var(--text-dim);
	}

	.samples.pending-note {
		font-family: var(--font-body);
		color: var(--red);
	}

	.group-lead {
		margin: 0;
		padding-top: 0.75rem;
		border-top: 1px solid var(--border-strong);
		font-size: 0.83rem;
		color: var(--text-muted);
	}

	.defaults {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1.35rem;
	}

	.wide {
		grid-column: 1 / -1;
	}

	@media (max-width: 900px) {
		.row {
			grid-template-columns: minmax(0, 1fr);
			align-items: stretch;
		}

		.samples {
			margin-bottom: 0;
		}
	}

	@media (max-width: 640px) {
		.defaults {
			grid-template-columns: minmax(0, 1fr);
		}

		.sheet {
			width: 100%;
		}
	}
</style>
