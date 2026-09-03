<script lang="ts">
	import { PortfolioEntryForm } from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { params, data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>Agregar Activo al Portafolio - FINEXIA</title>
	<meta name="description" content="Añade un nuevo activo a tu portafolio de inversiones" />
</svelte:head>

<PortfolioEntryForm
	portfolioId={params.id}
	platforms={data?.platforms ?? []}
	submitError={form?.success === false}
	submitErrorDetail={form?.error ?? ''}
/>
