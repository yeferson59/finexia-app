<script lang="ts">
	const portfolioData = [
		{ month: 'Ene', value: 1100000 },
		{ month: 'Feb', value: 1150000 },
		{ month: 'Mar', value: 1120000 },
		{ month: 'Abr', value: 1180000 },
		{ month: 'May', value: 1205000 },
		{ month: 'Jun', value: 1250000 }
	];

	const maxValue = Math.max(...portfolioData.map(d => d.value));
	const minValue = Math.min(...portfolioData.map(d => d.value));
	const range = maxValue - minValue;
</script>

<div class="portfolio-card">
	<div class="card-header">
		<h2 class="card-title">Desempeño del Portafolio</h2>
		<div class="header-actions">
			<button class="period-button" class:active={true}>6M</button>
			<button class="period-button">1A</button>
			<button class="period-button">Todo</button>
		</div>
	</div>

	<div class="chart-container">
		<svg class="chart" viewBox="0 0 600 300" preserveAspectRatio="xMidYMid meet">
			<!-- Grid lines -->
			<defs>
				<linearGradient id="chartGradient" x1="0%" y1="0%" x2="0%" y2="100%">
					<stop offset="0%" style="stop-color: #d4af37; stop-opacity: 0.2" />
					<stop offset="100%" style="stop-color: #d4af37; stop-opacity: 0" />
				</linearGradient>
			</defs>

			<!-- Horizontal grid lines -->
			{#each Array.from({ length: 5 }) as _, i}
				<line
					x1="40"
					y1={40 + (i * 50)}
					x2="580"
					y2={40 + (i * 50)}
					stroke="rgba(212, 175, 55, 0.05)"
					stroke-width="1"
				/>
			{/each}

			<!-- Chart area -->
			<polyline
				points={portfolioData
					.map((d, i) => {
						const x = 40 + (i * 90);
						const y = 240 - ((d.value - minValue) / range) * 180;
						return `${x},${y}`;
					})
					.join(' ')}
				fill="url(#chartGradient)"
				stroke="none"
			/>

			<!-- Chart line -->
			<polyline
				points={portfolioData
					.map((d, i) => {
						const x = 40 + (i * 90);
						const y = 240 - ((d.value - minValue) / range) * 180;
						return `${x},${y}`;
					})
					.join(' ')}
				stroke="#d4af37"
				stroke-width="3"
				fill="none"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>

			<!-- Data points -->
			{#each portfolioData as d, i}
				<circle
					cx={40 + (i * 90)}
					cy={240 - ((d.value - minValue) / range) * 180}
					r="4"
					fill="#d4af37"
					stroke="rgba(15, 20, 25, 0.9)"
					stroke-width="2"
				/>
			{/each}

			<!-- X-axis labels -->
			{#each portfolioData as d, i}
				<text
					x={40 + (i * 90)}
					y="280"
					text-anchor="middle"
					fill="rgba(224, 224, 224, 0.5)"
					font-size="12"
					font-family="'Lato', sans-serif"
				>
					{d.month}
				</text>
			{/each}
		</svg>
	</div>

	<div class="chart-stats">
		<div class="stat">
			<span class="label">Ganancia YTD</span>
			<p class="value positive">+$150,000 (+13.6%)</p>
		</div>
		<div class="stat">
			<span class="label">Volatilidad</span>
			<p class="value">6.2%</p>
		</div>
		<div class="stat">
			<span class="label">Rentabilidad Anual</span>
			<p class="value highlight">7.4%</p>
		</div>
	</div>
</div>

<style>
	.portfolio-card {
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
		align-items: center;
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

	.header-actions {
		display: flex;
		gap: 0.5rem;
	}

	.period-button {
		padding: 0.5rem 1rem;
		background: transparent;
		border: 1px solid rgba(212, 175, 55, 0.1);
		color: rgba(224, 224, 224, 0.6);
		border-radius: 6px;
		font-weight: 600;
		font-size: 0.8rem;
		cursor: pointer;
		transition: all 0.25s ease;
		text-transform: uppercase;
		letter-spacing: 0.3px;
	}

	.period-button:hover {
		border-color: rgba(212, 175, 55, 0.3);
		color: #e0e0e0;
	}

	.period-button.active {
		background: rgba(212, 175, 55, 0.15);
		border-color: rgba(212, 175, 55, 0.3);
		color: #d4af37;
	}

	.chart-container {
		margin-bottom: 2rem;
		overflow-x: auto;
	}

	.chart {
		width: 100%;
		min-height: 300px;
		display: block;
	}

	.chart-stats {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 1.5rem;
		padding-top: 1.5rem;
		border-top: 1px solid rgba(212, 175, 55, 0.1);
	}

	.stat {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.label {
		font-size: 0.75rem;
		color: rgba(224, 224, 224, 0.5);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.value {
		font-size: 0.95rem;
		font-weight: 700;
		color: #e0e0e0;
		margin: 0;
	}

	.value.positive {
		color: #2ecc71;
	}

	.value.highlight {
		color: #d4af37;
	}

	@media (max-width: 1024px) {
		.card-header {
			flex-direction: column;
			gap: 1rem;
			align-items: flex-start;
		}

		.chart-stats {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 768px) {
		.portfolio-card {
			padding: 1.5rem;
		}

		.chart-stats {
			grid-template-columns: 1fr;
		}
	}
</style>
