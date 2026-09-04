<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import Button from '$lib/ui/button.svelte';
	import Modal from '$lib/ui/modal.svelte';
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
			<Button variant="secondary" size="sm" type="button" onclick={() => (showCreateForm = true)}>
				Nuevo activo
			</Button>
			<Button variant="secondary" size="sm" type="button" onclick={() => (showImportForm = true)}>
				Importar CSV/Excel
			</Button>
		</div>
	{/snippet}
</PageHeader>

<Modal
	open={showCreateForm}
	title="Nuevo activo"
	description="Se añade al catálogo compartido, así que queda disponible para todas las cuentas."
	onClose={() => (showCreateForm = false)}
	size="lg"
>
	<AssetCreateForm
		error={form?.createError ?? ''}
		onCancel={() => (showCreateForm = false)}
		onSuccess={() => {
			showCreateForm = false;
			created.show('Activo creado correctamente.');
		}}
	/>
</Modal>

<Modal
	open={showImportForm}
	title="Importar activos desde CSV/Excel"
	onClose={() => (showImportForm = false)}
	size="lg"
>
	<ImportCard
		action="importAssets"
		error={form?.importError ?? ''}
		result={importResult}
		onCancel={() => (showImportForm = false)}
		onSuccess={() => (showImportForm = false)}
	>
		{#snippet hint()}
			El archivo debe tener columnas <code>ticker</code>, <code>name</code>,
			<code>assetType</code> y <code>currency</code> (opcional: <code>exchange</code>). Se admite
			.csv, .xlsx y .xls.
		{/snippet}
	</ImportCard>
</Modal>

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
