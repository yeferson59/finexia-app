<script lang="ts">
	import type { Asset } from '$lib/api/types';

	let {
		selected = $bindable(null),
		search = $bindable('')
	}: { selected?: Asset | null; search?: string } = $props();

	let showSuggestions = $state(false);
	let suggestions = $state<Asset[]>([]);
	let isSearching = $state(false);
	let comboboxEl = $state<HTMLDivElement | null>(null);
	let debounceTimer: ReturnType<typeof setTimeout>;

	function selectAsset(asset: Asset) {
		selected = asset;
		search = asset.ticker;
		showSuggestions = false;
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
			class="form-input combobox-input"
			placeholder="Escribe el ticker o nombre, ej: AAPL, Bitcoin…"
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

	{#if showSuggestions && suggestions.length > 0}
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
						<span class="option-type">{asset.assetType}</span>
					</div>
					<div class="option-right">
						<span class="option-name">{asset.name}</span>
						{#if asset.exchange || asset.currency}
							<span class="option-meta">
								{[asset.exchange, asset.currency].filter(Boolean).join(' · ')}
							</span>
						{/if}
					</div>
					{#if asset.currentPrice}
						<span class="option-price">
							{new Intl.NumberFormat('en-US', {
								style: 'currency',
								currency: asset.currency && asset.currency !== 'XXX' ? asset.currency : 'USD',
								minimumFractionDigits: 2,
								maximumFractionDigits: 4
							}).format(parseFloat(asset.currentPrice.value))}
						</span>
					{/if}
				</li>
			{/each}
		</ul>
	{:else if showSuggestions && !isSearching && search.trim().length > 0}
		<div class="combobox-empty">
			No se encontró ningún activo con "<strong>{search}</strong>"
		</div>
	{/if}
</div>

<style>
	.form-input {
		padding: 0.85rem 1rem;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.022);
		color: var(--text);
		font-size: 0.95rem;
		font-family: var(--font-body);
		transition: all 0.3s ease;
	}

	.form-input::placeholder {
		color: rgba(236, 234, 229, 0.55);
	}

	.form-input:focus {
		outline: none;
		border-color: var(--amber);
		background: rgba(255, 255, 255, 0.022);
		box-shadow: 0 0 0 3px var(--border);
	}

	.combobox {
		position: relative;
	}

	.combobox-input-wrap {
		position: relative;
		display: flex;
		align-items: center;
	}

	.combobox-input {
		width: 100%;
		padding-right: 2.5rem;
	}

	.combobox-spinner {
		position: absolute;
		right: 0.9rem;
		width: 14px;
		height: 14px;
		border: 2px solid rgba(212, 145, 42, 0.25);
		border-top-color: var(--amber);
		border-radius: 50%;
		animation: spin 0.6s linear infinite;
	}

	.combobox-clear {
		position: absolute;
		right: 0.75rem;
		background: transparent;
		border: none;
		color: rgba(236, 234, 229, 0.4);
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
		background: var(--surface);
		border: 1.5px solid rgba(212, 145, 42, 0.35);
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
		background: rgba(212, 145, 42, 0.1);
	}

	.option-left {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		min-width: 0;
	}

	.option-ticker {
		font-family: var(--font-mono);
		font-weight: 700;
		font-size: 0.92rem;
		color: var(--amber);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.option-type {
		font-size: 0.68rem;
		color: rgba(236, 234, 229, 0.4);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
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
		color: rgba(236, 234, 229, 0.4);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.option-price {
		font-family: var(--font-mono);
		font-size: 0.82rem;
		color: rgba(212, 145, 42, 0.7);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
		flex-shrink: 0;
	}

	.combobox-empty {
		padding: 0.75rem 1rem;
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.5);
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
