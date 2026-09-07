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

<!--
	Una acción de la pantalla y una de segunda fila. Antes eran dos botones con
	borde ámbar del mismo peso: dar de alta un activo y subir una hoja no se
	hacen con la misma frecuencia ni por la misma razón.
-->
<PageHeader
	title="Activos"
	subtitle="El catálogo que comparten todas las cuentas, con su precio de respaldo escrito a mano."
>
	{#snippet actions()}
		<button class="row-action" type="button" onclick={() => (showImportForm = true)}>
			Importar CSV/Excel
		</button>
		<Button type="button" onclick={() => (showCreateForm = true)}>Nuevo activo</Button>
	{/snippet}
</PageHeader>

{#if created.text}
	<p class="feedback success page-flash">{created.text}</p>
{/if}

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
			created.show('Activo creado. Ya está en el catálogo.');
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
			Una fila por activo, con las columnas <code>ticker</code>, <code>name</code>,
			<code>assetType</code> y <code>currency</code>. <code>exchange</code> es opcional. Se admiten .csv,
			.xlsx y .xls.
		{/snippet}
	</ImportCard>
</Modal>

<AssetsTable assets={data.assets} {form} />

<style>
	.page-flash {
		margin: -1rem 0 2rem;
	}
</style>
