<script lang="ts">
	/*
	 * Paso uno: dónde entran los movimientos y de qué archivo salen.
	 *
	 * La zona de arrastre era un `<div role="button">` con su propio manejador de
	 * teclado; ahora es la etiqueta del `<input type="file">`, así que abrir el
	 * selector con el teclado y anunciar el control lo hace el navegador. El
	 * anillo de foco se pinta sobre la etiqueta, que es lo que se ve.
	 */
	import { resolve } from '$app/paths';
	import RailSection from '$lib/ui/rail-section.svelte';
	import type { ImportPortfolioOption, ImportPlatformOption } from '../types';

	let {
		portfolios,
		platforms,
		portfolioId = $bindable(),
		sourceId = $bindable(),
		loading,
		missingDestination,
		fileName,
		onSelectFile
	}: {
		portfolios: ImportPortfolioOption[];
		platforms: ImportPlatformOption[];
		portfolioId: string;
		sourceId: string;
		loading: boolean;
		missingDestination: boolean;
		fileName: string | undefined;
		onSelectFile: (file: File | undefined | null) => void;
	} = $props();

	let dragOver = $state(false);

	function onDrop(event: DragEvent) {
		event.preventDefault();
		dragOver = false;
		if (missingDestination) return;
		onSelectFile(event.dataTransfer?.files?.[0]);
	}
</script>

<RailSection
	title="Destino"
	description="Dónde quedan registrados los movimientos del archivo. Un archivo entra entero en un solo portafolio."
	fields
>
	<div class="pair">
		<div class="field">
			<label for="portfolio">Portafolio</label>
			<select id="portfolio" bind:value={portfolioId} disabled={portfolios.length === 0}>
				{#each portfolios as p (p.id)}
					<option value={p.id}>{p.name} ({p.baseCurrency})</option>
				{/each}
			</select>
			{#if portfolios.length === 0}
				<p class="hint">
					Todavía no tienes portafolios.
					<a href={resolve('/dashboard/portfolios/add')}>Crea el primero</a> y vuelve aquí.
				</p>
			{/if}
		</div>

		<div class="field">
			<label for="platform">Plataforma</label>
			<select id="platform" bind:value={sourceId} disabled={platforms.length === 0}>
				{#each platforms as p (p.id)}
					<option value={p.id}>{p.name}</option>
				{/each}
			</select>
			{#if platforms.length === 0}
				<p class="hint">
					Todavía no registras plataformas.
					<a href={resolve('/dashboard/platforms/add')}>Registra tu bróker</a> y vuelve aquí.
				</p>
			{/if}
		</div>
	</div>
</RailSection>

<RailSection
	title="Tu archivo"
	description="El Excel o el CSV donde llevas tus operaciones, con las columnas tal como estén. En el siguiente paso dices qué es cada una."
>
	<!--
		El contenedor solo aporta el arrastrar-y-soltar, que es un atajo de ratón:
		el control accesible es el `<input type="file">` con su etiqueta, y ese sí
		lo anuncia el navegador. De ahí `presentation`.
	-->
	<div
		class="drop"
		class:over={dragOver}
		class:busy={loading}
		role="presentation"
		ondragover={(e) => {
			e.preventDefault();
			if (!missingDestination) dragOver = true;
		}}
		ondragleave={() => (dragOver = false)}
		ondrop={onDrop}
	>
		<input
			id="import-file"
			type="file"
			accept=".xlsx,.csv"
			class="sr-only"
			disabled={loading || missingDestination}
			onchange={(e) => onSelectFile(e.currentTarget.files?.[0])}
		/>

		{#if loading}
			<p class="drop-label reading">
				<span class="spinner"></span>
				Leyendo <span class="figure">{fileName}</span>
			</p>
			<p class="drop-hint">Buscamos los encabezados y proponemos una asignación de columnas.</p>
		{:else if missingDestination}
			<p class="drop-label off">Primero elige el destino</p>
			<p class="drop-hint">Con un portafolio y una plataforma arriba, aquí sueltas el archivo.</p>
		{:else}
			<label class="drop-label" for="import-file">Arrastra el archivo o elígelo</label>
			<p class="drop-hint">.xlsx o .csv, hasta 8 MB. Da igual cómo se llamen tus columnas.</p>
		{/if}
	</div>
</RailSection>

<style>
	/*
	 * Sigue siendo un rectángulo punteado —es la señal universal de «suéltalo
	 * aquí» y quitarla no mejora nada—, pero con el filete de un pelo del resto
	 * del panel en lugar de dos píxeles ámbar, y el texto alineado a la izquierda
	 * como todo lo demás de la página.
	 */
	.drop {
		padding: 2.25rem 1.75rem;
		border: 1px dashed var(--border-strong);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.012);
		transition:
			border-color 0.2s ease,
			background 0.2s ease;
	}

	.drop.over {
		border-color: var(--amber);
		border-style: solid;
		background: rgba(212, 145, 42, 0.06);
	}

	.drop.busy {
		border-style: solid;
	}

	.drop-label {
		display: flex;
		align-items: center;
		gap: 0.55rem;
		margin: 0;
		font-size: 1rem;
		font-weight: 500;
		color: var(--text);
		cursor: pointer;
	}

	.drop-label.reading,
	.drop-label.off {
		cursor: default;
	}

	.drop-label.off {
		color: var(--text-muted);
	}

	.drop-hint {
		max-width: 42ch;
		margin: 0.5rem 0 0;
		font-size: 0.83rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	/* El `input` está oculto pero sigue siendo el control que recibe el foco: el
	   anillo se pinta sobre la etiqueta, que es lo que hay en pantalla. */
	.drop input:focus-visible + .drop-label {
		outline: 2px solid var(--amber);
		outline-offset: 4px;
		border-radius: 4px;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
		border: 0;
	}

	@media (prefers-reduced-motion: reduce) {
		.drop {
			transition: none;
		}
	}
</style>
