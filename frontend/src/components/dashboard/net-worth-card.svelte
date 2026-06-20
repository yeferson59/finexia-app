<script lang="ts">
	let netWorth = $state(1250000);
	let monthlyChange = $state(45000);
	let monthlyChangePercent = $state(3.7);
	let isIncreasing = $derived(monthlyChange >= 0);
</script>

<div class="net-worth-card">
	<div class="card-header">
		<div>
			<p class="card-eyebrow">Patrimonio total</p>
			<h2 class="card-title">Patrimonio Neto</h2>
		</div>
		<span class="status-pill">Mes actual</span>
	</div>

	<div class="net-worth-content">
		<div class="main-metric">
			<h1 class="amount">
				${new Intl.NumberFormat('es-CO').format(netWorth)}
			</h1>
			<p class="amount-delta" class:positive={isIncreasing} class:negative={!isIncreasing}>
				{isIncreasing ? '+' : '−'}${new Intl.NumberFormat('es-CO').format(Math.abs(monthlyChange))}
				· {new Intl.NumberFormat('es-CO', { minimumFractionDigits: 1 }).format(
					monthlyChangePercent
				)}% este mes
			</p>
		</div>

		<div class="metric-stats">
			<div class="stat-item">
				<span class="stat-label">Clases de activo</span>
				<p class="stat-value">5</p>
			</div>

			<div class="stat-item">
				<span class="stat-label">Tasa promedio</span>
				<p class="stat-value highlight">7,4%<span class="stat-unit">anual</span></p>
			</div>

			<div class="stat-item">
				<span class="stat-label">Liquidez</span>
				<p class="stat-value positive">72h</p>
			</div>
		</div>
	</div>

	<div class="card-footer">
		<button class="action-button primary">Agregar fondos</button>
		<button class="action-button secondary">Ver detalles</button>
	</div>
</div>

<style>
	.net-worth-card {
		position: relative;
		overflow: hidden;
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		padding: 2rem;
		backdrop-filter: blur(10px);
	}

	/* Warm amber wash anchoring the hero figure */
	.net-worth-card::before {
		content: '';
		position: absolute;
		inset: 0;
		background: radial-gradient(
			ellipse 60% 90% at 0% 0%,
			rgba(212, 145, 42, 0.07),
			transparent 55%
		);
		pointer-events: none;
	}

	.net-worth-card > * {
		position: relative;
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		padding-bottom: 1.25rem;
		border-bottom: 1px solid var(--border);
	}

	.card-eyebrow {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		font-weight: 500;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin: 0 0 0.4rem 0;
	}

	.card-title {
		font-family: var(--font-display);
		font-size: 1.15rem;
		font-weight: 500;
		letter-spacing: -0.01em;
		color: var(--text);
		margin: 0;
	}

	.status-pill {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--text-muted);
		background: var(--surface-2);
		border: 1px solid var(--border);
		padding: 0.25rem 0.6rem;
		border-radius: 4px;
		white-space: nowrap;
	}

	.net-worth-content {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 3rem;
		margin-bottom: 2rem;
		align-items: end;
	}

	.main-metric {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.amount {
		font-family: var(--font-mono);
		font-size: clamp(2.25rem, 4.5vw, 3rem);
		font-weight: 600;
		color: var(--text);
		margin: 0;
		line-height: 1;
		letter-spacing: -0.03em;
		font-variant-numeric: tabular-nums;
	}

	.amount-delta {
		font-size: 0.85rem;
		font-weight: 400;
		margin: 0;
	}

	.amount-delta.positive {
		color: var(--green);
	}

	.amount-delta.negative {
		color: var(--red);
	}

	.metric-stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1.5rem;
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.stat-label {
		font-family: var(--font-mono);
		font-size: 0.625rem;
		color: var(--text-dim);
		text-transform: uppercase;
		letter-spacing: 0.12em;
		font-weight: 500;
	}

	.stat-value {
		display: flex;
		align-items: baseline;
		gap: 0.4rem;
		font-family: var(--font-mono);
		font-size: 1.15rem;
		font-weight: 600;
		color: var(--text);
		margin: 0;
		font-variant-numeric: tabular-nums;
	}

	.stat-value.positive {
		color: var(--green);
	}

	.stat-value.highlight {
		color: var(--amber-light);
	}

	.stat-unit {
		font-size: 0.7rem;
		font-weight: 400;
		color: var(--text-dim);
	}

	.card-footer {
		display: flex;
		gap: 0.75rem;
	}

	.action-button {
		flex: 1;
		padding: 0.75rem 1.5rem;
		border-radius: 6px;
		font-weight: 600;
		font-size: 0.85rem;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease,
			color 0.2s ease,
			transform 0.15s ease;
		font-family: var(--font-body);
	}

	.action-button.primary {
		border: none;
		background: var(--amber);
		color: #0d0800;
	}

	.action-button.primary:hover {
		background: var(--amber-light);
		transform: translateY(-1px);
	}

	.action-button.primary:active {
		transform: none;
	}

	.action-button.secondary {
		background: transparent;
		border: 1px solid var(--border-strong);
		color: var(--text);
	}

	.action-button.secondary:hover {
		background: rgba(212, 145, 42, 0.06);
		border-color: rgba(212, 145, 42, 0.4);
		color: var(--amber-light);
	}

	@media (prefers-reduced-motion: reduce) {
		.action-button.primary:hover {
			transform: none;
		}
	}

	@media (max-width: 1024px) {
		.net-worth-content {
			grid-template-columns: 1fr;
			gap: 2rem;
		}

		.metric-stats {
			grid-template-columns: repeat(3, 1fr);
		}
	}

	@media (max-width: 768px) {
		.net-worth-card {
			padding: 1.5rem;
		}

		.net-worth-content {
			gap: 1.5rem;
		}

		.action-button {
			padding: 0.75rem 1rem;
			font-size: 0.85rem;
		}
	}
</style>
