<script lang="ts">
	import { getRiskColor, type InvestmentProduct } from '../investments';

	let {
		investment,
		dateFormatter
	}: { investment: InvestmentProduct; dateFormatter: Intl.DateTimeFormat } = $props();
</script>

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
				<span
					class="detail-value risk-badge"
					style={`--risk-color: ${getRiskColor(investment.riskLevel)}`}
				>
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
				<span class="detail-value"
					>${new Intl.NumberFormat('es-CO').format(investment.minimumInvestment)}</span
				>
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

<style>
	.details-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 2rem;
		margin-bottom: 2rem;
		animation: fade-in 0.5s ease-out 0.2s both;
	}

	.detail-card {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
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
		color: var(--text);
		font-family: var(--font-body);
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
		border-bottom: 1px solid var(--border);
	}

	.detail-item:last-child {
		padding-bottom: 0;
		border-bottom: none;
	}

	.detail-label {
		font-size: 0.9rem;
		color: rgba(236, 234, 229, 0.6);
		font-weight: 500;
	}

	.detail-value {
		font-size: 0.95rem;
		color: var(--text);
		font-weight: 600;
		text-align: right;
	}

	.risk-badge {
		display: inline-block;
		padding: 0.35rem 0.75rem;
		border-radius: 6px;
		background: var(--border);
		color: var(--risk-color);
		font-weight: 700;
		font-size: 0.8rem;
	}

	.status-badge {
		display: inline-block;
		padding: 0.35rem 0.75rem;
		border-radius: 6px;
		background: rgba(34, 201, 126, 0.15);
		color: var(--green);
		font-weight: 700;
		font-size: 0.8rem;
	}

	@media (max-width: 1024px) {
		.details-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
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
