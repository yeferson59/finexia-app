<script lang="ts">
	import type { PageProps } from './$types';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	const { params }: PageProps = $props();

	interface AssetInfo {
		name: string;
		type: string;
		sector: string;
		icon: string;
		quantity: number;
		averageCost: number;
		currentPrice: number;
		dayChange: number;
		totalCost: number;
		totalValue: number;
		gainLoss: number;
		gainLossPercent: number;
		allocation: number;
		broker: string;
		riskLevel: string;
		volatility: number;
		beta: number;
	}

	// Mock asset data
	const assetData: Record<string, AssetInfo> = {
		AAPL: {
			name: 'Apple Inc.',
			type: 'Stock',
			sector: 'Technology',
			icon: '📱',
			quantity: 150,
			averageCost: 156.5,
			currentPrice: 183.75,
			dayChange: 1.4,
			totalCost: 23475,
			totalValue: 27562.5,
			gainLoss: 4087.5,
			gainLossPercent: 17.4,
			allocation: 22,
			broker: 'Interactive Brokers',
			riskLevel: 'Moderate',
			volatility: 4.8,
			beta: 1.2
		},
		MSFT: {
			name: 'Microsoft Corp.',
			type: 'Stock',
			sector: 'Technology',
			icon: '💻',
			quantity: 100,
			averageCost: 298.2,
			currentPrice: 345.6,
			dayChange: 0.8,
			totalCost: 29820,
			totalValue: 34560,
			gainLoss: 4740,
			gainLossPercent: 15.9,
			allocation: 18,
			broker: 'Interactive Brokers',
			riskLevel: 'Moderate',
			volatility: 4.2,
			beta: 0.95
		},
		BTC: {
			name: 'Bitcoin',
			type: 'Cryptocurrency',
			sector: 'Digital Assets',
			icon: '₿',
			quantity: 0.85,
			averageCost: 41200,
			currentPrice: 67450,
			dayChange: 3.2,
			totalCost: 35020,
			totalValue: 57332.5,
			gainLoss: 22312.5,
			gainLossPercent: 63.7,
			allocation: 28,
			broker: 'Kraken',
			riskLevel: 'High',
			volatility: 18.5,
			beta: 2.1
		},
		ETH: {
			name: 'Ethereum',
			type: 'Cryptocurrency',
			sector: 'Digital Assets',
			icon: '⟠',
			quantity: 5.2,
			averageCost: 2100,
			currentPrice: 3450,
			dayChange: 2.8,
			totalCost: 10920,
			totalValue: 17940,
			gainLoss: 7020,
			gainLossPercent: 64.3,
			allocation: 14,
			broker: 'Kraken',
			riskLevel: 'High',
			volatility: 16.2,
			beta: 1.95
		}
	};

	const asset = $derived(assetData[params.symbol] || assetData.AAPL);

	// Mock transaction history
	const transactions = [
		{ date: '2024-01-15', type: 'buy', quantity: 50, price: 142.5, total: 7125 },
		{ date: '2024-02-10', type: 'buy', quantity: 75, price: 158.2, total: 11865 },
		{ date: '2024-03-20', type: 'buy', quantity: 25, price: 175.3, total: 4382.5 }
	];

	function goBack() {
		goto(resolve('/dashboard/portfolios/[id]', { id: params.id }));
	}

	function handleSell() {
		alert(`Vender ${params.symbol} - Página en desarrollo`);
	}

	function handleEdit() {
		alert(`Editar ${params.symbol} - Página en desarrollo`);
	}

	function handleDelete() {
		if (confirm(`¿Deseas eliminar la posición de ${params.symbol}?`)) {
			goto(resolve('/dashboard/portfolios/[id]', { id: params.id }));
		}
	}
</script>

<svelte:head>
	<title>{params.symbol} - Portfolio - FINEXIA</title>
	<meta name="description" content="Detalles de la posición {params.symbol}" />
</svelte:head>

<div class="container">
	<!-- Header -->
	<div class="header-section">
		<button class="btn-back" onclick={goBack}>
			<svg
				width="20"
				height="20"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<path d="M15 19l-7-7 7-7" />
			</svg>
			Volver
		</button>

		<div class="header-content">
			<div class="symbol-badge">
				<span class="icon">{asset.icon}</span>
				<div class="symbol-info">
					<h1>{params.symbol}</h1>
					<p>{asset.name}</p>
				</div>
			</div>

			<div class="price-display">
				<p class="current-price">
					${asset.currentPrice.toLocaleString('es-CO', { minimumFractionDigits: 2 })}
				</p>
				<p class={`price-change ${asset.dayChange >= 0 ? 'positive' : 'negative'}`}>
					{asset.dayChange >= 0 ? '+' : ''}{asset.dayChange}% hoy
				</p>
			</div>
		</div>

		<div class="actions">
			<button class="btn btn-primary" onclick={handleSell}>Vender</button>
			<button class="btn btn-secondary" onclick={handleEdit}>Editar</button>
			<button class="btn btn-danger" onclick={handleDelete}>Eliminar</button>
		</div>
	</div>

	<!-- Holdings Summary -->
	<section class="panel holdings-summary">
		<header class="panel-header">
			<h2>Resumen de Posición</h2>
		</header>

		<div class="metrics-grid">
			<article class="metric-card">
				<p class="metric-label">Cantidad</p>
				<p class="metric-value">
					{asset.quantity}
					<span class="metric-unit">{params.symbol}</span>
				</p>
			</article>

			<article class="metric-card">
				<p class="metric-label">Costo Promedio</p>
				<p class="metric-value">
					${asset.averageCost.toLocaleString('es-CO', { maximumFractionDigits: 2 })}
				</p>
			</article>

			<article class="metric-card">
				<p class="metric-label">Precio Actual</p>
				<p class="metric-value">
					${asset.currentPrice.toLocaleString('es-CO', { minimumFractionDigits: 2 })}
				</p>
			</article>

			<article class="metric-card">
				<p class="metric-label">Costo Total</p>
				<p class="metric-value">
					${asset.totalCost.toLocaleString('es-CO', { maximumFractionDigits: 0 })}
				</p>
			</article>

			<article class="metric-card">
				<p class="metric-label">Valor Actual</p>
				<p class="metric-value">
					${asset.totalValue.toLocaleString('es-CO', { maximumFractionDigits: 0 })}
				</p>
			</article>

			<article class="metric-card gain">
				<p class="metric-label">Ganancia/Pérdida</p>
				<p class={`metric-value ${asset.gainLoss >= 0 ? 'positive' : 'negative'}`}>
					{asset.gainLoss >= 0 ? '+' : ''}${Math.abs(asset.gainLoss).toLocaleString('es-CO', {
						maximumFractionDigits: 0
					})}
				</p>
				<p class={`metric-pct ${asset.gainLoss >= 0 ? 'positive' : 'negative'}`}>
					{asset.gainLoss >= 0 ? '+' : ''}{asset.gainLossPercent.toFixed(1)}%
				</p>
			</article>
		</div>
	</section>

	<!-- Performance & Risk -->
	<section class="panel performance">
		<header class="panel-header">
			<h2>Métricas de Desempeño</h2>
		</header>

		<div class="perf-grid">
			<article class="perf-card">
				<h3>Tipo de Activo</h3>
				<p class="perf-value">{asset.type}</p>
			</article>

			<article class="perf-card">
				<h3>Sector</h3>
				<p class="perf-value">{asset.sector}</p>
			</article>

			<article class="perf-card">
				<h3>Asignación</h3>
				<p class="perf-value">{asset.allocation}%</p>
				<div class="bar-wrap">
					<div class="bar-fill" style={`width: ${asset.allocation}%`}></div>
				</div>
			</article>

			<article class="perf-card">
				<h3>Nivel de Riesgo</h3>
				<p class={`perf-value risk-${asset.riskLevel.toLowerCase()}`}>{asset.riskLevel}</p>
			</article>

			<article class="perf-card">
				<h3>Volatilidad</h3>
				<p class="perf-value">{asset.volatility.toFixed(1)}%</p>
			</article>

			<article class="perf-card">
				<h3>Beta</h3>
				<p class="perf-value">{asset.beta.toFixed(2)}</p>
			</article>

			<article class="perf-card">
				<h3>Broker</h3>
				<p class="perf-value">{asset.broker}</p>
			</article>

			<article class="perf-card">
				<h3>ROI</h3>
				<p class={`perf-value ${asset.gainLossPercent >= 0 ? 'positive' : 'negative'}`}>
					{asset.gainLossPercent.toFixed(1)}%
				</p>
			</article>
		</div>
	</section>

	<!-- Transaction History -->
	<section class="panel transactions">
		<header class="panel-header">
			<h2>Historial de Transacciones</h2>
			<span>{transactions.length} transacciones</span>
		</header>

		<div class="transactions-table">
			<div class="table-header">
				<p>Fecha</p>
				<p>Tipo</p>
				<p>Cantidad</p>
				<p>Precio</p>
				<p>Total</p>
			</div>

			{#each transactions as tx (tx.date + tx.type + tx.quantity)}
				<div class="table-row">
					<p class="date">{new Date(tx.date).toLocaleDateString('es-CO')}</p>
					<p class={`type type-${tx.type}`}>{tx.type === 'buy' ? 'Compra' : 'Venta'}</p>
					<p class="qty">{tx.quantity} {params.symbol}</p>
					<p class="price">${tx.price.toLocaleString('es-CO', { minimumFractionDigits: 2 })}</p>
					<p class="total">${tx.total.toLocaleString('es-CO', { maximumFractionDigits: 0 })}</p>
				</div>
			{/each}
		</div>
	</section>
</div>

<style>
	.container {
		--gold: var(--amber);
		--gold-accent: var(--amber-light);
		--dark-bg: #0d0800;
		--dark-secondary: #08090a;
		--dark-tertiary: #08090a;
		--text-primary: var(--text);
		--text-secondary: rgba(236, 234, 229, 0.6);
		--gold-border: var(--border-strong);

		max-width: 1400px;
		margin: 0 auto;
		padding: 0 1.5rem;
	}

	/* Header Section */
	.header-section {
		margin-bottom: 2rem;
		padding-bottom: 2rem;
		border-bottom: 1px solid var(--gold-border);
	}

	.btn-back {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		margin-bottom: 1.5rem;
		background: transparent;
		border: 1px solid var(--gold-border);
		color: var(--text-primary);
		font-family: var(--font-body);
		font-size: 0.9rem;
		cursor: pointer;
		border-radius: 8px;
		transition: all 0.3s ease;
	}

	.btn-back:hover {
		border-color: var(--gold-accent);
		color: var(--gold-accent);
		background: rgba(232, 165, 53, 0.05);
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 2rem;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
	}

	.symbol-badge {
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}

	.icon {
		font-size: 3.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.symbol-info h1 {
		margin: 0 0 0.25rem;
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--gold-accent);
		letter-spacing: -0.5px;
	}

	.symbol-info p {
		margin: 0;
		color: var(--text-secondary);
		font-size: 1rem;
	}

	.price-display {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		text-align: right;
	}

	.current-price {
		margin: 0 0 0.5rem;
		font-size: 1.75rem;
		font-weight: 600;
		color: var(--text-primary);
	}

	.price-change {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		margin: 0;
		font-size: 0.95rem;
		font-weight: 500;
	}

	.price-change.positive {
		color: var(--green);
	}

	.price-change.negative {
		color: var(--red);
	}

	.actions {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-family: var(--font-body);
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-primary {
		background: linear-gradient(135deg, var(--gold) 0%, var(--gold-accent) 100%);
		color: #0d0800;
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 24px rgba(212, 145, 42, 0.3);
	}

	.btn-secondary {
		background: transparent;
		border: 1px solid var(--gold-border);
		color: var(--text-primary);
	}

	.btn-secondary:hover {
		border-color: var(--gold-accent);
		background: rgba(232, 165, 53, 0.05);
		color: var(--gold-accent);
	}

	.btn-danger {
		background: transparent;
		border: 1px solid rgba(224, 90, 90, 0.3);
		color: var(--red);
	}

	.btn-danger:hover {
		border-color: var(--red);
		background: rgba(224, 90, 90, 0.1);
	}

	/* Panels */
	.panel {
		background: rgba(255, 255, 255, 0.03);
		backdrop-filter: blur(16px);
		border: 1px solid var(--gold-border);
		border-radius: 12px;
		padding: 1.75rem;
		margin-bottom: 1.5rem;
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--border);
	}

	.panel-header h2 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--gold-accent);
	}

	.panel-header span {
		color: var(--text-secondary);
		font-size: 0.85rem;
	}

	/* Holdings Summary */
	.metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		gap: 1rem;
	}

	.metric-card {
		background: rgba(212, 145, 42, 0.06);
		border: 1px solid rgba(212, 145, 42, 0.2);
		border-radius: 10px;
		padding: 1.25rem;
		text-align: center;
		transition: all 0.3s ease;
	}

	.metric-card:hover {
		border-color: rgba(212, 145, 42, 0.4);
		transform: translateY(-2px);
	}

	.metric-card.gain {
		border-color: rgba(212, 145, 42, 0.25);
	}

	.metric-label {
		margin: 0 0 0.75rem;
		font-size: 0.8rem;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 500;
	}

	.metric-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		margin: 0;
		font-size: 1.35rem;
		font-weight: 700;
		color: var(--gold-accent);
	}

	.metric-value.positive {
		color: var(--green);
	}

	.metric-value.negative {
		color: var(--red);
	}

	.metric-unit {
		font-size: 0.7rem;
		color: var(--text-secondary);
		margin-left: 0.25rem;
		font-weight: 500;
	}

	.metric-pct {
		margin: 0.25rem 0 0;
		font-size: 0.85rem;
		font-weight: 600;
	}

	.metric-pct.positive {
		color: var(--green);
	}

	.metric-pct.negative {
		color: var(--red);
	}

	/* Performance Grid */
	.perf-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 1rem;
	}

	.perf-card {
		background: rgba(255, 255, 255, 0.022);
		border: 1px solid var(--border);
		border-radius: 10px;
		padding: 1.25rem;
		text-align: center;
	}

	.perf-card h3 {
		margin: 0 0 0.75rem;
		font-size: 0.8rem;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.perf-value {
		margin: 0;
		font-size: 1.2rem;
		font-weight: 700;
		color: var(--gold-accent);
	}

	.perf-value.positive {
		color: var(--green);
	}

	.perf-value.negative {
		color: var(--red);
	}

	.risk-moderate {
		color: var(--amber-light);
	}

	.risk-high {
		color: var(--red);
	}

	.risk-low {
		color: var(--green);
	}

	.bar-wrap {
		width: 100%;
		height: 4px;
		background: var(--border);
		border-radius: 2px;
		margin-top: 0.5rem;
		overflow: hidden;
	}

	.bar-fill {
		height: 100%;
		background: linear-gradient(90deg, var(--gold) 0%, var(--gold-accent) 100%);
		border-radius: 2px;
		transition: width 0.3s ease;
	}

	/* Transaction History */
	.transactions-table {
		overflow-x: auto;
	}

	.table-header {
		display: grid;
		grid-template-columns: 100px 80px 100px 120px 120px;
		gap: 1rem;
		padding: 1rem;
		background: rgba(0, 0, 0, 0.2);
		border-radius: 8px 8px 0 0;
		border-bottom: 1px solid var(--gold-border);
		font-weight: 600;
		font-size: 0.85rem;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.3px;
		margin-bottom: 0;
	}

	.table-row {
		display: grid;
		grid-template-columns: 100px 80px 100px 120px 120px;
		gap: 1rem;
		padding: 1rem;
		border-bottom: 1px solid var(--border);
		align-items: center;
		transition: background 0.2s ease;
	}

	.table-row:hover {
		background: rgba(232, 165, 53, 0.03);
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.table-row p {
		margin: 0;
		font-size: 0.9rem;
	}

	.date {
		color: var(--text-secondary);
	}

	.type {
		font-weight: 600;
		text-transform: capitalize;
	}

	.type-buy {
		color: var(--green);
	}

	.type-sell {
		color: var(--red);
	}

	.qty {
		color: var(--text-primary);
		font-weight: 500;
	}

	.price {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		color: var(--gold-accent);
		font-weight: 500;
	}

	.total {
		color: var(--text-primary);
		font-weight: 600;
	}

	/* Responsive */
	@media (max-width: 1024px) {
		.header-content {
			flex-direction: column;
			align-items: flex-start;
		}

		.price-display {
			text-align: left;
		}

		.metrics-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.perf-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.table-header,
		.table-row {
			grid-template-columns: 80px 70px 90px 100px 100px;
		}
	}

	@media (max-width: 768px) {
		.header-section {
			margin-bottom: 1.5rem;
		}

		.symbol-badge {
			width: 100%;
			gap: 1rem;
		}

		.icon {
			font-size: 2.5rem;
		}

		.symbol-info h1 {
			font-size: 1.75rem;
		}

		.header-content {
			flex-direction: column;
		}

		.actions {
			width: 100%;
		}

		.btn {
			flex: 1;
		}

		.metrics-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.perf-grid {
			grid-template-columns: 1fr;
		}

		.table-header,
		.table-row {
			grid-template-columns: 70px 60px 80px 90px 90px;
			font-size: 0.8rem;
		}

		.table-row p {
			font-size: 0.8rem;
		}
	}

	@media (max-width: 480px) {
		.container {
			padding: 0 1rem;
		}

		.symbol-info h1 {
			font-size: 1.5rem;
		}

		.current-price {
			font-size: 1.4rem;
		}

		.metrics-grid {
			grid-template-columns: 1fr;
		}

		.actions {
			flex-direction: column;
		}

		.btn {
			width: 100%;
		}

		.table-header,
		.table-row {
			grid-template-columns: 1fr;
			gap: 0.5rem;
		}

		.table-header {
			display: none;
		}

		.table-row {
			background: rgba(255, 255, 255, 0.022);
			border: 1px solid var(--gold-border);
			border-radius: 8px;
			margin-bottom: 1rem;
			grid-template-columns: 1fr;
		}

		.table-row p::before {
			content: attr(data-label);
			font-weight: 600;
			color: var(--text-secondary);
			display: block;
			margin-bottom: 0.25rem;
		}
	}
</style>
