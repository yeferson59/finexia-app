<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import { PlatformAddForm } from '$lib/features/platforms';
</script>

<svelte:head>
	<title>Agregar Plataforma - FINEXIA</title>
	<meta name="description" content="Registra una nueva plataforma de inversión" />
</svelte:head>

<PageHeader
	title="Agregar Plataforma de Inversión"
	subtitle="Completa los datos de tu nueva plataforma para comenzar a rastrear tus inversiones"
/>

<PlatformAddForm />
