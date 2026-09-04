<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import { AssetCreateForm, AssetsTable, ImportCard, type ImportResult } from '$lib/features/admin';
	import { flash } from '$lib/shared/flash.svelte';

	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	let showCreateForm = $state(false);
	let showImportForm = $state(false);
	const created = flash();

	const importResult = $derived((form?.importResult ?? null) as ImportResult | null);
</script>

<svelte:head>
	<title>Activos — Admin — FINEXIA</title>
</svelte:head>

<PageHeader
	eyebrow="Administración"
	title="Activos"
	subtitle="Catálogo compartido de activos y precio manual. Los precios de mercado los sincroniza cada usuario con su propia clave."
>
	{#snippet actions()}
		<div class="header-actions">
			{#if created.text}
				<span class="sync-success">{created.text}</span>
			{/if}
			<Button
				variant="secondary"
				size="sm"
				type="button"
				onclick={() => (showCreateForm = !showCreateForm)}
			>
				{showCreateForm ? 'Cancelar' : '+ Nuevo Activo'}
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
	<AssetCreateForm
		error={form?.createError ?? ''}
		onSuccess={() => {
			showCreateForm = false;
			created.show('Activo creado correctamente.');
		}}
	/>
{/if}

{#if showImportForm}
	<ImportCard
		title="Importar activos desde CSV/Excel"
		action="importAssets"
		error={form?.importError ?? ''}
		result={importResult}
		onSuccess={() => (showImportForm = false)}
	>
		{#snippet hint()}
			El archivo debe tener columnas <code>ticker</code>, <code>name</code>,
			<code>assetType</code> y <code>currency</code> (opcional: <code>exchange</code>). Se admite
			.csv, .xlsx y .xls.
		{/snippet}
	</ImportCard>
{/if}

<AssetsTable assets={data.assets} {form} />

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
</style>
