<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import { ImportWizard } from '$lib/features/transactions';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Importar transacciones - FINEXIA</title>
	<meta
		name="description"
		content="Sube tu Excel de inversiones y registra todas tus transacciones"
	/>
</svelte:head>

<PageHeader
	title="Importar transacciones"
	subtitle="Sube el Excel donde llevas tu registro de inversiones: detectamos tus columnas y las adaptamos automáticamente."
/>

<ImportWizard portfolios={data.portfolios} platforms={data.platforms} />
