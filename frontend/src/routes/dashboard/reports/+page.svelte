<script lang="ts">
	import {
		GrowthProjection,
		KeyStatistics,
		MonthlyReturns,
		RecordHeadline,
		ReportDownloads
	} from '$lib/features/reports';
	import PageHeader from '$lib/ui/page-header.svelte';

	import type { PageProps } from './$types';

	const { data }: PageProps = $props();
</script>

<svelte:head>
	<title>Reportes - FINEXIA</title>
	<meta name="description" content="El historial de resultados de tu cuenta" />
</svelte:head>

<PageHeader
	title="Reportes"
	subtitle="Cómo le ha ido a tu dinero desde que lo sigues aquí, y los archivos para llevártelo."
/>

{#if data.failed}
	<p class="empty">
		No pudimos cargar tu historial. Vuelve a intentarlo en un momento; tus datos siguen ahí.
	</p>
{:else if data.record}
	<RecordHeadline record={data.record} />

	{#if data.performanceCalendars.length > 0}
		<MonthlyReturns calendars={data.performanceCalendars} />
	{/if}

	{#if data.keyStatistics.length > 0}
		<KeyStatistics stats={data.keyStatistics} />
	{/if}

	<GrowthProjection
		projection={data.growthProjection}
		historyDays={data.historyDays}
		currency={data.record.currency}
	/>
{:else}
	<!-- Una sola invitación, no cinco bloques diciendo «sin datos» por separado:
	     todo lo de esta página sale de la misma serie, así que o está entera o
	     no está. Y es una invitación, no un aviso: quien llega aquí el primer
	     día no ha hecho nada mal. -->
	<p class="empty">
		Tu historial empieza con el primer cierre diario de la cartera. Registra una posición y mañana
		esta página tendrá su primera cifra; los archivos de abajo ya puedes descargarlos.
	</p>
{/if}

<ReportDownloads />

<style>
	.empty {
		max-width: 62ch;
		margin: 0;
		padding: 2.5rem 0;
		border-bottom: 1px solid var(--border);
		font-size: 0.9rem;
		line-height: 1.55;
		color: var(--text-muted);
	}
</style>
