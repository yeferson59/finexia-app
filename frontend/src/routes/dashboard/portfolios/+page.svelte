<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	function openPortfolio(id: string) {
		goto(`/dashboard/portfolios/${id}`);
	}

	function createPortfolio() {
		goto('/dashboard/portfolios/add');
	}
</script>

<svelte:head>
	<title>Portafolios - FINEXIA</title>
	<meta name="description" content="Gestión de múltiples portafolios de inversión" />
</svelte:head>

<header class="page-header">
	<div class="header-top">
		<div>
			<h1 class="page-title">Portafolios</h1>
			<p class="page-subtitle">Gestiona tus múltiples portafolios de inversión en un solo lugar.</p>
		</div>
		<button onclick={createPortfolio} class="btn-create-portfolio">
			<svg
				width="18"
				height="18"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<path d="M12 5v14M5 12h14" />
			</svg>
			Crear Portafolio
		</button>
	</div>
</header>

<section class="summary-cards">
	<article class="panel summary-card">
		<p class="eyebrow">Valor total</p>
		<h2 class="hero-value">$2,000,000</h2>
		<p class="hero-delta positive">+0.9% hoy</p>
	</article>

	<article class="panel summary-card">
		<p class="eyebrow">Portafolios activos</p>
		<h2 class="hero-value">3</h2>
		<p class="hero-delta">Diversificados</p>
	</article>

	<article class="panel summary-card">
		<p class="eyebrow">Rendimiento</p>
		<h2 class="hero-value">+12.3%</h2>
		<p class="hero-delta positive">Este año</p>
	</article>
</section>

<section class="portfolios-section">
	<h2 class="section-title">Tus Portafolios</h2>

	<div class="portfolios-grid">
		{#each data.portfolios as portfolio (portfolio.id)}
			<button
				class="portfolio-card"
				onclick={() => openPortfolio(portfolio.id)}
				aria-label={`Abrir ${portfolio.name}`}
			>
				<div class="card-header">
					<div class="icon-container">{portfolio.icon}</div>
					<div class="portfolio-info">
						<h3 class="portfolio-name">{portfolio.name}</h3>
						<p class="portfolio-type">{portfolio.type}</p>
					</div>
					<div
						class="risk-badge"
						class:low={portfolio.risk.name.toLowerCase().includes('bajo')}
						class:moderate={portfolio.risk.name.toLowerCase().includes('moderado')}
						class:high={portfolio.risk.name.toLowerCase().includes('alto')}
					>
						{portfolio.risk.name}
					</div>
				</div>

				<div class="card-metrics">
					<div class="metric">
						<p class="label">Valor</p>
						<p class="value">
							${new Intl.NumberFormat('es-CO').format(portfolio.priceValue.value)}
						</p>
					</div>

					<div class="metric">
						<p class="label">Activos</p>
						<p class="value">{portfolio.assets}</p>
					</div>

					<div class="metric">
						<p class={`value ${portfolio.dayChange >= 0 ? 'positive' : 'negative'}`}>
							{portfolio.dayChange >= 0 ? '+' : ''}{portfolio.dayChange}%
						</p>
						<p class="label">Hoy</p>
					</div>
				</div>

				<div class="card-footer">
					<div class="allocation-bar">
						<div class="bar-fill" style={`width: ${portfolio.allocation}%`}></div>
					</div>
					<p class="allocation-text">{portfolio.allocation}% de tu portafolio</p>
				</div>

				<svg
					class="arrow-icon"
					width="20"
					height="20"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<path d="M9 18l6-6-6-6" />
				</svg>
			</button>
		{/each}
	</div>
</section>

<style>
	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border);
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.5rem;
		font-weight: 300;
		letter-spacing: -0.02em;
		color: var(--text);
		font-family: var(--font-display);
	}

	.page-subtitle {
		margin: 0;
		color: rgba(236, 234, 229, 0.62);
		font-size: 1rem;
	}

	.header-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.btn-create-portfolio {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.85rem 1.5rem;
		border: none;
		border-radius: 10px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 0.95rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
		white-space: nowrap;
	}

	.btn-create-portfolio:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.summary-cards {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
		margin-bottom: 3rem;
	}

	.summary-card {
		padding: 1.35rem;
	}

	.eyebrow {
		margin: 0 0 0.55rem;
		font-size: 0.72rem;
		letter-spacing: 0.7px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.46);
	}

	.hero-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		margin: 0;
		font-size: 1.6rem;
		color: var(--text);
	}

	.hero-delta {
		margin: 0.4rem 0 0;
		font-size: 0.82rem;
		color: rgba(236, 234, 229, 0.55);
	}

	.hero-delta.positive {
		color: var(--green);
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

	.portfolios-section {
		margin-top: 2rem;
	}

	.section-title {
		font-size: 1.3rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
		margin: 0 0 1.5rem;
	}

	.portfolios-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 1.5rem;
	}

	.portfolio-card {
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		padding: 1.35rem;
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.portfolio-card:hover {
		background: var(--surface-2);
		border-color: rgba(212, 145, 42, 0.3);
		transform: translateY(-4px);
		box-shadow:
			0 30px 80px rgba(0, 0, 0, 0.4),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
	}

	.card-header {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		position: relative;
	}

	.icon-container {
		font-size: 2.5rem;
		min-width: 50px;
	}

	.portfolio-info {
		flex: 1;
	}

	.portfolio-name {
		margin: 0 0 0.3rem;
		font-size: 1.15rem;
		color: var(--text);
		font-weight: 600;
	}

	.portfolio-type {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.52);
	}

	.risk-badge {
		padding: 0.3rem 0.7rem;
		border-radius: 20px;
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		white-space: nowrap;
	}

	.risk-badge.low {
		background: rgba(34, 201, 126, 0.15);
		color: var(--green);
		border: 1px solid rgba(34, 201, 126, 0.3);
	}

	.risk-badge.moderate {
		background: rgba(241, 196, 15, 0.15);
		color: var(--amber-light);
		border: 1px solid rgba(241, 196, 15, 0.3);
	}

	.risk-badge.high {
		background: rgba(224, 90, 90, 0.15);
		color: var(--red);
		border: 1px solid rgba(224, 90, 90, 0.3);
	}

	.card-metrics {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1rem;
		padding: 1rem 0;
		border-top: 1px solid var(--border);
		border-bottom: 1px solid var(--border);
	}

	.metric {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
	}

	.metric .label {
		margin: 0;
		font-size: 0.7rem;
		letter-spacing: 0.5px;
		text-transform: uppercase;
		color: rgba(236, 234, 229, 0.46);
	}

	.metric .value {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 600;
		color: var(--amber-light);
	}

	.metric .value.positive {
		color: var(--green);
	}

	.metric .value.negative {
		color: var(--red);
	}

	.card-footer {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.allocation-bar {
		height: 6px;
		border-radius: 999px;
		background: rgba(236, 234, 229, 0.12);
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		border-radius: inherit;
		background: var(--amber);
	}

	.allocation-text {
		margin: 0;
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.5);
	}

	.arrow-icon {
		position: absolute;
		top: 1rem;
		right: 1rem;
		color: rgba(212, 145, 42, 0.3);
		transition: all 0.3s ease;
	}

	.portfolio-card:hover .arrow-icon {
		color: var(--amber-light);
		transform: translateX(4px);
	}

	@media (max-width: 1024px) {
		.summary-cards {
			grid-template-columns: 1fr;
		}

		.portfolios-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 2rem;
		}

		.summary-cards {
			grid-template-columns: 1fr;
		}

		.portfolios-grid {
			grid-template-columns: 1fr;
		}

		.card-metrics {
			grid-template-columns: repeat(3, 1fr);
		}
	}
</style>
