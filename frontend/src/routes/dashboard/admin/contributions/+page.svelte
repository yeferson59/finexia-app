<script lang="ts">
	import PageHeader from '$lib/ui/page-header.svelte';
	import { ContributionsTable, describeContributions } from '$lib/features/admin';

	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const summary = $derived(
		data.unavailable
			? 'No se pudieron cargar los aportes. Vuelve a intentarlo en un momento.'
			: data.summary
				? describeContributions(data.summary)
				: ''
	);
</script>

<svelte:head>
	<title>Aportes — Admin — FINEXIA</title>
</svelte:head>

<PageHeader
	title="Aportes"
	subtitle="Lo que se ha aportado desde /apoyar. Las devoluciones y anulaciones se hacen en el panel de Bold."
/>

<ContributionsTable
	contributions={data.contributions}
	meta={data.meta}
	status={data.status}
	{summary}
/>
