<script lang="ts">
	import CardHeader from '$lib/ui/card-header.svelte';
	import { privacy } from '$lib/stores/privacy.svelte';

	function fmtMoney(value: number): string {
		return privacy.money(
			'$' +
				new Intl.NumberFormat('es-CO', {
					minimumFractionDigits: 2,
					maximumFractionDigits: 2
				}).format(value)
		);
	}

	interface AllocationItem {
		category: string;
		marketValue: string;
		percent: number;
	}

	interface AssetEntry {
		name: string;
		value: number;
		percent: number;
		color: string;
	}

	const { allocation = [] }: { allocation: AllocationItem[] } = $props();

	const categoryLabels: Record<string, string> = {
		stocks: 'Acciones',
		etfs: 'ETFs',
		cryptos: 'Crypto',
		bonds: 'Bonos',
		cash: 'Efectivo',
		real_estates: 'Inmuebles',
		commodities: 'Commodities',
		others: 'Otros'
	};

	const categoryColors: Record<string, string> = {
		stocks: '#d4912a',
		etfs: '#22c97e',
		cryptos: '#6b8cef',
		bonds: '#b988e0',
		cash: '#8a8780',
		real_estates: '#e0885a',
		commodities: '#e0c15a',
		others: '#5ab4e0'
	};

	const assets = $derived<AssetEntry[]>(
		allocation.map((item) => ({
			name: categoryLabels[item.category] ?? item.category,
			value: parseFloat(item.marketValue || '0'),
			percent: item.percent,
			color: categoryColors[item.category] ?? '#5ab4e0'
		}))
	);

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

	function buildSlices(items: AssetEntry[]) {
		let angle = 0;
		return items.map((asset) => {
			const slice = generatePieSlice(asset.percent, angle);
			angle = slice.endAngle;
			return { ...asset, ...slice };
		});
	}

	const slices = $derived(buildSlices(assets));

	const totalPct = $derived(allocation.reduce((acc, item) => acc + item.percent, 0));
</script>

<div class="asset-card">
	<CardHeader eyebrow="Distribución" title="Asignación de Activos" />

	{#if allocation.length === 0}
		<div class="empty-state">
			<svg
				width="48"
				height="48"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="1.5"
			>
				<circle cx="12" cy="12" r="10" />
				<path d="M12 2a10 10 0 0 1 10 10" />
				<path d="M12 12L2 12" />
			</svg>
			<p>Sin posiciones registradas</p>
		</div>
	{:else}
		<div class="pie-container">
			<svg class="pie-chart" viewBox="0 0 200 200" preserveAspectRatio="xMidYMid meet">
				{#each slices as slice (slice.name)}
					<path
						d={slice.d}
						fill={slice.color}
						fill-opacity="0.9"
						stroke="#08090a"
						stroke-width="2"
					/>
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
					{Math.round(totalPct)}%
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
								{fmtMoney(asset.value)} ({new Intl.NumberFormat('es-CO', {
									minimumFractionDigits: 2,
									maximumFractionDigits: 2
								}).format(asset.percent)}%)
							</p>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

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

	.empty-state {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		padding: 2rem;
		color: rgba(236, 234, 229, 0.4);
		text-align: center;
	}

	.empty-state p {
		margin: 0;
		font-size: 0.9rem;
	}

	.pie-container {
		flex: 1;
		display: flex;
		gap: 2rem;
		margin-bottom: 1.5rem;
		align-items: center;
	}

	.pie-chart {
		width: clamp(80px, 20vw, 150px);
		height: clamp(80px, 20vw, 150px);
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
		overflow-wrap: anywhere;
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

		.legend-item {
			padding: 0.5rem;
		}
	}
</style>
