<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import { PortfolioAddForm } from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>Crear portafolio - FINEXIA</title>
	<meta name="description" content="Crea un nuevo portafolio de inversiones personalizado" />
</svelte:head>

<!-- Sin filete: lo pone el primer bloque del formulario. Y sin el botón
     «Volver» que flotaba encima sin alinear con nada: al listado se vuelve por
     el menú lateral o por «Cancelar», al final del formulario. -->
<PageHeader
	title="Crear portafolio"
	subtitle="Un portafolio agrupa parte de tu dinero para poder seguirlo por separado."
	divider={false}
/>

<PortfolioAddForm risks={data.risks} defaultCurrency={data.user?.preferredCurrency} {form} />
