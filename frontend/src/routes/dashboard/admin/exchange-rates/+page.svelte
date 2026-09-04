<script lang="ts">
	import { enhance } from '$app/forms';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import {
		ExchangeRateCreateForm,
		ExchangeRatesTable,
		ImportCard,
		type ImportResult
	} from '$lib/features/admin';
	import { flash } from '$lib/shared/flash.svelte';

	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	let showCreateForm = $state(false);
	let showImportForm = $state(false);
	let refreshing = $state(false);
	const created = flash();
	const refreshed = flash();

	const importResult = $derived((form?.importResult ?? null) as ImportResult | null);
</script>

<svelte:head>
	<title>Tasas de Cambio — Admin — FINEXIA</title>
</svelte:head>

<PageHeader
	eyebrow="Administración"
	title="Tasas de Cambio"
	subtitle="Tasas compartidas: la TRM la trae un feed público cada hora, el resto se mantiene a mano. Las tasas de mercado las sincroniza cada usuario con su propia clave."
>
	{#snippet actions()}
		<div class="header-actions">
			{#if created.text}
				<span class="sync-success">{created.text}</span>
			{/if}
			{#if refreshed.text}
				<span class="sync-success">{refreshed.text}</span>
			{/if}
			{#if form?.refreshError}
				<span class="sync-error">{form.refreshError}</span>
			{/if}
			<!--
				Un POST de verdad, no un `onclick`: la acción recarga el `load` de la
				página al terminar, que es lo que repinta la tabla con las tasas
				nuevas sin tener que sincronizarlas a mano en el cliente.
			-->
			<form
				method="POST"
				action="?/refreshRates"
				use:enhance={() => {
					refreshing = true;
					return async ({ result, update }) => {
						refreshing = false;
						await update({ reset: false });
						if (result.type === 'success') {
							const count = Number(result.data?.refreshedCount ?? 0);
							refreshed.show(
								`${count} tasa${count === 1 ? '' : 's'} actualizada${count === 1 ? '' : 's'} desde el feed.`
							);
						}
					};
				}}
			>
				<Button variant="secondary" size="sm" type="submit" loading={refreshing}>
					Actualizar desde el feed
				</Button>
			</form>
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
		</div>
	{/snippet}
</PageHeader>

{#if showCreateForm}
	<ExchangeRateCreateForm
		error={form?.createError ?? ''}
		onSuccess={() => {
			showCreateForm = false;
			created.show('Tasa de cambio creada correctamente.');
		}}
	/>
{/if}

{#if showImportForm}
	<ImportCard
		title="Importar tasas desde CSV/Excel"
		action="importRates"
		error={form?.importError ?? ''}
		result={importResult}
		onSuccess={() => (showImportForm = false)}
	>
		{#snippet hint()}
			El archivo debe tener columnas <code>fromCurrency</code>, <code>toCurrency</code> y
			<code>rate</code>. Se admite .csv, .xlsx y .xls.
		{/snippet}
	</ImportCard>
{/if}

<ExchangeRatesTable rates={data.rates} {form} />

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
		font-weight: 500;
	}
</style>
