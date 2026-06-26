<script lang="ts">
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const reports = [
		{ title: 'Resumen mensual', format: 'XLSX', type: 'summary' },
		{ title: 'Estado de resultados', format: 'XLSX', type: 'transactions' },
		{ title: 'Riesgo y volatilidad', format: 'XLSX', type: 'risk' }
	];

	const months = [
		'Ene',
		'Feb',
		'Mar',
		'Abr',
		'May',
		'Jun',
		'Jul',
		'Ago',
		'Sep',
		'Oct',
		'Nov',
		'Dic'
	];

	const projectionMax = $derived(
		data.growthProjection.length > 0 ? Math.max(...data.growthProjection.map((p) => p.value)) : 0
	);
	const projectionMin = $derived(
		data.growthProjection.length > 0 ? Math.min(...data.growthProjection.map((p) => p.value)) : 0
	);
	const projectionRange = $derived(projectionMax - projectionMin || 1);

	function performanceClass(value: number): string {
		if (value >= 2) return 'strong-positive';
		if (value >= 1) return 'positive';
		if (value >= 0) return 'flat-positive';
		if (value > -1) return 'negative';
		return 'strong-negative';
	}
</script>

<svelte:head>
	<title>Reportes - FINEXIA</title>
	<meta name="description" content="Centro de reportes financieros y extractos" />
</svelte:head>

<header class="page-header">
	<h1 class="page-title">Reportes</h1>
	<p class="page-subtitle">Gestiona y descarga documentos financieros de tu cuenta.</p>
</header>

<section class="analytics-grid">
	{#if data.performanceCalendars.length > 0}
		{#each data.performanceCalendars as calendar (calendar.year)}
			<article class="panel calendar-card">
				<div class="section-head">
					<h2>Performance Calendar (%)</h2>
					<span>{calendar.year}</span>
				</div>
				<div class="calendar-grid">
					{#each calendar.values as value, index (`${calendar.year}-${months[index]}`)}
						{#if value === null}
							<div class="month-cell null-cell">
								<p class="month">{months[index]}</p>
								<p class="percent">–</p>
							</div>
						{:else}
							<div class={`month-cell ${performanceClass(value)}`}>
								<p class="month">{months[index]}</p>
								<p class="percent">{value > 0 ? '+' : ''}{value.toFixed(1)}%</p>
							</div>
						{/if}
					{/each}
				</div>
			</article>
		{/each}
	{:else}
		<div class="panel empty-panel">
			<p class="empty-text">Sin datos históricos</p>
		</div>
	{/if}
</section>

<section class="insights-grid">
	<article class="panel stats-card">
		<div class="section-head">
			<h2>Key Statistics</h2>
		</div>
		<div class="stats-list">
			{#if data.keyStatistics.length > 0}
				{#each data.keyStatistics as stat (stat.label)}
					<div class="stat-row">
						<p>{stat.label}</p>
						<p>{stat.value}</p>
					</div>
				{/each}
			{:else}
				<p class="empty-text">Sin datos</p>
			{/if}
		</div>
	</article>

	<article class="panel projection-card">
		<div class="section-head">
			<h2>Growth Projection</h2>
		</div>
		{#if data.growthProjection.length > 0}
			<svg class="projection-chart" viewBox="0 0 600 280" preserveAspectRatio="xMidYMid meet">
				<defs>
					<linearGradient id="projectionGradient" x1="0%" y1="0%" x2="0%" y2="100%">
						<stop offset="0%" style="stop-color: var(--amber); stop-opacity: 0.25" />
						<stop offset="100%" style="stop-color: var(--amber); stop-opacity: 0" />
					</linearGradient>
				</defs>
				{#each Array.from({ length: 5 }) as _, i (i)}
					<line
						x1="40"
						y1={35 + i * 50}
						x2="560"
						y2={35 + i * 50}
						stroke="var(--border)"
						stroke-width="1"
					/>
				{/each}
				<polyline
					points={data.growthProjection
						.map((point, i) => {
							const x = 40 + i * 130;
							const y = 230 - ((point.value - projectionMin) / projectionRange) * 180;
							return `${x},${y}`;
						})
						.join(' ')}
					fill="none"
					stroke="var(--amber)"
					stroke-width="3"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
				<polygon
					points={`${data.growthProjection
						.map((point, i) => {
							const x = 40 + i * 130;
							const y = 230 - ((point.value - projectionMin) / projectionRange) * 180;
							return `${x},${y}`;
						})
						.join(' ')} 560,230 40,230`}
					fill="url(#projectionGradient)"
				/>
				{#each data.growthProjection as point, i (point.period)}
					<circle
						cx={40 + i * 130}
						cy={230 - ((point.value - projectionMin) / projectionRange) * 180}
						r="4"
						fill="var(--amber-light)"
						stroke="rgba(255, 255, 255, 0.022)"
						stroke-width="2"
					/>
					<text
						x={40 + i * 130}
						y="260"
						text-anchor="middle"
						fill="rgba(236, 234, 229,0.56)"
						font-size="12"
					>
						{point.period}
					</text>
				{/each}
			</svg>
		{:else}
			<div class="empty-chart">
				<p>Proyección disponible con al menos 6 meses de historial.</p>
			</div>
		{/if}
	</article>
</section>

<section class="cards-grid">
	{#each reports as report (report.title)}
		<article class="panel report-card">
			<div class="badge">{report.format}</div>
			<h2>{report.title}</h2>
			<a href={`/dashboard/reports/download?type=${report.type}`} class="download">
				Descargar
			</a>
		</article>
	{/each}
</section>

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

	.analytics-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.insights-grid {
		display: grid;
		grid-template-columns: 1fr 2fr;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.cards-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
	}

	.panel {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
	}

	.section-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.9rem;
	}

	.section-head h2 {
		margin: 0;
		font-size: 1rem;
		color: var(--text);
	}

	.section-head span {
		font-size: 0.75rem;
		padding: 0.25rem 0.6rem;
		border-radius: 999px;
		background: rgba(212, 145, 42, 0.12);
		color: var(--amber-light);
	}

	.calendar-card {
		padding: 1rem;
	}

	.calendar-grid {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 0.45rem;
	}

	.month-cell {
		padding: 0.5rem;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.022);
		border: 1px solid transparent;
	}

	.month {
		margin: 0;
		font-size: 0.65rem;
		color: rgba(236, 234, 229, 0.55);
	}

	.percent {
		margin: 0.18rem 0 0;
		font-size: 0.76rem;
		font-weight: 700;
	}

	.month-cell.strong-positive {
		background: rgba(34, 201, 126, 0.26);
		border-color: rgba(34, 201, 126, 0.45);
		color: var(--green);
	}

	.month-cell.positive {
		background: rgba(34, 201, 126, 0.18);
		border-color: rgba(34, 201, 126, 0.3);
		color: var(--green);
	}

	.month-cell.flat-positive {
		background: rgba(212, 145, 42, 0.2);
		border-color: rgba(212, 145, 42, 0.35);
		color: var(--amber-light);
	}

	.month-cell.negative {
		background: rgba(224, 90, 90, 0.16);
		border-color: rgba(224, 90, 90, 0.3);
		color: var(--red);
	}

	.month-cell.strong-negative {
		background: rgba(224, 90, 90, 0.26);
		border-color: rgba(224, 90, 90, 0.46);
		color: var(--red);
	}

	.month-cell.null-cell {
		background: rgba(255, 255, 255, 0.022);
		border-color: transparent;
		color: rgba(236, 234, 229, 0.28);
	}

	.stats-card {
		padding: 1rem;
	}

	.stats-list {
		display: grid;
		gap: 0.45rem;
	}

	.stat-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: rgba(255, 255, 255, 0.022);
		padding: 0.6rem 0.75rem;
		border-radius: 8px;
	}

	.stat-row p {
		margin: 0;
		font-size: 0.8rem;
	}

	.stat-row p:first-child {
		color: rgba(236, 234, 229, 0.62);
	}

	.stat-row p:last-child {
		font-weight: 700;
		color: var(--amber-light);
	}

	.projection-card {
		padding: 1rem;
	}

	.projection-chart {
		width: 100%;
		min-height: 280px;
		display: block;
	}

	.report-card {
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

	.meta {
		margin: 0;
		font-size: 0.82rem;
		color: rgba(236, 234, 229, 0.56);
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

	.empty-panel {
		padding: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		grid-column: 1 / -1;
	}

	.empty-text {
		margin: 0;
		font-size: 0.85rem;
		color: rgba(236, 234, 229, 0.45);
	}

	.empty-chart {
		padding: 3rem 2rem;
		text-align: center;
		color: rgba(236, 234, 229, 0.45);
		font-size: 0.82rem;
		border: 1px dashed var(--border);
		border-radius: 8px;
		line-height: 1.6;
	}

	.empty-chart p {
		margin: 0;
	}

	@media (max-width: 1024px) {
		.analytics-grid {
			grid-template-columns: 1fr;
		}

		.insights-grid {
			grid-template-columns: 1fr;
		}

		.cards-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}
	}
</style>
