<script lang="ts">
	import { GuideViewer, GuideContents, manual } from '$lib/features/guide';
	import PageHeader from '$lib/ui/page-header.svelte';
</script>

<svelte:head>
	<title>Guía de usuario - FINEXIA</title>
	<meta name="description" content="Manual de usuario de Finexia, para consultar o descargar" />
</svelte:head>

<!-- Sin filete: lo pone la ficha del documento, y dos rayas seguidas no separan
     nada. Sin antetítulo tampoco: el menú lateral ya dijo dónde estás. -->
<PageHeader
	title="Guía de usuario"
	subtitle="Cómo funciona cada parte de Finexia, en {manual.sections
		.length} capítulos. Léelos aquí o descarga el manual para tenerlo a mano."
	divider={false}
/>

<GuideViewer />
<GuideContents />
