<script lang="ts">
	import type { InvestmentProduct } from '../investments';

	let { investment }: { investment: InvestmentProduct } = $props();
</script>

<section class="hero-section">
	<article class="metric-card" style="--metric-color: var(--amber)">
		<p class="metric-label">ROI Esperado</p>
		<h2 class="metric-value">{investment.expectedROI}%</h2>
		<p class="metric-secondary">Proyectado {investment.horizon} meses</p>
	</article>

	<article class="metric-card" style="--metric-color: var(--green)">
		<p class="metric-label">ROI Actual</p>
		<h2 class="metric-value">{investment.currentROI}%</h2>
		<p class="metric-secondary">Rendimiento desde inicio</p>
	</article>

	<article class="metric-card" style="--metric-color: #6b8cef">
		<p class="metric-label">Total Invertido</p>
		<h2 class="metric-value">
			${new Intl.NumberFormat('en-US', { notation: 'compact', compactDisplay: 'short' }).format(
				investment.totalInvested
			)}
		</h2>
		<p class="metric-secondary">{investment.investors} inversores</p>
	</article>

	<article class="metric-card" style="--metric-color: #b988e0">
		<p class="metric-label">Valor Actual</p>
		<h2 class="metric-value">
			${new Intl.NumberFormat('en-US', { notation: 'compact', compactDisplay: 'short' }).format(
				investment.currentValue
			)}
		</h2>
		<p class="metric-secondary">
			+${new Intl.NumberFormat('en-US', { notation: 'compact', compactDisplay: 'short' }).format(
				investment.currentValue - investment.totalInvested
			)}
		</p>
	</article>
</section>

<style>
	.hero-section {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 1.5rem;
		margin-bottom: 2rem;
		animation: fade-in 0.5s ease-out 0.1s both;
	}

	.metric-card {
		border: 1px solid var(--border-strong);
		border-radius: 16px;
		background: var(--surface);
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
		color: rgba(236, 234, 229, 0.5);
	}

	.metric-value {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
		margin: 0 0 0.4rem;
		font-size: 1.85rem;
		font-weight: 800;
		color: var(--metric-color);
		font-family: var(--font-body);
	}

	.metric-secondary {
		margin: 0;
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.5);
	}

	@media (max-width: 1024px) {
		.hero-section {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 768px) {
		.hero-section {
			grid-template-columns: 1fr;
		}
	}
</style>
