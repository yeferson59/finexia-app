<script lang="ts">
	import { formatAssetType } from '$lib/shared/format/asset-type';
	import { formatCurrency } from '$lib/shared/format/money';
	import type { Asset } from '$lib/api/types';
	import AssetCreateInline from './asset-create-inline.svelte';

	let {
		selected = $bindable(null),
		search = $bindable('')
	}: { selected?: Asset | null; search?: string } = $props();

	let showSuggestions = $state(false);
	let suggestions = $state<Asset[]>([]);
	let isSearching = $state(false);
	let isCreating = $state(false);
	let comboboxEl = $state<HTMLDivElement | null>(null);
	let debounceTimer: ReturnType<typeof setTimeout>;

	function selectAsset(asset: Asset) {
		selected = asset;
		search = asset.ticker;
		showSuggestions = false;
		isCreating = false;
	}

	async function fetchSuggestions(q: string) {
		isSearching = true;
		try {
			const url = q.trim()
				? `/api/assets?search=${encodeURIComponent(q.trim())}&limit=10`
				: `/api/assets?limit=10`;
			const res = await fetch(url);
			const json = await res.json();
			suggestions = json.success ? (json.data ?? []) : [];
		} catch {
			suggestions = [];
		} finally {
			isSearching = false;
		}
	}

	function onSearchInput() {
		selected = null;
		showSuggestions = true;
		// Escribir otro ticker invalida el formulario abierto: se creaba el que
		// estaba en pantalla al abrirlo, no el que se acaba de teclear.
		isCreating = false;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => fetchSuggestions(search), 300);
	}

	function onSearchFocus() {
		showSuggestions = true;
		if (suggestions.length === 0 && !search) fetchSuggestions('');
	}

	function clickOutside(node: HTMLElement, handler: (e: MouseEvent) => void) {
		function listener(e: MouseEvent) {
			handler(e);
		}
		document.addEventListener('mousedown', listener);
		return {
			destroy() {
				document.removeEventListener('mousedown', listener);
			}
		};
	}
</script>

<div
	class="combobox"
	bind:this={comboboxEl}
	use:clickOutside={(e) => {
		if (comboboxEl && !comboboxEl.contains(e.target as Node)) showSuggestions = false;
	}}
>
	<div class="combobox-input-wrap">
		<input
			id="asset-search"
			type="text"
			class="combobox-input"
			placeholder="AAPL, Bitcoin, Vanguard…"
			autocomplete="off"
			bind:value={search}
			oninput={onSearchInput}
			onfocus={onSearchFocus}
		/>
		{#if isSearching}
			<span class="combobox-spinner"></span>
		{:else if selected}
			<button
				type="button"
				class="combobox-clear"
				aria-label="Limpiar selección"
				onclick={() => {
					selected = null;
					search = '';
					showSuggestions = false;
				}}>✕</button
			>
		{/if}
	</div>

	<!-- El alta va primero: mientras el formulario está abierto manda sobre la
	     lista, para que una búsqueda anterior no lo tape a medio rellenar. -->
	{#if isCreating}
		<AssetCreateInline
			ticker={search}
			oncreated={selectAsset}
			oncancel={() => (isCreating = false)}
		/>
	{:else if showSuggestions && suggestions.length > 0}
		<ul class="combobox-list" role="listbox">
			{#each suggestions as asset (asset.id)}
				<li
					role="option"
					aria-selected={asset.id === selected?.id}
					class="combobox-option"
					class:selected={asset.id === selected?.id}
					onmousedown={() => selectAsset(asset)}
				>
					<div class="option-left">
						<span class="option-ticker">{asset.ticker}</span>
						<span class="option-type">{formatAssetType(asset.assetType)}</span>
					</div>
					<div class="option-right">
						<span class="option-name">{asset.name}</span>
						{#if asset.exchange || asset.currency}
							<span class="option-meta">
								{[asset.exchange, asset.currency].filter(Boolean).join(', ')}
							</span>
						{/if}
					</div>
					{#if asset.currentPrice}
						<span class="option-price">
							{formatCurrency(
								parseFloat(asset.currentPrice.value),
								asset.currency && asset.currency !== 'XXX' ? asset.currency : 'USD',
								4
							)}
						</span>
					{/if}
				</li>
			{/each}
		</ul>
	{:else if showSuggestions && !isSearching && search.trim().length > 0}
		<!--
			Crear un activo y añadirlo al portafolio son dos cosas distintas, y aquí
			se ven las dos: el catálogo no tiene el instrumento, así que primero se
			da de alta y después se sigue con la posición.
		-->
		<div class="combobox-empty">
			<span>
				No hay ningún activo que se llame <strong>{search.trim().toUpperCase()}</strong>. Créalo y
				sigue con la posición.
			</span>
			<button type="button" class="combobox-create" onclick={() => (isCreating = true)}>
				Crear {search.trim().toUpperCase()}
			</button>
		</div>
	{/if}
</div>

<style>
	/* Los campos los pinta el bloque que envuelve el formulario; aquí solo el
	   hueco del aspa y del indicador de búsqueda, a la derecha. */
	.combobox input[type='text'] {
		width: 100%;
		padding-right: 2.5rem;
	}

	.combobox {
		position: relative;
	}

	.combobox-input-wrap {
		position: relative;
		display: flex;
		align-items: center;
	}

	.combobox-spinner {
		position: absolute;
		right: 0.9rem;
		width: 14px;
		height: 14px;
		border: 2px solid var(--border-strong);
		border-top-color: var(--text-muted);
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	.combobox-clear {
		position: absolute;
		right: 0.75rem;
		background: transparent;
		border: none;
		color: var(--text-dim);
		font-size: 0.85rem;
		cursor: pointer;
		padding: 0.2rem 0.3rem;
		border-radius: 4px;
		transition: color 0.2s ease;
		line-height: 1;
	}

	.combobox-clear:hover {
		color: var(--text);
	}

	.combobox-list {
		position: absolute;
		top: calc(100% + 4px);
		left: 0;
		right: 0;
		z-index: 50;
		list-style: none;
		margin: 0;
		padding: 0.35rem 0;
		background: #101114;
		border: 1px solid var(--border-strong);
		border-radius: 10px;
		box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
		max-height: 280px;
		overflow-y: auto;
	}

	.combobox-option {
		display: grid;
		grid-template-columns: 100px 1fr auto;
		align-items: center;
		gap: 0.75rem;
		padding: 0.65rem 1rem;
		cursor: pointer;
		transition: background 0.15s ease;
	}

	.combobox-option:hover,
	.combobox-option.selected {
		background: var(--surface-2);
	}

	.option-left {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		min-width: 0;
	}

	/* El ticker en gris y no en ámbar: en esta pantalla el ámbar es el total.
	   Lo que lo distingue del nombre es la tipografía de máquina. */
	.option-ticker {
		font-family: var(--font-mono);
		font-weight: 500;
		font-size: 0.9rem;
		color: var(--text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.option-type {
		font-size: 0.72rem;
		color: var(--text-dim);
	}

	.option-right {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		min-width: 0;
	}

	.option-name {
		font-size: 0.88rem;
		color: var(--text);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.option-meta {
		font-size: 0.75rem;
		color: var(--text-dim);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.option-price {
		font-family: var(--font-mono);
		font-size: 0.82rem;
		color: var(--text-muted);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		flex-shrink: 0;
	}

	.combobox-empty {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: 0.75rem;
		margin-top: 0.65rem;
		font-size: 0.83rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	.combobox-empty strong {
		font-family: var(--font-mono);
		font-weight: 500;
		color: var(--text);
	}

	/* Callado: crear un activo es la salida de emergencia del buscador, no lo
	   que se viene a hacer a esta pantalla. */
	.combobox-create {
		flex-shrink: 0;
		padding: 0.4rem 0.8rem;
		border: 1px solid var(--border-strong);
		border-radius: 8px;
		background: none;
		color: var(--text);
		font-family: var(--font-body);
		font-size: 0.82rem;
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			background 0.2s ease;
	}

	.combobox-create:hover {
		border-color: var(--text-dim);
		background: var(--panel);
	}

	@media (prefers-reduced-motion: reduce) {
		.combobox-create {
			transition: none;
		}
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
