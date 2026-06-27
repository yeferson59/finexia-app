<script lang="ts">
	import { enhance } from '$app/forms';
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import Badge from '$components/ui/badge.svelte';
	import Button from '$components/ui/button.svelte';

	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	const ASSET_TYPES = [
		{ value: 'stock', label: 'Acción (Stock)' },
		{ value: 'etf', label: 'ETF' },
		{ value: 'crypto', label: 'Cripto' },
		{ value: 'bond', label: 'Bono' },
		{ value: 'real_estate', label: 'Bienes raíces' },
		{ value: 'commodity', label: 'Commodities' },
		{ value: 'cash', label: 'Efectivo' },
		{ value: 'other', label: 'Otro' }
	];

	let syncing = $state(false);
	let updatingId = $state<string | null>(null);
	let syncingAssetId = $state<string | null>(null);
	let creating = $state(false);
	let showCreateForm = $state(false);
	let syncMessage = $state<string | null>(null);
	let createMessage = $state<string | null>(null);
	let priceInputs = $state<Record<string, string>>({});

	$effect(() => {
		for (const asset of data.assets) {
			if (!(asset.id in priceInputs)) {
				priceInputs[asset.id] = asset.currentPrice?.value ?? '';
			}
		}
	});

	$effect(() => {
		if (form?.syncSuccess) {
			syncMessage = `${form.synced} activo${form.synced === 1 ? '' : 's'} sincronizado${form.synced === 1 ? '' : 's'}.`;
			setTimeout(() => (syncMessage = null), 4000);
		}
		if (form?.createSuccess) {
			showCreateForm = false;
			createMessage = 'Activo creado correctamente.';
			setTimeout(() => (createMessage = null), 4000);
		}
		if (form?.syncAssetSuccess) {
			syncMessage = `Precio de activo actualizado.`;
			setTimeout(() => (syncMessage = null), 4000);
		}
	});

	function formatPrice(price: { value: string; currency: string } | null): string {
		if (!price) return '—';
		const num = parseFloat(price.value);
		if (isNaN(num)) return price.value;
		const currency = price.currency || 'USD';
		try {
			return new Intl.NumberFormat('en-US', {
				style: 'currency',
				currency,
				currencyDisplay: 'narrowSymbol',
				minimumFractionDigits: 2,
				maximumFractionDigits: 4
			}).format(num);
		} catch {
			return `${currency} ${num.toFixed(2)}`;
		}
	}

	function formatDate(iso: string | null): string {
		if (!iso) return '—';
		return new Intl.DateTimeFormat('es', {
			dateStyle: 'short',
			timeStyle: 'short'
		}).format(new Date(iso));
	}
</script>

<svelte:head>
	<title>Activos — Admin — FINEXIA</title>
</svelte:head>

<PageHeader eyebrow="Administración" title="Activos" subtitle="Gestiona precios y sincronización de activos.">
	{#snippet actions()}
		<div class="header-actions">
			{#if syncMessage}
				<span class="sync-success">{syncMessage}</span>
			{/if}
			{#if createMessage}
				<span class="sync-success">{createMessage}</span>
			{/if}
			{#if form?.syncError}
				<span class="sync-error">{form.syncError}</span>
			{/if}
			<Button
				variant="secondary"
				size="sm"
				type="button"
				onclick={() => (showCreateForm = !showCreateForm)}
			>
				{showCreateForm ? 'Cancelar' : '+ Nuevo Activo'}
			</Button>
			<form
				method="POST"
				action="?/syncPrices"
				use:enhance={() => {
					syncing = true;
					return async ({ update }) => {
						syncing = false;
						await update({ reset: false });
					};
				}}
			>
				<Button type="submit" loading={syncing} size="sm">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polyline points="23 4 23 10 17 10"></polyline>
						<polyline points="1 20 1 14 7 14"></polyline>
						<path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
					</svg>
					Sincronizar Precios
				</Button>
			</form>
		</div>
	{/snippet}
</PageHeader>

{#if showCreateForm}
	<div class="create-form-wrap">
		<Card padding="md">
			<h2 class="form-title">Nuevo activo</h2>
			<form
				method="POST"
				action="?/createAsset"
				use:enhance={() => {
					creating = true;
					return async ({ update }) => {
						creating = false;
						await update();
					};
				}}
			>
				<div class="form-grid">
					<div class="form-field">
						<label class="field-label" for="ticker">Ticker <span class="required">*</span></label>
						<input id="ticker" name="ticker" class="field-input" placeholder="AAPL" required />
					</div>
					<div class="form-field">
						<label class="field-label" for="name">Nombre <span class="required">*</span></label>
						<input id="name" name="name" class="field-input" placeholder="Apple Inc." required />
					</div>
					<div class="form-field">
						<label class="field-label" for="assetType">Tipo <span class="required">*</span></label>
						<select id="assetType" name="assetType" class="field-input field-select" required>
							<option value="" disabled selected>Seleccionar tipo</option>
							{#each ASSET_TYPES as t}
								<option value={t.value}>{t.label}</option>
							{/each}
						</select>
					</div>
					<div class="form-field">
						<label class="field-label" for="currency">Moneda <span class="required">*</span></label>
						<input id="currency" name="currency" class="field-input" placeholder="USD" maxlength="3" required />
					</div>
					<div class="form-field">
						<label class="field-label" for="exchange">Exchange <span class="optional">(opcional)</span></label>
						<input id="exchange" name="exchange" class="field-input" placeholder="NASDAQ" />
					</div>
				</div>
				{#if form?.createError}
					<p class="form-error">{form.createError}</p>
				{/if}
				<div class="form-actions">
					<Button type="submit" loading={creating}>Crear activo</Button>
				</div>
			</form>
		</Card>
	</div>
{/if}

<Card padding="none">
	{#if data.assets.length === 0}
		<p class="empty-state">No hay activos en el sistema.</p>
	{:else}
		<div class="table-wrapper">
			<table class="assets-table">
				<thead>
					<tr>
						<th>Ticker</th>
						<th>Nombre</th>
						<th>Tipo</th>
						<th>Precio actual</th>
						<th>Actualizado</th>
						<th>Nuevo precio</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each data.assets as asset (asset.id)}
						{@const isUpdating = updatingId === asset.id}
						{@const hasUpdateError = form?.updateError && form?.errorId === asset.id}
						{@const hasUpdateSuccess = form?.updateSuccess && form?.updatedId === asset.id}
						{@const hasSyncError = form?.syncAssetError && form?.syncAssetId === asset.id}
						{@const isSyncingThis = syncingAssetId === asset.id}
						<tr class:row-success={hasUpdateSuccess || (form?.syncAssetSuccess && form?.syncAssetId === asset.id)}>
							<td class="cell-ticker">{asset.ticker}</td>
							<td class="cell-name">{asset.name}</td>
							<td>
								<Badge tone="neutral">{asset.assetType}</Badge>
							</td>
							<td class="cell-price">{formatPrice(asset.currentPrice)}</td>
							<td class="cell-date">{formatDate(asset.priceUpdatedAt)}</td>
							<td class="cell-update">
								<form
									method="POST"
									action="?/updatePrice"
									use:enhance={() => {
										updatingId = asset.id;
										return async ({ update }) => {
											updatingId = null;
											await update({ reset: false });
										};
									}}
								>
									<input type="hidden" name="id" value={asset.id} />
									<input type="hidden" name="currency" value={asset.currentPrice?.currency ?? asset.currency ?? 'USD'} />
									<div class="update-row">
										<input
											type="number"
											name="price"
											class="price-input"
											class:input-error={hasUpdateError}
											bind:value={priceInputs[asset.id]}
											min="0.0001"
											step="any"
											placeholder="0.00"
											required
										/>
										<Button type="submit" size="sm" variant="secondary" loading={isUpdating}>
											OK
										</Button>
									</div>
									{#if hasUpdateError}
										<p class="row-error">{form.updateError}</p>
									{/if}
								</form>
							</td>
							<td class="cell-sync">
								<form
									method="POST"
									action="?/syncAsset"
									use:enhance={() => {
										syncingAssetId = asset.id;
										return async ({ update }) => {
											syncingAssetId = null;
											await update({ reset: false });
										};
									}}
								>
									<input type="hidden" name="id" value={asset.id} />
									<Button type="submit" size="sm" variant="ghost" loading={isSyncingThis}>
										<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="23 4 23 10 17 10"></polyline>
											<polyline points="1 20 1 14 7 14"></polyline>
											<path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
										</svg>
										Sync
									</Button>
									{#if hasSyncError}
										<p class="row-error">{form.syncAssetError}</p>
									{/if}
								</form>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</Card>

<style>
	.header-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.sync-success {
		font-size: 0.82rem;
		color: var(--green);
		font-weight: 500;
	}

	.sync-error {
		font-size: 0.82rem;
		color: var(--red);
	}

	.table-wrapper {
		overflow-x: auto;
	}

	.assets-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.assets-table th {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		color: var(--text-dim);
		padding: 0.875rem 1.25rem;
		text-align: left;
		border-bottom: 1px solid var(--border);
		white-space: nowrap;
	}

	.assets-table td {
		padding: 0.75rem 1.25rem;
		color: var(--text-muted);
		border-bottom: 1px solid var(--border);
		vertical-align: middle;
	}

	.assets-table tbody tr:last-child td {
		border-bottom: none;
	}

	.assets-table tbody tr:hover td {
		background: var(--surface-2);
	}

	.row-success td {
		background: rgba(76, 175, 80, 0.05) !important;
	}

	.cell-ticker {
		font-family: var(--font-mono);
		font-weight: 600;
		color: var(--amber-light) !important;
		font-size: 0.85rem;
	}

	.cell-name {
		color: var(--text) !important;
		font-weight: 500;
	}

	.cell-price {
		font-family: var(--font-mono);
		font-weight: 600;
		color: var(--text) !important;
		white-space: nowrap;
	}

	.cell-date {
		font-family: var(--font-mono);
		font-size: 0.78rem;
		white-space: nowrap;
	}

	.cell-update {
		min-width: 160px;
	}

	.cell-sync {
		white-space: nowrap;
	}

	.update-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.price-input {
		width: 90px;
		padding: 0.4rem 0.6rem;
		background: var(--surface-2);
		border: 1px solid var(--border-strong);
		border-radius: 6px;
		color: var(--text);
		font-family: var(--font-mono);
		font-size: 0.82rem;
		transition: border-color 0.2s ease;
	}

	.price-input:focus {
		outline: none;
		border-color: var(--amber);
	}

	.price-input.input-error {
		border-color: var(--red);
	}

	.row-error {
		font-size: 0.75rem;
		color: var(--red);
		margin: 0.25rem 0 0 0;
	}

	.create-form-wrap {
		margin-bottom: 1.5rem;
	}

	.form-title {
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--text);
		margin: 0 0 1.25rem 0;
	}

	.form-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.form-field {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.field-label {
		font-family: var(--font-mono);
		font-size: 0.65rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--text-dim);
	}

	.required {
		color: var(--red);
	}

	.optional {
		color: var(--text-dim);
		font-weight: 400;
	}

	.field-input {
		padding: 0.55rem 0.75rem;
		background: var(--surface-2);
		border: 1px solid var(--border-strong);
		border-radius: 6px;
		color: var(--text);
		font-size: 0.875rem;
		font-family: var(--font-body);
		transition: border-color 0.2s ease;
	}

	.field-input:focus {
		outline: none;
		border-color: var(--amber);
	}

	.field-select {
		cursor: pointer;
		appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23888' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 0.6rem center;
		padding-right: 2rem;
	}

	.field-select option {
		background: var(--bg);
		color: var(--text);
	}

	.form-error {
		font-size: 0.82rem;
		color: var(--red);
		margin: 0 0 0.75rem 0;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
	}

	.empty-state {
		text-align: center;
		padding: 3rem;
		color: var(--text-dim);
		font-size: 0.9rem;
	}
</style>
