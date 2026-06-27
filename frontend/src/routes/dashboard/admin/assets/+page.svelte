<script lang="ts">
	import { enhance } from '$app/forms';
	import PageHeader from '$components/ui/page-header.svelte';
	import Card from '$components/ui/card.svelte';
	import Badge from '$components/ui/badge.svelte';
	import Button from '$components/ui/button.svelte';

	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	let syncing = $state(false);
	let updatingId = $state<string | null>(null);
	let syncMessage = $state<string | null>(null);
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
	});

	function formatPrice(price: { value: string; currency: string } | null): string {
		if (!price) return '—';
		const num = parseFloat(price.value);
		if (isNaN(num)) return price.value;
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: price.currency || 'USD',
			minimumFractionDigits: 2,
			maximumFractionDigits: 4
		}).format(num);
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
			{#if form?.syncError}
				<span class="sync-error">{form.syncError}</span>
			{/if}
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
					</tr>
				</thead>
				<tbody>
					{#each data.assets as asset (asset.id)}
						{@const isUpdating = updatingId === asset.id}
						{@const hasUpdateError = form?.updateError && form?.errorId === asset.id}
						{@const hasUpdateSuccess = form?.updateSuccess && form?.updatedId === asset.id}
						<tr class:row-success={hasUpdateSuccess}>
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

	.empty-state {
		text-align: center;
		padding: 3rem;
		color: var(--text-dim);
		font-size: 0.9rem;
	}
</style>
