<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	interface Investment {
		id: string;
		name: string;
		type: string;
		category: string;
		description: string;
		riskLevel: string;
		expectedROI: number;
		currentROI: number;
		horizon: number;
		minimumInvestment: number;
		totalInvested: number;
		currentValue: number;
		investors: number;
		status: string;
		startDate: string;
		maturityDate: string;
		highlights: string[];
	}

	// Mock data - would come from backend
	const investmentDetails: Record<string, Investment> = {
		'1': {
			id: '1',
			name: 'Fondo Crecimiento Tecnológico',
			type: 'Fondo',
			category: 'Tecnología',
			description:
				'Fondo diversificado enfocado en empresas tecnológicas de alto crecimiento. Nuestro equipo de gestores expertos selecciona las mejores oportunidades en el sector tech global.',
			riskLevel: 'Medio',
			expectedROI: 15.2,
			currentROI: 12.8,
			horizon: 24,
			minimumInvestment: 5000,
			totalInvested: 2500000,
			currentValue: 2820000,
			investors: 342,
			status: 'Activo',
			startDate: '2023-01-15',
			maturityDate: '2025-12-31',
			highlights: [
				'Cartera diversificada en 15+ empresas tech líderes',
				'Gestor con 10+ años de experiencia',
				'Comisión de gestión competitiva (1.5% anual)',
				'Rebalanceo trimestral automático'
			]
		},
		'2': {
			id: '2',
			name: 'ETF Mercados Emergentes',
			type: 'ETF',
			category: 'Mercados Emergentes',
			description:
				'Exposición amplia a mercados emergentes de rápido crecimiento. Este ETF rastrea el desempeño de índices de economías emergentes con mayor potencial de apreciación.',
			riskLevel: 'Alto',
			expectedROI: 18.5,
			currentROI: 14.2,
			horizon: 36,
			minimumInvestment: 1000,
			totalInvested: 8750000,
			currentValue: 9980000,
			investors: 1204,
			status: 'Activo',
			startDate: '2022-06-10',
			maturityDate: '2026-06-10',
			highlights: [
				'Rastreo de índices de mercados emergentes',
				'Comisión ultra baja (0.35% anual)',
				'Liquidez diaria',
				'Diversificación en 25+ países'
			]
		}
	};

	let investment = $state<Investment | null>(null);

	$effect(() => {
		const id = page.params.id;
		investment = investmentDetails[id] || null;
	});

	function getRiskColor(risk: string): string {
		switch (risk) {
			case 'Bajo':
				return '#2ecc71';
			case 'Medio':
				return '#e8c547';
			case 'Alto':
				return '#e67e22';
			case 'Muy Alto':
				return '#e74c3c';
			default:
				return '#e0e0e0';
		}
	}

	function handleInvest() {
		alert('¡Pronto! Funcionalidad de inversión en desarrollo.');
	}

	function handleBack() {
		goto('/dashboard/investments');
	}

	const dateFormatter = new Intl.DateTimeFormat('es-CO', {
		year: 'numeric',
		month: 'long',
		day: 'numeric'
	});
</script>

<svelte:head>
	<title>{investment?.name} - FINEXIA</title>
	<meta name="description" content={investment?.description} />
</svelte:head>

<button class="back-button" onclick={handleBack} aria-label="Volver a inversiones">
	<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
	<section class="hero-section">
		<article class="metric-card" style="--metric-color: #d4af37">
			<p class="metric-label">ROI Esperado</p>
			<h2 class="metric-value">{investment.expectedROI}%</h2>
			<p class="metric-secondary">Proyectado {investment.horizon} meses</p>
		</article>

		<article class="metric-card" style="--metric-color: #2ecc71">
			<p class="metric-label">ROI Actual</p>
			<h2 class="metric-value">{investment.currentROI}%</h2>
			<p class="metric-secondary">Rendimiento desde inicio</p>
		</article>

		<article class="metric-card" style="--metric-color: #3498db">
			<p class="metric-label">Total Invertido</p>
			<h2 class="metric-value">${new Intl.NumberFormat('en-US', { notation: 'compact', compactDisplay: 'short' }).format(investment.totalInvested)}</h2>
			<p class="metric-secondary">{investment.investors} inversores</p>
		</article>

		<article class="metric-card" style="--metric-color: #9b59b6">
			<p class="metric-label">Valor Actual</p>
			<h2 class="metric-value">${new Intl.NumberFormat('en-US', { notation: 'compact', compactDisplay: 'short' }).format(investment.currentValue)}</h2>
			<p class="metric-secondary">+${new Intl.NumberFormat('en-US', { notation: 'compact', compactDisplay: 'short' }).format(investment.currentValue - investment.totalInvested)}</p>
		</article>
	</section>

	<!-- Description Section -->
	<section class="content-panel">
		<h2 class="section-title">Descripción del Producto</h2>
		<p class="description-text">{investment.description}</p>

		<button onclick={handleInvest} class="cta-button">Invertir Ahora</button>
	</section>

	<!-- Details Grid -->
	<section class="details-grid">
		<article class="detail-card">
			<h3 class="detail-title">Información Clave</h3>
			<div class="detail-list">
				<div class="detail-item">
					<span class="detail-label">Tipo de Instrumento</span>
					<span class="detail-value">{investment.type}</span>
				</div>
				<div class="detail-item">
					<span class="detail-label">Categoría</span>
					<span class="detail-value">{investment.category}</span>
				</div>
				<div class="detail-item">
					<span class="detail-label">Nivel de Riesgo</span>
					<span class="detail-value risk-badge" style={`--risk-color: ${getRiskColor(investment.riskLevel)}`}>
						{investment.riskLevel}
					</span>
				</div>
				<div class="detail-item">
					<span class="detail-label">Estado</span>
					<span class="detail-value status-badge">{investment.status}</span>
				</div>
			</div>
		</article>

		<article class="detail-card">
			<h3 class="detail-title">Parámetros de Inversión</h3>
			<div class="detail-list">
				<div class="detail-item">
					<span class="detail-label">Inversión Mínima</span>
					<span class="detail-value">${new Intl.NumberFormat('es-CO').format(investment.minimumInvestment)}</span>
				</div>
				<div class="detail-item">
					<span class="detail-label">Horizonte Temporal</span>
					<span class="detail-value">{investment.horizon} meses</span>
				</div>
				<div class="detail-item">
					<span class="detail-label">Fecha de Inicio</span>
					<span class="detail-value">{dateFormatter.format(new Date(investment.startDate))}</span>
				</div>
				<div class="detail-item">
					<span class="detail-label">Fecha de Vencimiento</span>
					<span class="detail-value">{dateFormatter.format(new Date(investment.maturityDate))}</span>
				</div>
			</div>
		</article>
	</section>

	<!-- Highlights Section -->
	<section class="content-panel highlights-section">
		<h2 class="section-title">Características Destacadas</h2>
		<ul class="highlights-list">
			{#each investment.highlights as highlight}
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
					stroke="#d4af37"
					stroke-width="2"
				/>
				<polyline
					points="0,120 50,100 100,80 150,60 200,75 250,50 300,65 350,40 400,35"
					fill="url(#gradient)"
					opacity="0.1"
				/>
				<defs>
					<linearGradient id="gradient" x1="0%" y1="0%" x2="0%" y2="100%">
						<stop offset="0%" style="stop-color:#d4af37;stop-opacity:0.3" />
						<stop offset="100%" style="stop-color:#d4af37;stop-opacity:0" />
					</linearGradient>
				</defs>
			</svg>
		</div>
	</section>
{:else}
	<section class="error-state">
		<h2>Producto no encontrado</h2>
		<p>Lo sentimos, no pudimos encontrar los detalles del producto solicitado.</p>
		<button onclick={handleBack} class="btn-back">Volver a Inversiones</button>
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
		border: 1.5px solid rgba(212, 175, 55, 0.25);
		border-radius: 8px;
		color: #d4af37;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.back-button:hover {
		background: rgba(212, 175, 55, 0.1);
		border-color: #d4af37;
	}

	.investment-header {
		margin-bottom: 2.5rem;
		padding-bottom: 2rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
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
		background: rgba(212, 175, 55, 0.1);
		border: 1px solid rgba(212, 175, 55, 0.25);
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
		color: #e0e0e0;
		font-family: 'Poppins', system-ui, sans-serif;
		letter-spacing: -0.5px;
	}

	.investment-type {
		margin: 0.5rem 0 0;
		font-size: 1rem;
		color: rgba(224, 224, 224, 0.6);
		font-weight: 500;
	}

	.hero-section {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 1.5rem;
		margin-bottom: 2rem;
		animation: fade-in 0.5s ease-out 0.1s both;
	}

	.metric-card {
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		padding: 1.5rem;
		text-align: center;
		border-top: 3px solid var(--metric-color);
	}

	.metric-label {
		margin: 0 0 0.6rem;
		font-size: 0.75rem;
		letter-spacing: 0.7px;
		text-transform: uppercase;
		color: rgba(224, 224, 224, 0.5);
	}

	.metric-value {
		margin: 0 0 0.4rem;
		font-size: 1.85rem;
		font-weight: 800;
		color: var(--metric-color);
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.metric-secondary {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(224, 224, 224, 0.5);
	}

	.content-panel {
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
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
		font-weight: 700;
		color: #e0e0e0;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.description-text {
		margin: 0 0 1.5rem;
		font-size: 1rem;
		line-height: 1.7;
		color: rgba(224, 224, 224, 0.75);
	}

	.cta-button {
		padding: 1rem 2rem;
		border: none;
		border-radius: 12px;
		background: linear-gradient(135deg, #d4af37, #e8c547);
		color: #0f1419;
		font-weight: 700;
		font-family: 'Poppins', system-ui, sans-serif;
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.3s ease;
		letter-spacing: 0.3px;
	}

	.cta-button:hover {
		transform: translateY(-3px);
		box-shadow: 0 15px 35px rgba(212, 175, 55, 0.3);
	}

	.details-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 2rem;
		margin-bottom: 2rem;
		animation: fade-in 0.5s ease-out 0.2s both;
	}

	.detail-card {
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		padding: 2rem;
	}

	.detail-title {
		margin: 0 0 1.5rem;
		font-size: 1.15rem;
		font-weight: 700;
		color: #e0e0e0;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.detail-list {
		display: grid;
		gap: 1.25rem;
	}

	.detail-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-bottom: 1rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.08);
	}

	.detail-item:last-child {
		padding-bottom: 0;
		border-bottom: none;
	}

	.detail-label {
		font-size: 0.9rem;
		color: rgba(224, 224, 224, 0.6);
		font-weight: 500;
	}

	.detail-value {
		font-size: 0.95rem;
		color: #e0e0e0;
		font-weight: 600;
		text-align: right;
	}

	.risk-badge {
		display: inline-block;
		padding: 0.35rem 0.75rem;
		border-radius: 6px;
		background: rgba(212, 175, 55, 0.1);
		color: var(--risk-color);
		font-weight: 700;
		font-size: 0.8rem;
	}

	.status-badge {
		display: inline-block;
		padding: 0.35rem 0.75rem;
		border-radius: 6px;
		background: rgba(46, 204, 113, 0.15);
		color: #2ecc71;
		font-weight: 700;
		font-size: 0.8rem;
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
		color: rgba(224, 224, 224, 0.75);
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
		background: rgba(212, 175, 55, 0.15);
		color: #2ecc71;
		font-weight: 700;
		font-size: 0.85rem;
	}

	.chart-placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 2rem 1rem;
		background: rgba(15, 20, 25, 0.5);
		border-radius: 12px;
		min-height: 200px;
		color: rgba(224, 224, 224, 0.5);
		animation: fade-in 0.5s ease-out 0.3s both;
	}

	.chart-line {
		width: 100%;
		max-width: 100%;
	}

	.error-state {
		text-align: center;
		padding: 3rem 1rem;
		border: 2px dashed rgba(212, 175, 55, 0.2);
		border-radius: 16px;
		background: rgba(26, 31, 46, 0.5);
	}

	.error-state h2 {
		color: #d4af37;
		font-family: 'Poppins', system-ui, sans-serif;
		margin-bottom: 1rem;
	}

	.error-state p {
		color: rgba(224, 224, 224, 0.6);
		margin-bottom: 1.5rem;
	}

	.btn-back {
		padding: 0.8rem 1.5rem;
		background: linear-gradient(135deg, #d4af37, #e8c547);
		color: #0f1419;
		border: none;
		border-radius: 8px;
		font-weight: 700;
		cursor: pointer;
		font-family: 'Poppins', system-ui, sans-serif;
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

	@media (max-width: 1024px) {
		.hero-section {
			grid-template-columns: repeat(2, 1fr);
		}

		.details-grid {
			grid-template-columns: 1fr;
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

		.hero-section {
			grid-template-columns: 1fr;
		}

		.content-panel {
			padding: 1.5rem;
		}

		.detail-card {
			padding: 1.5rem;
		}

		.detail-item {
			flex-direction: column;
			align-items: flex-start;
			gap: 0.5rem;
		}

		.detail-value {
			text-align: left;
		}
	}
</style>
