<script lang="ts">
	import { findInvestmentProduct, getRiskColor } from '../investments';
	import InvestmentMetrics from './investment-metrics.svelte';
	import InvestmentKeyFacts from './investment-key-facts.svelte';
	import type { InvestmentProduct } from '../investments';

	let { id, onBack }: { id: string; onBack: () => void } = $props();

	const investment = $derived<InvestmentProduct | null>(findInvestmentProduct(id));

	function handleInvest() {
		alert('¡Pronto! Funcionalidad de inversión en desarrollo.');
	}

	const dateFormatter = new Intl.DateTimeFormat('es-CO', {
		year: 'numeric',
		month: 'long',
		day: 'numeric'
	});
</script>

<button class="back-button" onclick={onBack} aria-label="Volver a inversiones">
	<svg
		width="20"
		height="20"
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="2"
	>
		<path d="M19 12H5M12 19l-7-7 7-7" />
	</svg>
	Volver
</button>

{#if investment}
	<header class="investment-header">
		<div class="header-content">
			<span class="category-badge" style={`--risk-color: ${getRiskColor(investment.riskLevel)}`}>
				{investment.riskLevel}
			</span>
			<h1 class="investment-title">{investment.name}</h1>
			<p class="investment-type">{investment.type} • {investment.category}</p>
		</div>
	</header>

	<!-- Hero Section with Key Metrics -->
	<InvestmentMetrics {investment} />

	<!-- Description Section -->
	<section class="content-panel">
		<h2 class="section-title">Descripción del Producto</h2>
		<p class="description-text">{investment.description}</p>

		<button onclick={handleInvest} class="cta-button">Invertir Ahora</button>
	</section>

	<!-- Details Grid -->
	<InvestmentKeyFacts {investment} {dateFormatter} />

	<!-- Highlights Section -->
	<section class="content-panel highlights-section">
		<h2 class="section-title">Características Destacadas</h2>
		<ul class="highlights-list">
			{#each investment.highlights as highlight (highlight)}
				<li class="highlight-item">
					<span class="highlight-icon">✓</span>
					{highlight}
				</li>
			{/each}
		</ul>
	</section>

	<!-- Performance Chart Placeholder -->
	<section class="content-panel">
		<h2 class="section-title">Rendimiento Histórico</h2>
		<div class="chart-placeholder">
			<p>Gráfico de rendimiento disponible en breve</p>
			<svg width="100%" height="150" viewBox="0 0 400 150" class="chart-line">
				<polyline
					points="0,120 50,100 100,80 150,60 200,75 250,50 300,65 350,40 400,35"
					fill="none"
					stroke="var(--amber)"
					stroke-width="2"
				/>
				<polyline
					points="0,120 50,100 100,80 150,60 200,75 250,50 300,65 350,40 400,35"
					fill="url(#gradient)"
					opacity="0.1"
				/>
				<defs>
					<linearGradient id="gradient" x1="0%" y1="0%" x2="0%" y2="100%">
						<stop offset="0%" style="stop-color:var(--amber);stop-opacity:0.3" />
						<stop offset="100%" style="stop-color:var(--amber);stop-opacity:0" />
					</linearGradient>
				</defs>
			</svg>
		</div>
	</section>
{:else}
	<section class="error-state">
		<h2>Producto no encontrado</h2>
		<p>Lo sentimos, no pudimos encontrar los detalles del producto solicitado.</p>
		<button onclick={onBack} class="btn-back">Volver a Inversiones</button>
	</section>
{/if}

<style>
	.back-button {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 1.5rem;
		padding: 0.65rem 1rem;
		background: transparent;
		border: 1.5px solid rgba(212, 145, 42, 0.25);
		border-radius: 8px;
		color: var(--amber);
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		font-family: var(--font-body);
	}

	.back-button:hover {
		background: var(--border);
		border-color: var(--amber);
	}

	.investment-header {
		margin-bottom: 2.5rem;
		padding-bottom: 2rem;
		border-bottom: 1px solid var(--border);
		animation: fade-in 0.5s ease-out;
	}

	.header-content {
		display: flex;
		align-items: flex-start;
		gap: 1.5rem;
		flex-wrap: wrap;
	}

	.category-badge {
		display: inline-block;
		padding: 0.6rem 1rem;
		border-radius: 20px;
		background: var(--border);
		border: 1px solid rgba(212, 145, 42, 0.25);
		color: var(--risk-color);
		font-size: 0.85rem;
		font-weight: 700;
		letter-spacing: 0.3px;
		text-transform: uppercase;
	}

	.investment-title {
		margin: 0;
		font-size: 2.8rem;
		font-weight: 800;
		color: var(--text);
		font-family: var(--font-body);
		letter-spacing: -0.5px;
	}

	.investment-type {
		margin: 0.5rem 0 0;
		font-size: 1rem;
		color: rgba(236, 234, 229, 0.6);
		font-weight: 500;
	}

	.content-panel {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		padding: 2rem;
		margin-bottom: 2rem;
		animation: fade-in 0.5s ease-out 0.15s both;
	}

	.section-title {
		margin: 0 0 1.5rem;
		font-size: 1.35rem;
		font-weight: 400;
		color: var(--text);
		font-family: var(--font-display);
	}

	.description-text {
		margin: 0 0 1.5rem;
		font-size: 1rem;
		line-height: 1.7;
		color: rgba(236, 234, 229, 0.75);
	}

	.cta-button {
		padding: 1rem 2rem;
		border: none;
		border-radius: 12px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 700;
		font-family: var(--font-body);
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
	}

	.cta-button:hover {
		transform: translateY(-3px);
		box-shadow: 0 15px 35px rgba(212, 145, 42, 0.3);
	}

	.highlights-section {
		animation: fade-in 0.5s ease-out 0.25s both;
	}

	.highlights-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: 1rem;
	}

	.highlight-item {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		font-size: 0.95rem;
		color: rgba(236, 234, 229, 0.75);
		line-height: 1.6;
	}

	.highlight-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		background: var(--border-strong);
		color: var(--green);
		font-weight: 700;
		font-size: 0.85rem;
	}

	.chart-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 2rem 1rem;
		background: rgba(255, 255, 255, 0.022);
		border-radius: 12px;
		min-height: 200px;
		color: rgba(236, 234, 229, 0.5);
		animation: fade-in 0.5s ease-out 0.3s both;
	}

	.chart-line {
		width: 100%;
		max-width: 100%;
	}

	.error-state {
		text-align: center;
		padding: 3rem 1rem;
		border: 2px dashed rgba(212, 145, 42, 0.2);
		border-radius: 16px;
		background: rgba(255, 255, 255, 0.03);
	}

	.error-state h2 {
		color: var(--amber);
		font-family: var(--font-body);
		margin-bottom: 1rem;
	}

	.error-state p {
		color: rgba(236, 234, 229, 0.6);
		margin-bottom: 1.5rem;
	}

	.btn-back {
		padding: 0.8rem 1.5rem;
		background: var(--amber);
		color: #0d0800;
		border: none;
		border-radius: 8px;
		font-weight: 700;
		cursor: pointer;
		font-family: var(--font-body);
	}

	@keyframes fade-in {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@media (max-width: 768px) {
		.investment-title {
			font-size: 2rem;
		}

		.header-content {
			flex-direction: column;
			gap: 1rem;
		}

		.content-panel {
			padding: 1.5rem;
		}
	}
</style>
