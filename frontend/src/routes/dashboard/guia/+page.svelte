<script lang="ts">
	import { GuideViewer, GuideContents } from '$lib/features/guide';
	import PageHeader from '$lib/ui/page-header.svelte';
</script>

<svelte:head>
	<title>Guía de usuario - FINEXIA</title>
	<meta name="description" content="Manual de usuario de Finexia, para consultar o descargar" />
</svelte:head>

<PageHeader
	eyebrow="Ayuda"
	title="Guía de usuario"
	subtitle="El manual completo de Finexia: léelo aquí mismo o descárgalo para tenerlo a mano."
/>

<div class="stack">
	<GuideViewer />
	<GuideContents />
</div>

<style>
	.stack {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}
</style>
