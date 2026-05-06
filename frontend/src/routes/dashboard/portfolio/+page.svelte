<script lang="ts">
	const holdings = [
		{ symbol: 'AAPL', name: 'Apple Inc.', allocation: 22, value: 275000, day: 1.4 },
		{ symbol: 'MSFT', name: 'Microsoft Corp.', allocation: 18, value: 225000, day: 0.8 },
		{ symbol: 'NVDA', name: 'NVIDIA Corp.', allocation: 16, value: 200000, day: 2.1 },
		{ symbol: 'AMZN', name: 'Amazon.com', allocation: 12, value: 150000, day: -0.6 },
		{ symbol: 'GOOGL', name: 'Alphabet Inc.', allocation: 10, value: 125000, day: 0.5 }
	];
</script>

<svelte:head>
	<title>Portafolio - FINEXIA</title>
	<meta name="description" content="Detalle de posiciones y asignación de portafolio" />
</svelte:head>

<header class="page-header">
	<h1 class="page-title">Portafolio</h1>
	<p class="page-subtitle">Visión detallada de posiciones, asignación y rendimiento diario.</p>
</header>

<section class="cards-grid">
	<article class="panel summary">
		<p class="eyebrow">Valor total</p>
		<h2 class="hero-value">$1,250,000</h2>
		<p class="hero-delta positive">+3.7% este mes</p>
	</article>

	<article class="panel summary">
		<p class="eyebrow">Riesgo estimado</p>
		<h2 class="hero-value">Moderado</h2>
		<p class="hero-delta">Volatilidad 6.2%</p>
	</article>

	<article class="panel summary">
		<p class="eyebrow">Diversificación</p>
		<h2 class="hero-value">5 sectores</h2>
		<p class="hero-delta">Balance global</p>
	</article>
</section>

<section class="panel holdings">
	<header class="panel-header">
		<h2>Posiciones principales</h2>
		<span>Actualizado hace 5 min</span>
	</header>

	<div class="holdings-list">
		{#each holdings as holding (holding.symbol)}
			<article class="holding-row">
				<div class="holding-main">
					<p class="symbol">{holding.symbol}</p>
					<p class="name">{holding.name}</p>
				</div>
				<div class="bar-wrap" aria-label={`Asignación ${holding.allocation}%`}>
					<div class="bar-fill" style={`width: ${holding.allocation}%`}></div>
				</div>
				<p class="metric">${new Intl.NumberFormat('es-CO').format(holding.value)}</p>
				<p class={`metric delta ${holding.day >= 0 ? 'positive' : 'negative'}`}>
					{holding.day >= 0 ? '+' : ''}
					{holding.day}%
				</p>
			</article>
		{/each}
	</div>
</section>

<style>
	.page-header {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
		font-weight: 700;
		letter-spacing: 0.5px;
		color: #d4af37;
		font-family: 'Poppins', system-ui, sans-serif;
	}

	.page-subtitle {
		margin: 0;
		color: rgba(224, 224, 224, 0.62);
		font-size: 1rem;
	}

	.cards-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.panel {
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
	}

	.summary {
		padding: 1.35rem;
	}

	.eyebrow {
		margin: 0 0 0.55rem;
		font-size: 0.72rem;
		letter-spacing: 0.7px;
		text-transform: uppercase;
		color: rgba(224, 224, 224, 0.46);
	}

	.hero-value {
		margin: 0;
		font-size: 1.6rem;
		color: #e0e0e0;
	}

	.hero-delta {
		margin: 0.4rem 0 0;
		font-size: 0.82rem;
		color: rgba(224, 224, 224, 0.55);
	}

	.hero-delta.positive {
		color: #2ecc71;
	}

	.holdings {
		padding: 1.5rem;
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}

	.panel-header h2 {
		margin: 0;
		font-size: 1.2rem;
		color: #e0e0e0;
	}

	.panel-header span {
		font-size: 0.8rem;
		color: rgba(224, 224, 224, 0.52);
	}

	.holdings-list {
		display: grid;
		gap: 0.75rem;
	}

	.holding-row {
		display: grid;
		grid-template-columns: 1.25fr 1.4fr 0.8fr 0.45fr;
		align-items: center;
		gap: 0.9rem;
		padding: 0.85rem;
		border-radius: 12px;
		background: rgba(15, 20, 25, 0.45);
	}

	.holding-main .symbol {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 700;
		color: #e8c547;
	}

	.holding-main .name {
		margin: 0.15rem 0 0;
		font-size: 0.78rem;
		color: rgba(224, 224, 224, 0.6);
	}

	.bar-wrap {
		height: 8px;
		border-radius: 999px;
		background: rgba(224, 224, 224, 0.12);
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		border-radius: inherit;
		background: linear-gradient(90deg, #d4af37, #2ecc71);
	}

	.metric {
		margin: 0;
		font-size: 0.84rem;
		font-weight: 600;
		text-align: right;
		color: #e0e0e0;
	}

	.metric.delta.positive {
		color: #2ecc71;
	}

	.metric.delta.negative {
		color: #e74c3c;
	}

	@media (max-width: 1024px) {
		.cards-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.page-title {
			font-size: 1.85rem;
		}

		.holding-row {
			grid-template-columns: 1fr;
			gap: 0.55rem;
		}

		.metric {
			text-align: left;
		}
	}
</style>
