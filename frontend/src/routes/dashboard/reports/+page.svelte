<script lang="ts">
	import {
		GrowthProjection,
		KeyStatistics,
		PerformanceCalendars,
		ReportDownloads
	} from '$lib/features/reports';

	import type { PageProps } from './$types';

	const { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Reportes - FINEXIA</title>
	<meta name="description" content="Centro de reportes financieros y extractos" />
</svelte:head>

<header class="page-header">
	<h1 class="page-title">Reportes</h1>
	<p class="page-subtitle">Gestiona y descarga documentos financieros de tu cuenta.</p>
</header>

<PerformanceCalendars calendars={data.performanceCalendars} />

<section class="insights-grid">
	<KeyStatistics statistics={data.keyStatistics} />
	<GrowthProjection projection={data.growthProjection} />
</section>

<ReportDownloads />

<style>
	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
		font-weight: 300;
		color: var(--text);
		font-family: var(--font-display);
		letter-spacing: -0.02em;
	}

	.page-subtitle {
		margin: 0;
		color: rgba(236, 234, 229, 0.62);
	}

	/* La rejilla es de la página; el padding de los dos paneles que aloja viaja
	   con ella porque los renderiza `report-panel`, en la feature. */
	.insights-grid {
		display: grid;
		grid-template-columns: 1fr 2fr;
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

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}
	}
</style>
