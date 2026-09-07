<script lang="ts">
	import { enhance } from '$app/forms';
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
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

<PageHeader title="Tasas de cambio" subtitle="Las paridades que comparten todas las cuentas.">
	{#snippet actions()}
		<!--
			Un POST de verdad, no un `onclick`: la acción recarga el `load` de la
			página al terminar, que es lo que repinta la tabla con las tasas nuevas
			sin tener que sincronizarlas a mano en el cliente.
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
			<button class="row-action" type="submit" disabled={refreshing}>
				{refreshing ? 'Actualizando…' : 'Actualizar desde el feed'}
			</button>
		</form>
		<button class="row-action" type="button" onclick={() => (showImportForm = true)}>
			Importar CSV/Excel
		</button>
		<Button type="button" onclick={() => (showCreateForm = true)}>Nueva tasa</Button>
	{/snippet}
</PageHeader>

{#if created.text || refreshed.text}
	<p class="feedback success page-flash">{created.text || refreshed.text}</p>
{/if}
{#if form?.refreshError}
	<p class="feedback error page-flash" role="alert">{form.refreshError}</p>
{/if}

<Modal
	open={showCreateForm}
	title="Nueva tasa de cambio"
	description="Las tasas son compartidas: la que guardes aquí la usan todas las cuentas."
	onClose={() => (showCreateForm = false)}
	size="lg"
>
	<ExchangeRateCreateForm
		error={form?.createError ?? ''}
		onCancel={() => (showCreateForm = false)}
		onSuccess={() => {
			showCreateForm = false;
			created.show('Tasa creada. Ya la usan todas las cuentas.');
		}}
	/>
</Modal>

<Modal
	open={showImportForm}
	title="Importar tasas desde CSV/Excel"
	onClose={() => (showImportForm = false)}
	size="lg"
>
	<ImportCard
		action="importRates"
		error={form?.importError ?? ''}
		result={importResult}
		onCancel={() => (showImportForm = false)}
		onSuccess={() => (showImportForm = false)}
	>
		{#snippet hint()}
			Una fila por paridad, con las columnas <code>fromCurrency</code>,
			<code>toCurrency</code> y <code>rate</code>. Se admiten .csv, .xlsx y .xls.
		{/snippet}
	</ImportCard>
</Modal>

<ExchangeRatesTable rates={data.rates} {form} />

<style>
	.page-flash {
		margin: -1rem 0 2rem;
	}
</style>
