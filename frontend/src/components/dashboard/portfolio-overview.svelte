<script lang="ts">
	import Card from '$components/ui/card.svelte';
	import CardHeader from '$components/ui/card-header.svelte';
	import Stat from '$components/ui/stat.svelte';

	const portfolioData = [
		{ month: 'Ene', value: 1100000 },
		{ month: 'Feb', value: 1150000 },
		{ month: 'Mar', value: 1120000 },
		{ month: 'Abr', value: 1180000 },
		{ month: 'May', value: 1205000 },
		{ month: 'Jun', value: 1250000 }
	];

	const maxValue = Math.max(...portfolioData.map((d) => d.value));
	const minValue = Math.min(...portfolioData.map((d) => d.value));
	const range = maxValue - minValue;
</script>

<Card>
	<CardHeader eyebrow="Evolución" title="Desempeño del Portafolio">
		{#snippet action()}
			<div class="header-actions">
				<button class="period-button" class:active={true}>6M</button>
				<button class="period-button">1A</button>
				<button class="period-button">Todo</button>
			</div>
		{/snippet}
	</CardHeader>

	<div class="chart-container">
		<svg class="chart" viewBox="0 0 600 300" preserveAspectRatio="xMidYMid meet">
			<defs>
				<linearGradient id="chartGradient" x1="0%" y1="0%" x2="0%" y2="100%">
					<stop offset="0%" style="stop-color: #d4912a; stop-opacity: 0.22" />
					<stop offset="100%" style="stop-color: #d4912a; stop-opacity: 0" />
				</linearGradient>
				<linearGradient id="chartStroke" x1="0%" y1="0%" x2="100%" y2="0%">
					<stop offset="0%" style="stop-color: #22c97e" />
					<stop offset="100%" style="stop-color: #d4912a" />
				</linearGradient>
			</defs>

			<!-- Horizontal grid lines -->
			{#each Array.from({ length: 5 }) as _, i (i)}
				<line
					x1="40"
					y1={40 + i * 50}
					x2="580"
					y2={40 + i * 50}
					stroke="rgba(255, 255, 255, 0.04)"
					stroke-width="1"
				/>
			{/each}

			<!-- Chart area -->
			<polyline
				points={portfolioData
					.map((d, i) => {
						const x = 40 + i * 90;
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
						const x = 40 + i * 90;
						const y = 240 - ((d.value - minValue) / range) * 180;
						return `${x},${y}`;
					})
					.join(' ')}
				stroke="url(#chartStroke)"
				stroke-width="2.5"
				fill="none"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>

			<!-- Data points -->
			{#each portfolioData as d, i (d.month)}
				<circle
					cx={40 + i * 90}
					cy={240 - ((d.value - minValue) / range) * 180}
					r="3.5"
					fill="#d4912a"
					stroke="#08090a"
					stroke-width="2.5"
				/>
			{/each}

			<!-- X-axis labels -->
			{#each portfolioData as d, i (d.month)}
				<text
					x={40 + i * 90}
					y="280"
					text-anchor="middle"
					fill="#8a8780"
					font-size="11"
					font-family="'JetBrains Mono', monospace"
				>
					{d.month}
				</text>
			{/each}
		</svg>
	</div>

	<div class="chart-stats">
		<Stat label="Ganancia YTD" tone="positive" value="+$150.000 · +13,6%" />
		<Stat label="Volatilidad" value="6,2%" />
		<Stat label="Rentabilidad Anual" tone="highlight" value="7,4%" />
	</div>
</Card>

<style>
	.header-actions {
		display: flex;
		gap: 0.4rem;
		flex-shrink: 0;
	}

	.period-button {
		padding: 0.4rem 0.75rem;
		background: transparent;
		border: 1px solid var(--border);
		color: var(--text-muted);
		border-radius: 5px;
		font-family: var(--font-mono);
		font-weight: 500;
		font-size: 0.7rem;
		cursor: pointer;
		transition:
			border-color 0.2s ease,
			background 0.2s ease,
			color 0.2s ease;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	.period-button:hover {
		border-color: var(--border-strong);
		color: var(--text);
	}

	.period-button.active {
		background: rgba(212, 145, 42, 0.1);
		border-color: rgba(212, 145, 42, 0.3);
		color: var(--amber-light);
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
		border-top: 1px solid var(--border);
	}

	@media (max-width: 1024px) {
		.chart-stats {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 768px) {
		.chart-stats {
			grid-template-columns: 1fr;
		}
	}
</style>
