<script lang="ts">
	import CardHeader from '$components/ui/card-header.svelte';

	const assets = [
		{ name: 'Acciones', value: 450000, percent: 36, color: '#d4912a' },
		{ name: 'Bonos', value: 350000, percent: 28, color: '#22c97e' },
		{ name: 'Fondos Mutuos', value: 250000, percent: 20, color: '#6b8cef' },
		{ name: 'Inmuebles', value: 150000, percent: 12, color: '#b988e0' },
		{ name: 'Efectivo', value: 50000, percent: 4, color: '#8a8780' }
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
		startAngle: number
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
	<CardHeader eyebrow="Distribución" title="Asignación de Activos" />

	<div class="pie-container">
		<svg class="pie-chart" viewBox="0 0 200 200" preserveAspectRatio="xMidYMid meet">
			{#each slices as slice (slice.name)}
				<path d={slice.d} fill={slice.color} fill-opacity="0.9" stroke="#08090a" stroke-width="2" />
			{/each}

			<!-- Center circle -->
			<circle cx="100" cy="100" r="45" fill="#08090a" />
			<text
				x="100"
				y="98"
				text-anchor="middle"
				fill="#e8a535"
				font-size="20"
				font-weight="600"
				font-family="'JetBrains Mono', monospace"
			>
				100%
			</text>
			<text
				x="100"
				y="114"
				text-anchor="middle"
				fill="#8a8780"
				font-size="8"
				letter-spacing="1"
				font-family="'JetBrains Mono', monospace"
			>
				DIVERSIFICADO
			</text>
		</svg>

		<div class="pie-legend">
			{#each assets as asset (asset.name)}
				<div class="legend-item">
					<div class="legend-color" style="background-color: {asset.color}"></div>
					<div class="legend-text">
						<p class="legend-label">{asset.name}</p>
						<p class="legend-value">
							${new Intl.NumberFormat('es-CO').format(asset.value)} ({asset.percent}%)
						</p>
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
		background: var(--surface);
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		padding: 2rem;
		backdrop-filter: blur(10px);
		display: flex;
		flex-direction: column;
		height: 100%;
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
		background: var(--surface);
	}

	.legend-color {
		width: 10px;
		height: 10px;
		border-radius: 3px;
		flex-shrink: 0;
	}

	.legend-text {
		flex: 1;
		min-width: 0;
	}

	.legend-label {
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--text);
		margin: 0;
	}

	.legend-value {
		font-family: var(--font-mono);
		font-size: 0.7rem;
		color: var(--text-dim);
		margin: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		font-variant-numeric: tabular-nums;
	}

	.card-footer {
		border-top: 1px solid var(--border);
		padding-top: 1.5rem;
	}

	.footer-button {
		width: 100%;
		padding: 0.75rem 1.5rem;
		background: transparent;
		border: 1px solid var(--border-strong);
		color: var(--text);
		border-radius: 6px;
		font-weight: 600;
		font-size: 0.85rem;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease,
			color 0.2s ease;
		font-family: var(--font-body);
	}

	.footer-button:hover {
		background: rgba(212, 145, 42, 0.06);
		border-color: rgba(212, 145, 42, 0.4);
		color: var(--amber-light);
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
