<script lang="ts">
	let netWorth = $state(1250000);
	let monthlyChange = $state(45000);
	let monthlyChangePercent = $state(3.7);
	let isIncreasing = $derived(monthlyChange >= 0);
</script>

<div class="net-worth-card">
	<div class="card-header">
		<h2 class="card-title">Patrimonio Neto</h2>
		<span class="period-label">Mes actual</span>
	</div>

	<div class="net-worth-content">
		<div class="main-metric">
			<span class="label">Tu patrimonio total</span>
			<h1 class="amount">
				${new Intl.NumberFormat('es-CO').format(netWorth)}
			</h1>
		</div>

		<div class="metric-stats">
			<div class="stat-item">
				<span class="stat-label">Cambio mensual</span>
				<p class="stat-value" class:positive={isIncreasing} class:negative={!isIncreasing}>
					<span class="direction-icon">
						{#if isIncreasing}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
								<path d="M12 2l7 7h-5v8h-4v-8H5l7-7z"></path>
							</svg>
						{:else}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
								<path d="M12 22l-7-7h5V7h4v8h5l-7 7z"></path>
							</svg>
						{/if}
					</span>
					${new Intl.NumberFormat('es-CO').format(Math.abs(monthlyChange))}
					<span class="percentage">({monthlyChangePercent}%)</span>
				</p>
			</div>

			<div class="stat-item">
				<span class="stat-label">Activos invertidos</span>
				<p class="stat-value highlight">5 clases</p>
			</div>

			<div class="stat-item">
				<span class="stat-label">Tasa promedio</span>
				<p class="stat-value highlight">7.4% anual</p>
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
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		padding: 2rem;
		box-shadow:
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
	}

	.card-title {
		font-size: 1.25rem;
		font-weight: 700;
		color: #e0e0e0;
		margin: 0;
		letter-spacing: 0.5px;
	}

	.period-label {
		font-size: 0.8rem;
		color: rgba(224, 224, 224, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.net-worth-content {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 3rem;
		margin-bottom: 2rem;
		align-items: start;
	}

	.main-metric {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.label {
		font-size: 0.85rem;
		color: rgba(224, 224, 224, 0.6);
		font-weight: 500;
		letter-spacing: 0.3px;
	}

	.amount {
		font-size: 2.75rem;
		font-weight: 700;
		color: #d4af37;
		margin: 0;
		line-height: 1.1;
		letter-spacing: -0.5px;
		font-family: 'Poppins', system-ui, sans-serif;
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
		font-size: 0.75rem;
		color: rgba(224, 224, 224, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.stat-value {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 1rem;
		font-weight: 700;
		color: #e0e0e0;
		margin: 0;
	}

	.stat-value.positive {
		color: #2ecc71;
	}

	.stat-value.negative {
		color: #e74c3c;
	}

	.stat-value.highlight {
		color: #d4af37;
	}

	.direction-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
	}

	.percentage {
		font-size: 0.85rem;
		opacity: 0.7;
	}

	.card-footer {
		display: flex;
		gap: 1rem;
	}

	.action-button {
		flex: 1;
		padding: 0.875rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.25s ease;
		font-family: 'Poppins', system-ui, sans-serif;
		letter-spacing: 0.3px;
	}

	.action-button.primary {
		background: linear-gradient(135deg, #d4af37 0%, #e8c547 100%);
		color: #0f1419;
		box-shadow: 0 4px 12px rgba(212, 175, 55, 0.2);
	}

	.action-button.primary:hover {
		background: linear-gradient(135deg, #e8c547 0%, #d4af37 100%);
		transform: translateY(-2px);
		box-shadow: 0 6px 20px rgba(212, 175, 55, 0.3);
	}

	.action-button.secondary {
		background: transparent;
		border: 1.5px solid rgba(212, 175, 55, 0.2);
		color: #e0e0e0;
	}

	.action-button.secondary:hover {
		background: rgba(212, 175, 55, 0.1);
		border-color: rgba(212, 175, 55, 0.3);
		color: #d4af37;
	}

	@media (max-width: 1024px) {
		.net-worth-content {
			grid-template-columns: 1fr;
			gap: 2rem;
		}

		.metric-stats {
			grid-template-columns: repeat(2, 1fr);
		}

		.amount {
			font-size: 2.25rem;
		}
	}

	@media (max-width: 768px) {
		.net-worth-card {
			padding: 1.5rem;
		}

		.net-worth-content {
			gap: 1.5rem;
		}

		.metric-stats {
			grid-template-columns: 1fr;
		}

		.amount {
			font-size: 1.875rem;
		}

		.action-button {
			padding: 0.75rem 1rem;
			font-size: 0.85rem;
		}
	}
</style>
