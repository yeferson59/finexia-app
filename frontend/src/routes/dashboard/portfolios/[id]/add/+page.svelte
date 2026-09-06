<script lang="ts">
	import { PortfolioEntryForm } from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { params, data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>Añadir activo - FINEXIA</title>
	<meta name="description" content="Añade un activo que ya tienes a tu portafolio" />
</svelte:head>

<PortfolioEntryForm
	portfolioId={params.id}
	platforms={data?.platforms ?? []}
	submitError={form?.success === false}
	submitErrorDetail={form?.error ?? ''}
/>
