<script lang="ts">
	import type { PageProps } from './$types';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	const { params }: PageProps = $props();

	const holdings = [
		{ symbol: 'AAPL', name: 'Apple Inc.', allocation: 22, value: 275000, day: 1.4 },
		{ symbol: 'MSFT', name: 'Microsoft Corp.', allocation: 18, value: 225000, day: 0.8 },
		{ symbol: 'NVDA', name: 'NVIDIA Corp.', allocation: 16, value: 200000, day: 2.1 },
		{ symbol: 'AMZN', name: 'Amazon.com', allocation: 12, value: 150000, day: -0.6 },
		{ symbol: 'GOOGL', name: 'Alphabet Inc.', allocation: 10, value: 125000, day: 0.5 }
	];

	function goBack() {
		goto(resolve('/dashboard/portfolios'));
	}

	function addAsset() {
		goto(resolve('/dashboard/portfolios/[id]/add', { id: params.id }));
	}

	function viewAssetDetails(symbol: string) {
		goto(resolve('/dashboard/portfolios/[id]/assets/[symbol]', { id: params.id, symbol }));
	}
</script>

<svelte:head>
	<title>Portafolio - FINEXIA</title>
	<meta name="description" content="Detalle de posiciones y asignación de portafolio" />
</svelte:head>

<header class="page-header">
	<div class="header-top">
		<div>
			<button onclick={goBack} class="btn-back" aria-label="Volver a portafolios">
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
			</button>
			<h1 class="page-title">{params.id}</h1>
			<p class="page-subtitle">Visión detallada de posiciones, asignación y rendimiento diario.</p>
		</div>
		<button onclick={addAsset} class="btn-add-asset">
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
			Agregar Activo
		</button>
	</div>
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
			<button
				class="holding-row"
				onclick={() => viewAssetDetails(holding.symbol)}
				aria-label={`View details for ${holding.symbol}`}
			>
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

	.header-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.page-title {
		margin: 0 0 0.5rem;
		font-size: 2.35rem;
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

	.btn-add-asset {
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

	.btn-add-asset:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.btn-back {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		padding: 0;
		margin-right: 1rem;
		border: 1px solid rgba(212, 145, 42, 0.3);
		border-radius: 8px;
		background: var(--border);
		color: var(--amber);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-back:hover {
		background: rgba(212, 145, 42, 0.2);
		border-color: rgba(212, 145, 42, 0.5);
		transform: translateX(-2px);
	}

	.header-top > div {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
	}

	.btn-add-asset {
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

	.btn-add-asset:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 25px rgba(212, 145, 42, 0.25);
	}

	.cards-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1rem;
		margin-bottom: 1.5rem;
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

	.summary {
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
		color: var(--text);
	}

	.panel-header span {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.52);
	}

	.holdings-list {
		display: grid;
		gap: 0.75rem;
	}

	.holding-row {
		display: grid;
		grid-template-columns: 1.25fr 1.4fr 0.8fr 0.45fr auto;
		align-items: center;
		gap: 0.9rem;
		padding: 0.85rem;
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.022);
		cursor: pointer;
		transition: all 0.3s ease;
		border: 1px solid rgba(212, 145, 42, 0);
	}

	.holding-row:hover {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(212, 145, 42, 0.2);
		transform: translateX(4px);
	}

	.holding-main .symbol {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 700;
		color: var(--amber-light);
	}

	.holding-main .name {
		margin: 0.15rem 0 0;
		font-size: 0.78rem;
		color: rgba(236, 234, 229, 0.6);
	}

	.bar-wrap {
		height: 8px;
		border-radius: 999px;
		background: rgba(236, 234, 229, 0.12);
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		border-radius: inherit;
		background: linear-gradient(90deg, var(--amber), var(--green));
	}

	.metric {
		margin: 0;
		font-size: 0.84rem;
		font-weight: 600;
		text-align: right;
		color: var(--text);
	}

	.metric.delta.positive {
		color: var(--green);
	}

	.metric.delta.negative {
		color: var(--red);
	}

	.arrow-icon {
		color: rgba(212, 145, 42, 0.4);
		transition: all 0.3s ease;
	}

	.holding-row:hover .arrow-icon {
		color: var(--amber-light);
		transform: translateX(2px);
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

		.header-top {
			flex-direction: column;
		}

		.btn-add-asset {
			width: 100%;
			justify-content: center;
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
