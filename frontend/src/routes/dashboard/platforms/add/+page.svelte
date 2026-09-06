<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import { PlatformAddForm } from '$lib/features/platforms';
	import type { PageProps } from './$types';

	const { form }: PageProps = $props();
</script>

<svelte:head>
	<title>Crear plataforma - FINEXIA</title>
	<meta name="description" content="Registra una nueva plataforma de inversión" />
</svelte:head>

<!-- Sin subtítulo: lo que hay que saber antes de escribir no cabe en una línea
     y lo cuenta el propio formulario, sobre el primer campo. -->
<PageHeader title="Crear plataforma" />

<PlatformAddForm {form} />
