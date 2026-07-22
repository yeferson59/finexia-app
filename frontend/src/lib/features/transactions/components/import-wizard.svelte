<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Card from '$lib/ui/card.svelte';
	import ImportUploadStep from './import-upload-step.svelte';
	import ImportMappingStep from './import-mapping-step.svelte';
	import ImportResultStep from './import-result-step.svelte';
	import {
		emptyMapping,
		type ImportMapping,
		type ImportPreview,
		type ImportResult,
		type ImportStep,
		type ImportPortfolioOption,
		type ImportPlatformOption
	} from '../types';

	let {
		portfolios,
		platforms
	}: { portfolios: ImportPortfolioOption[]; platforms: ImportPlatformOption[] } = $props();

	let step: ImportStep = $state('upload');
	let file: File | null = $state(null);
	// The load data only seeds the initial selection; the user owns it afterwards.
	// svelte-ignore state_referenced_locally
	let portfolioId = $state(portfolios.find((p) => p.isDefault)?.id ?? portfolios[0]?.id ?? '');
	// svelte-ignore state_referenced_locally
	let sourceId = $state(platforms[0]?.id ?? '');
	let sheet = $state('');
	let preview: ImportPreview | null = $state(null);
	let result: ImportResult | null = $state(null);
	let mapping: ImportMapping = $state({ ...emptyMapping });
	let defaults = $state({ type: 'buy', currency: 'USD', category: 'stock', dateFormat: 'auto' });
	let loading = $state(false);
	let importing = $state(false);
	let errorMsg = $state('');

	const canImport = $derived.by(() => {
		if (!preview || loading || importing) return false;
		return (
			preview.validRows > 0 && preview.missingFields.length === 0 && !!portfolioId && !!sourceId
		);
	});

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
</script>

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
		<ImportUploadStep
			{portfolios}
			{platforms}
			bind:portfolioId
			bind:sourceId
			{loading}
			onSelectFile={selectFile}
		/>
	</Card>
{:else if step === 'map' && preview}
	<Card variant="elevated" padding="md">
		<ImportMappingStep
			{preview}
			fileName={file?.name}
			{sheet}
			{mapping}
			bind:defaults
			{loading}
			{importing}
			{canImport}
			onChangeSheet={changeSheet}
			onSetMappingColumn={setMappingColumn}
			onRefreshDefaults={refreshWithDefaults}
			onRestart={restart}
			onImport={doImport}
		/>
	</Card>
{:else if step === 'done' && result}
	<Card variant="elevated" padding="md">
		<ImportResultStep
			{result}
			onRestart={restart}
			onViewTransactions={() => goto(resolve('/dashboard/transactions'))}
		/>
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

	.error-banner {
		border-radius: 10px;
		padding: 0.8rem 1rem;
		font-size: 0.85rem;
		margin-bottom: 1.2rem;
		background: rgba(224, 90, 90, 0.12);
		border: 1px solid rgba(224, 90, 90, 0.4);
		color: #e05a5a;
	}
</style>
