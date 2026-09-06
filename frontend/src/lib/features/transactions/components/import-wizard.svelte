<script lang="ts">
	/*
	 * El asistente para pasar a Finexia el Excel donde el usuario lleva sus
	 * operaciones: elegir archivo → decir qué es cada columna → confirmar.
	 *
	 * Lo que cambió respecto a la versión anterior, que era de la etapa de las
	 * tarjetas con sombra:
	 *
	 *  - Los tres pasos eran cápsulas en versalitas ámbar y, además,
	 *    `aria-hidden`: decoración que no le decía nada a quien no las ve. Ahora
	 *    son una lista ordenada de verdad, con el paso actual marcado, y el filete
	 *    que llevan encima hace de barra de avance.
	 *  - Cada paso traía su propia copia del CSS de campos, botones y `spinner`,
	 *    y ya habían empezado a divergir. Los aporta este componente una sola vez.
	 *  - El aviso de error era una caja roja rellena; ahora es prosa con un filete
	 *    al lado, el idioma que hablan configuración y el alta de portafolio.
	 */
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
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

	const STEPS: { id: ImportStep; label: string }[] = [
		{ id: 'upload', label: 'Archivo' },
		{ id: 'map', label: 'Columnas' },
		{ id: 'done', label: 'Resultado' }
	];

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
	// `costCurrency` arranca vacía a propósito: vacía significa «cada fila liquida
	// en la moneda en la que cotiza», que es lo que hacía toda importación
	// anterior y lo que describe un extracto de una sola moneda. Rellenarla es
	// una declaración del usuario, no un valor que la app deba suponer.
	let defaults = $state({
		type: 'buy',
		currency: 'USD',
		costCurrency: '',
		category: 'stock',
		dateFormat: 'auto'
	});
	let loading = $state(false);
	let importing = $state(false);
	let errorMsg = $state('');

	const stepIndex = $derived(STEPS.findIndex((s) => s.id === step));

	/*
	 * Sin portafolio o sin plataforma no hay dónde meter las transacciones. Antes
	 * se podía subir el archivo, mapearlo entero y descubrirlo al final, con el
	 * botón de importar apagado y sin decir por qué; ahora se dice en el paso uno.
	 */
	const missingDestination = $derived(portfolios.length === 0 || platforms.length === 0);

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
			errorMsg = `«${candidate.name}» no es un .xlsx ni un .csv. Exporta tu hoja a uno de los dos y vuelve a subirla.`;
			return;
		}
		if (candidate.size > 8 * 1024 * 1024) {
			errorMsg = 'El archivo pasa de 8 MB. Divídelo por años o por cuenta y súbelo por partes.';
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
				errorMsg = body?.details || body?.message || 'No pudimos leer el archivo.';
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
			errorMsg = 'No pudimos conectar para leer el archivo. Inténtalo de nuevo.';
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
				errorMsg = body?.details || body?.message || 'No pudimos importar las transacciones.';
				return;
			}
			result = body.data as ImportResult;
			step = 'done';
		} catch {
			errorMsg = 'No pudimos conectar para importar. Tus transacciones no se han guardado.';
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

<div class="wizard">
	<ol class="steps" aria-label="Pasos de la importación">
		{#each STEPS as s, i (s.id)}
			<li
				class="step"
				class:reached={i <= stepIndex}
				aria-current={s.id === step ? 'step' : undefined}
			>
				{s.label}
			</li>
		{/each}
	</ol>

	{#if errorMsg}
		<p class="feedback error" role="alert">{errorMsg}</p>
	{/if}

	{#if step === 'upload'}
		<ImportUploadStep
			{portfolios}
			{platforms}
			bind:portfolioId
			bind:sourceId
			{loading}
			{missingDestination}
			fileName={file?.name}
			onSelectFile={selectFile}
		/>
	{:else if step === 'map' && preview}
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
	{:else if step === 'done' && result}
		<ImportResultStep
			{result}
			onRestart={restart}
			onViewTransactions={() => goto(resolve('/dashboard/transactions'))}
		/>
	{/if}
</div>

<style>
	/*
	 * El avance es el filete que llevan los pasos encima, no una cápsula: la misma
	 * línea que separa los bloques del formulario, encendida hasta donde vas.
	 */
	.steps {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.75rem;
		max-width: 32rem;
		margin: 0 0 1rem;
		padding: 0;
		list-style: none;
	}

	.step {
		padding-top: 0.55rem;
		border-top: 2px solid var(--border-strong);
		font-size: 0.82rem;
		color: var(--text-dim);
		transition:
			color 0.25s ease,
			border-color 0.25s ease;
	}

	.step.reached {
		border-top-color: var(--amber);
		color: var(--text-muted);
	}

	.step[aria-current='step'] {
		color: var(--text);
	}

	/* Solo el margen: la forma y el color de un aviso los pone
	   `routes/layout.css`. */
	.feedback {
		margin: 1.5rem 0 0;
	}

	/*
	 * Los campos, sus etiquetas y las ayudas los aporta `ui/rail-section`: los
	 * pasos de este asistente son bloques de ese carril, igual que los de
	 * configuración o los del alta de portafolio. Aquí solo queda lo que es de
	 * este asistente y vive fuera de los bloques.
	 */

	/* Mono para lo que salió del archivo del usuario: es lo que distingue sus
	   datos de las palabras de la aplicación. */
	.wizard :global(.figure) {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.wizard :global(.actions) {
		display: flex;
		align-items: center;
		gap: 1.25rem;
		flex-wrap: wrap;
		padding-top: 2.25rem;
		border-top: 1px solid var(--border-strong);
	}

	/* Sin el halo ámbar de `ui/button`, como en configuración. */
	.wizard :global(.actions .btn-primary) {
		box-shadow: none;
	}

	.wizard :global(.quiet-action) {
		border: none;
		background: none;
		padding: 0;
		font-family: var(--font-body);
		font-size: 0.85rem;
		color: var(--text-muted);
		cursor: pointer;
		transition: color 0.2s ease;
	}

	.wizard :global(.quiet-action:hover:not(:disabled)) {
		color: var(--text);
	}

	.wizard :global(.quiet-action:disabled) {
		cursor: not-allowed;
		opacity: 0.6;
	}

	.wizard :global(.spinner) {
		display: inline-block;
		width: 13px;
		height: 13px;
		flex-shrink: 0;
		border: 2px solid rgba(212, 145, 42, 0.25);
		border-top-color: var(--amber);
		border-radius: 50%;
		animation: spin 0.7s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.step,
		.wizard :global(.quiet-action) {
			transition: none;
		}

		.wizard :global(.spinner) {
			animation-duration: 2s;
		}
	}
</style>
