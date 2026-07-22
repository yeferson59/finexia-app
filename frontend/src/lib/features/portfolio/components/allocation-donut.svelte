<script lang="ts">
	import Card from '$lib/ui/card.svelte';
	import { computeDonutSegments, DONUT_RADIUS, type TypeBreakdownSlice } from '../portfolio';

	let {
		typeBreakdown,
		totalValue,
		formatCurrency
	}: {
		typeBreakdown: TypeBreakdownSlice[];
		totalValue: number;
		formatCurrency: (value: number) => string;
	} = $props();

	const donutSegments = $derived(computeDonutSegments(typeBreakdown));
</script>

<Card variant="elevated" padding="none">
	<div class="distribution">
		<header class="panel-header">
			<h2>Distribución por tipo</h2>
			<span>{typeBreakdown.length} {typeBreakdown.length === 1 ? 'tipo' : 'tipos'}</span>
		</header>

		<div class="distribution-body">
			<div class="donut-wrap">
				<svg
					class="donut"
					viewBox="0 0 160 160"
					role="img"
					aria-label="Distribución por tipo de activo"
				>
					<circle
						cx="80"
						cy="80"
						r={DONUT_RADIUS}
						fill="none"
						stroke="rgba(236, 234, 229, 0.08)"
						stroke-width="22"
					/>
					<g transform="rotate(-90 80 80)">
						{#each donutSegments as slice (slice.type)}
							<circle
								cx="80"
								cy="80"
								r={DONUT_RADIUS}
								fill="none"
								stroke={slice.color}
								stroke-width="22"
								stroke-linecap="round"
								stroke-dasharray={slice.dasharray}
								stroke-dashoffset={slice.dashoffset}
							>
								<title>{slice.label}: {slice.pct.toFixed(1)}%</title>
							</circle>
						{/each}
					</g>
					<text
						x="80"
						y="76"
						text-anchor="middle"
						fill="var(--text)"
						font-size="15"
						font-family="var(--font-mono)"
						font-weight="700">{formatCurrency(totalValue)}</text
					>
					<text
						x="80"
						y="94"
						text-anchor="middle"
						fill="rgba(236, 234, 229, 0.5)"
						font-size="9"
						font-family="var(--font-mono)">Valor total</text
					>
				</svg>
			</div>

			<div class="type-legend">
				{#each typeBreakdown as slice (slice.type)}
					<div class="legend-item">
						<span class="legend-dot" style="background: {slice.color};"></span>
						<span class="legend-label">{slice.label}</span>
						<span class="legend-pct">{slice.pct.toFixed(1)}%</span>
						<span class="legend-value">{formatCurrency(slice.value)}</span>
					</div>
				{/each}
			</div>
		</div>
	</div>
</Card>

<style>
	.distribution {
		padding: 1.5rem;
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1rem;
	}

	.panel-header h2 {
		margin: 0;
		font-size: 1.2rem;
		color: var(--text);
	}

	.panel-header span {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.52);
	}

	.distribution-body {
		display: flex;
		align-items: center;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.donut-wrap {
		flex-shrink: 0;
		width: 160px;
	}

	.donut {
		width: 100%;
		height: auto;
		display: block;
	}

	.donut circle {
		transition:
			stroke-dasharray 0.4s ease,
			opacity 0.2s ease;
	}

	.donut circle:hover {
		opacity: 0.85;
	}

	.type-legend {
		flex: 1;
		min-width: 220px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		gap: 0.55rem;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 0.55rem;
		padding: 0.55rem 0.7rem;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.022);
	}

	.legend-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.legend-label {
		flex: 1;
		font-size: 0.82rem;
		color: var(--text);
	}

	.legend-pct {
		font-size: 0.82rem;
		font-weight: 700;
		font-family: var(--font-mono);
		color: var(--text);
	}

	.legend-value {
		font-size: 0.75rem;
		color: rgba(236, 234, 229, 0.5);
		font-family: var(--font-mono);
		min-width: 4rem;
		text-align: right;
	}
</style>
