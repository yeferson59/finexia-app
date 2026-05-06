<script lang="ts">
	const assets = [
		{ name: 'Acciones', value: 450000, percent: 36, color: '#d4af37' },
		{ name: 'Bonos', value: 350000, percent: 28, color: '#2ecc71' },
		{ name: 'Fondos Mutuos', value: 250000, percent: 20, color: '#3498db' },
		{ name: 'Inmuebles', value: 150000, percent: 12, color: '#9b59b6' },
		{ name: 'Efectivo', value: 50000, percent: 4, color: '#e74c3c' }
	];

	function polarToCartesian(angle: number, radius: number, cx = 100, cy = 100) {
		const radians = (angle - 90) * (Math.PI / 180);
		return {
			x: cx + radius * Math.cos(radians),
			y: cy + radius * Math.sin(radians)
		};
	}

	function generatePieSlice(
		percent: number,
		startAngle: number,
	): { d: string; startAngle: number; endAngle: number } {
		const cx = 100;
		const cy = 100;
		const radius = 75;
		const endAngle = startAngle + (percent / 100) * 360;
		const largeArc = endAngle - startAngle > 180 ? 1 : 0;

		const startPoint = polarToCartesian(startAngle, radius, cx, cy);
		const endPoint = polarToCartesian(endAngle, radius, cx, cy);

		const d = [
			`M ${cx} ${cy}`,
			`L ${startPoint.x} ${startPoint.y}`,
			`A ${radius} ${radius} 0 ${largeArc} 1 ${endPoint.x} ${endPoint.y}`,
			'Z'
		].join(' ');

		return { d, startAngle, endAngle };
	}

	let currentAngle = 0;
	const slices = assets.map((asset) => {
		const slice = generatePieSlice(asset.percent, currentAngle);
		currentAngle = slice.endAngle;
		return { ...asset, ...slice };
	});
</script>

<div class="asset-card">
	<div class="card-header">
		<h2 class="card-title">Asignación de Activos</h2>
	</div>

	<div class="pie-container">
		<svg class="pie-chart" viewBox="0 0 200 200" preserveAspectRatio="xMidYMid meet">
			{#each slices as slice (slice.name)}
				<path d={slice.d} fill={slice.color} fill-opacity="0.85" stroke="rgba(15, 20, 25, 0.95)" stroke-width="2" />
			{/each}

			<!-- Center circle -->
			<circle cx="100" cy="100" r="45" fill="rgba(15, 20, 25, 0.95)" />
			<text x="100" y="98" text-anchor="middle" fill="#d4af37" font-size="22" font-weight="700" font-family="'Poppins', sans-serif">
				100%
			</text>
			<text x="100" y="113" text-anchor="middle" fill="rgba(224, 224, 224, 0.5)" font-size="9" font-family="'Lato', sans-serif">
				Diversificado
			</text>
		</svg>

		<div class="pie-legend">
			{#each assets as asset}
				<div class="legend-item">
					<div class="legend-color" style="background-color: {asset.color}"></div>
					<div class="legend-text">
						<p class="legend-label">{asset.name}</p>
						<p class="legend-value">${new Intl.NumberFormat('es-CO').format(asset.value)} ({asset.percent}%)</p>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<div class="card-footer">
		<button class="footer-button">Rebalancear portafolio</button>
	</div>
</div>

<style>
	.asset-card {
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.9) 0%, rgba(32, 39, 56, 0.9) 100%);
		border: 1px solid rgba(212, 175, 55, 0.15);
		border-radius: 16px;
		padding: 2rem;
		box-shadow: 
			0 20px 60px rgba(0, 0, 0, 0.3),
			inset 0 1px 0 rgba(255, 255, 255, 0.05);
		backdrop-filter: blur(16px);
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.card-header {
		margin-bottom: 1.5rem;
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

	.pie-container {
		flex: 1;
		display: flex;
		gap: 2rem;
		margin-bottom: 1.5rem;
		align-items: center;
	}

	.pie-chart {
		width: 150px;
		height: 150px;
		flex-shrink: 0;
		filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.3));
	}

	.pie-legend {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		border-radius: 8px;
		transition: background 0.25s ease;
	}

	.legend-item:hover {
		background: rgba(212, 175, 55, 0.1);
	}

	.legend-color {
		width: 12px;
		height: 12px;
		border-radius: 3px;
		flex-shrink: 0;
	}

	.legend-text {
		flex: 1;
		min-width: 0;
	}

	.legend-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: #e0e0e0;
		margin: 0;
		letter-spacing: 0.3px;
	}

	.legend-value {
		font-size: 0.7rem;
		color: rgba(224, 224, 224, 0.5);
		margin: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.card-footer {
		border-top: 1px solid rgba(212, 175, 55, 0.1);
		padding-top: 1.5rem;
	}

	.footer-button {
		width: 100%;
		padding: 0.875rem 1.5rem;
		background: transparent;
		border: 1.5px solid rgba(212, 175, 55, 0.2);
		color: #e0e0e0;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.25s ease;
		font-family: 'Poppins', system-ui, sans-serif;
		letter-spacing: 0.3px;
	}

	.footer-button:hover {
		background: rgba(212, 175, 55, 0.1);
		border-color: rgba(212, 175, 55, 0.3);
		color: #d4af37;
	}

	@media (max-width: 1024px) {
		.pie-container {
			flex-direction: column;
			gap: 1.5rem;
		}

		.pie-chart {
			width: 120px;
			height: 120px;
		}

		.pie-legend {
			gap: 0.5rem;
		}

		.legend-value {
			font-size: 0.65rem;
		}
	}

	@media (max-width: 768px) {
		.asset-card {
			padding: 1.5rem;
		}

		.pie-container {
			gap: 1rem;
		}

		.pie-chart {
			width: 100px;
			height: 100px;
		}

		.legend-item {
			padding: 0.5rem;
		}
	}
</style>
