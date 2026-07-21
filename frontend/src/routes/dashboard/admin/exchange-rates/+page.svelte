<script lang="ts">
	import { enhance } from '$app/forms';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Card from '$lib/ui/card.svelte';
	import Button from '$lib/ui/button.svelte';

	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	interface ImportResult {
		totalRows: number;
		imported: number;
		skipped: number;
		errors: { row: number; message: string }[];
	}

	let syncing = $state(false);
	let updatingId = $state<string | null>(null);
	let syncingRateId = $state<string | null>(null);
	let creating = $state(false);
	let importing = $state(false);
	let showCreateForm = $state(false);
	let showImportForm = $state(false);
	let syncMessage = $state<string | null>(null);
	let createMessage = $state<string | null>(null);
	let rateInputs = $state<Record<string, string>>({});

	$effect(() => {
		for (const rate of data.rates) {
			if (!(rate.id in rateInputs)) {
				rateInputs[rate.id] = rate.rate ?? '';
			}
		}
	});

	$effect(() => {
		if (form?.syncSuccess) {
			syncMessage = `${form.synced} tasa${form.synced === 1 ? '' : 's'} sincronizada${form.synced === 1 ? '' : 's'}.`;
			setTimeout(() => (syncMessage = null), 4000);
		}
		if (form?.createSuccess) {
			showCreateForm = false;
			createMessage = 'Tasa de cambio creada correctamente.';
			setTimeout(() => (createMessage = null), 4000);
		}
		if (form?.syncRateSuccess) {
			syncMessage = `Tasa de cambio actualizada.`;
			setTimeout(() => (syncMessage = null), 4000);
		}
		if (form?.importSuccess) {
			showImportForm = false;
		}
	});

	const importResult = $derived((form?.importResult ?? null) as ImportResult | null);

	function formatRate(rate: string): string {
		const num = parseFloat(rate);
		if (isNaN(num)) return rate;
		return num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 });
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
	<title>Tasas de Cambio — Admin — FINEXIA</title>
</svelte:head>

<PageHeader
	eyebrow="Administración"
	title="Tasas de Cambio"
	subtitle="Gestiona pares de divisas y sincronización de tasas."
>
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
				{showCreateForm ? 'Cancelar' : '+ Nueva Tasa'}
			</Button>
			<Button
				variant="secondary"
				size="sm"
				type="button"
				onclick={() => (showImportForm = !showImportForm)}
			>
				{showImportForm ? 'Cancelar' : 'Importar CSV/Excel'}
			</Button>
			<form
				method="POST"
				action="?/syncRates"
				use:enhance={() => {
					syncing = true;
					return async ({ update }) => {
						syncing = false;
						await update({ reset: false });
					};
				}}
			>
				<Button type="submit" loading={syncing} size="sm">
					<svg
						width="14"
						height="14"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<polyline points="23 4 23 10 17 10"></polyline>
						<polyline points="1 20 1 14 7 14"></polyline>
						<path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
					</svg>
					Sincronizar Tasas
				</Button>
			</form>
		</div>
	{/snippet}
</PageHeader>

{#if showCreateForm}
	<div class="create-form-wrap">
		<Card padding="md">
			<h2 class="form-title">Nueva tasa de cambio</h2>
			<form
				method="POST"
				action="?/createRate"
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
						<label class="field-label" for="fromCurrency"
							>Moneda origen <span class="required">*</span></label
						>
						<input
							id="fromCurrency"
							name="fromCurrency"
							class="field-input"
							placeholder="USD"
							maxlength="3"
							required
						/>
					</div>
					<div class="form-field">
						<label class="field-label" for="toCurrency"
							>Moneda destino <span class="required">*</span></label
						>
						<input
							id="toCurrency"
							name="toCurrency"
							class="field-input"
							placeholder="COP"
							maxlength="3"
							required
						/>
					</div>
					<div class="form-field">
						<label class="field-label" for="rate">Tasa <span class="required">*</span></label>
						<input
							id="rate"
							name="rate"
							type="number"
							class="field-input"
							placeholder="4000.00"
							min="0.00000001"
							step="any"
							required
						/>
					</div>
				</div>
				{#if form?.createError}
					<p class="form-error">{form.createError}</p>
				{/if}
				<div class="form-actions">
					<Button type="submit" loading={creating}>Crear tasa</Button>
				</div>
			</form>
		</Card>
	</div>
{/if}

{#if showImportForm}
	<div class="create-form-wrap">
		<Card padding="md">
			<h2 class="form-title">Importar tasas desde CSV/Excel</h2>
			<p class="import-hint">
				El archivo debe tener columnas <code>fromCurrency</code>, <code>toCurrency</code> y
				<code>rate</code>. Se admite .csv, .xlsx y .xls.
			</p>
			<form
				method="POST"
				action="?/importRates"
				enctype="multipart/form-data"
				use:enhance={() => {
					importing = true;
					return async ({ update }) => {
						importing = false;
						await update();
					};
				}}
			>
				<div class="import-row">
					<input type="file" name="file" accept=".csv,.xlsx,.xls" class="field-input" required />
					<Button type="submit" loading={importing}>Importar</Button>
				</div>
				{#if form?.importError}
					<p class="form-error">{form.importError}</p>
				{/if}
			</form>
			{#if importResult}
				<div class="import-result">
					<p class="import-summary">
						{importResult.imported} de {importResult.totalRows} fila{importResult.totalRows === 1
							? ''
							: 's'} importada{importResult.imported === 1 ? '' : 's'}{importResult.skipped > 0
							? `, ${importResult.skipped} omitida${importResult.skipped === 1 ? '' : 's'}`
							: ''}.
					</p>
					{#if importResult.errors.length > 0}
						<ul class="import-errors">
							{#each importResult.errors as e (e.row)}
								<li>Fila {e.row}: {e.message}</li>
							{/each}
						</ul>
					{/if}
				</div>
			{/if}
		</Card>
	</div>
{/if}

<Card padding="none">
	{#if data.rates.length === 0}
		<p class="empty-state">No hay tasas de cambio en el sistema.</p>
	{:else}
		<div class="table-wrapper">
			<table class="assets-table">
				<thead>
					<tr>
						<th>Par</th>
						<th>Tasa actual</th>
						<th>Fecha de tasa</th>
						<th>Actualizado</th>
						<th>Nueva tasa</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each data.rates as rate (rate.id)}
						{@const isUpdating = updatingId === rate.id}
						{@const hasUpdateError = form?.updateError && form?.errorId === rate.id}
						{@const hasUpdateSuccess = form?.updateSuccess && form?.updatedId === rate.id}
						{@const hasSyncError = form?.syncRateError && form?.syncRateId === rate.id}
						{@const isSyncingThis = syncingRateId === rate.id}
						<tr
							class:row-success={hasUpdateSuccess ||
								(form?.syncRateSuccess && form?.syncRateId === rate.id)}
						>
							<td class="cell-ticker">
								{rate.fromCurrency}/{rate.toCurrency}
							</td>
							<td class="cell-price">{formatRate(rate.rate)}</td>
							<td class="cell-date">{formatDate(rate.rateDate)}</td>
							<td class="cell-date">{formatDate(rate.createdAt)}</td>
							<td class="cell-update">
								<form
									method="POST"
									action="?/updateRate"
									use:enhance={() => {
										updatingId = rate.id;
										return async ({ update }) => {
											updatingId = null;
											await update({ reset: false });
										};
									}}
								>
									<input type="hidden" name="id" value={rate.id} />
									<div class="update-row">
										<input
											type="number"
											name="rate"
											class="price-input"
											class:input-error={hasUpdateError}
											bind:value={rateInputs[rate.id]}
											min="0.00000001"
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
									action="?/syncRate"
									use:enhance={() => {
										syncingRateId = rate.id;
										return async ({ update }) => {
											syncingRateId = null;
											await update({ reset: false });
										};
									}}
								>
									<input type="hidden" name="id" value={rate.id} />
									<Button type="submit" size="sm" variant="ghost" loading={isSyncingThis}>
										<svg
											width="13"
											height="13"
											viewBox="0 0 24 24"
											fill="none"
											stroke="currentColor"
											stroke-width="2"
										>
											<polyline points="23 4 23 10 17 10"></polyline>
											<polyline points="1 20 1 14 7 14"></polyline>
											<path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"
											></path>
										</svg>
										Sync
									</Button>
									{#if hasSyncError}
										<p class="row-error">{form.syncRateError}</p>
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
		display: flex;
		align-items: center;
		gap: 0.5rem;
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

	.import-hint {
		font-size: 0.82rem;
		color: var(--text-muted);
		margin: 0 0 1rem 0;
	}

	.import-hint code {
		font-family: var(--font-mono);
		font-size: 0.78rem;
		color: var(--amber-light);
	}

	.import-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.import-row .field-input {
		flex: 1;
	}

	.import-result {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
	}

	.import-summary {
		font-size: 0.85rem;
		color: var(--text);
		margin: 0;
	}

	.import-errors {
		margin: 0.6rem 0 0 0;
		padding-left: 1.1rem;
		font-size: 0.78rem;
		color: var(--red);
		max-height: 200px;
		overflow-y: auto;
	}

	.import-errors li {
		margin-bottom: 0.25rem;
	}
</style>
