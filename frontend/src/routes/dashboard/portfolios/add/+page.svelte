<script lang="ts">
	import { PortfolioAddForm } from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Crear Nuevo Portafolio - FINEXIA</title>
	<meta name="description" content="Crea un nuevo portafolio de inversiones personalizado" />
</svelte:head>

<PortfolioAddForm risks={data.risks} />
