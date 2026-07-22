<script lang="ts">
	import { resolve } from '$app/paths';
	import type { ImportPortfolioOption, ImportPlatformOption } from '../types';

	let {
		portfolios,
		platforms,
		portfolioId = $bindable(),
		sourceId = $bindable(),
		loading,
		onSelectFile
	}: {
		portfolios: ImportPortfolioOption[];
		platforms: ImportPlatformOption[];
		portfolioId: string;
		sourceId: string;
		loading: boolean;
		onSelectFile: (file: File | undefined | null) => void;
	} = $props();

	let dragOver = $state(false);
	let fileInput: HTMLInputElement | null = $state(null);

	function onDrop(event: DragEvent) {
		event.preventDefault();
		dragOver = false;
		onSelectFile(event.dataTransfer?.files?.[0]);
	}
</script>

<div class="upload-grid">
	<div class="form-group">
		<label class="form-label" for="portfolio">Portafolio destino</label>
		<select id="portfolio" class="form-select" bind:value={portfolioId}>
			{#each portfolios as p (p.id)}
				<option value={p.id}>{p.name} ({p.baseCurrency})</option>
			{/each}
		</select>
		{#if portfolios.length === 0}
			<span class="field-hint">
				Primero crea un portafolio en
				<a href={resolve('/dashboard/portfolios/add')}>Portafolios</a>.
			</span>
		{/if}
	</div>

	<div class="form-group">
		<label class="form-label" for="platform">Plataforma / broker</label>
		<select id="platform" class="form-select" bind:value={sourceId}>
			{#each platforms as p (p.id)}
				<option value={p.id}>{p.name}</option>
			{/each}
		</select>
		{#if platforms.length === 0}
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
		onchange={(e) => onSelectFile(e.currentTarget.files?.[0])}
	/>
	{#if loading}
		<span class="spinner"></span>
		<p class="dropzone-title">Analizando tu archivo…</p>
	{:else}
		<p class="dropzone-title">Arrastra tu Excel aquí o haz clic para buscarlo</p>
		<p class="dropzone-hint">
			Formatos .xlsx y .csv · máx. 8 MB. No importa cómo se llamen tus columnas: podrás asignarlas
			en el siguiente paso.
		</p>
	{/if}
</div>

<style>
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

	.field-hint {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.field-hint a {
		color: var(--amber);
	}

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

	@media (max-width: 768px) {
		.upload-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
