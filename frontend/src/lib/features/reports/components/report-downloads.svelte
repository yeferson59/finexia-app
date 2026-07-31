<script lang="ts">
	/** Tarjetas de descarga de los reportes en XLSX. */
	import { resolve } from '$app/paths';
	import ReportPanel from './report-panel.svelte';
	import { REPORT_DOWNLOADS } from '../reports';
</script>

<section class="cards-grid">
	{#each REPORT_DOWNLOADS as report (report.title)}
		<ReportPanel class="report-card">
			<div class="badge">{report.format}</div>
			<h2>{report.title}</h2>
			<a href={resolve(`/dashboard/reports/download?type=${report.type}`)} class="download">
				Descargar
			</a>
		</ReportPanel>
	{/each}
</section>

<style>
	.cards-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
	}

	.cards-grid :global(.report-card) {
		padding: 1.25rem;
		display: grid;
		gap: 0.7rem;
	}

	.badge {
		width: fit-content;
		font-size: 0.68rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		padding: 0.3rem 0.55rem;
		border-radius: 999px;
		background: var(--border-strong);
		color: var(--amber-light);
	}

	h2 {
		margin: 0;
		font-size: 1.05rem;
		color: var(--text);
	}

	.download {
		display: block;
		text-align: center;
		text-decoration: none;
		margin-top: 0.35rem;
		border: 1px solid rgba(212, 145, 42, 0.25);
		border-radius: 8px;
		padding: 0.75rem 1rem;
		background: var(--border);
		color: var(--text);
		font-weight: 600;
		cursor: pointer;
	}

	@media (max-width: 1024px) {
		.cards-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
