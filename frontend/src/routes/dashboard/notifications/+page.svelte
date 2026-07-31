<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import { EmailPreferences, InAppAlerts } from '$lib/features/notifications';

	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
</script>

<svelte:head>
	<title>Notificaciones - FINEXIA</title>
	<meta name="description" content="Configura tus preferencias de notificaciones en FINEXIA" />
</svelte:head>

<PageHeader
	title="Notificaciones"
	subtitle="Elige cómo y cuándo quieres recibir alertas sobre tu portafolio."
/>

<div class="notifications-layout">
	<EmailPreferences preferences={data.preferences} {form} />
	<InAppAlerts />
</div>

<style>
	.notifications-layout {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
		align-items: start;
	}

	@media (max-width: 1024px) {
		.notifications-layout {
			grid-template-columns: 1fr;
		}
	}
</style>
