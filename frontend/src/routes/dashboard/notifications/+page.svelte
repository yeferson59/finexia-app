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

<!-- Sin filete: lo pone el primer canal, y dos rayas a cuarenta píxeles una de
     otra no separan nada. -->
<PageHeader title="Notificaciones" subtitle="Lo que Finexia te envía, y a dónde." divider={false} />

<EmailPreferences preferences={data.preferences} user={data.user} {form} />
<InAppAlerts />
