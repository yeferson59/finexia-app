<script lang="ts">
	import {
		GrowthProjection,
		KeyStatistics,
		PerformanceCalendars,
		ReportDownloads
	} from '$lib/features/reports';
	import PageHeader from '$lib/ui/page-header.svelte';

	import type { PageProps } from './$types';

	const { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Reportes - FINEXIA</title>
	<meta name="description" content="Centro de reportes financieros y extractos" />
</svelte:head>

<PageHeader
	eyebrow="Documentos"
	title="Reportes"
	subtitle="Gestiona y descarga documentos financieros de tu cuenta."
/>

<PerformanceCalendars calendars={data.performanceCalendars} />

<section class="insights-grid">
	<KeyStatistics groups={data.keyStatistics} />
	<GrowthProjection projection={data.growthProjection} historyDays={data.historyDays} />
</section>

<ReportDownloads />

<style>
	/* La rejilla es de la página; el padding de los dos paneles que aloja viaja
	   con ella porque los renderiza `report-panel`, en la feature. */
	/* `align-items: start` para que el panel de estadísticas no se estire hasta
	   el alto de la gráfica y deje media tarjeta vacía. */
	/* Columnas iguales: con la de estadísticas a un tercio, sus doce métricas
	   caían en una sola columna y la tarjeta triplicaba el alto de la gráfica. */
	.insights-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		align-items: start;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.insights-grid :global(.stats-card),
	.insights-grid :global(.projection-card) {
		padding: 1rem;
	}

	@media (max-width: 1024px) {
		.insights-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
