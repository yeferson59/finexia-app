<script lang="ts">
	/*
	 * Donut de asignación por tipo de activo.
	 *
	 * La porción y su entrada en la leyenda son el mismo dato: señalar cualquiera
	 * de las dos resalta la otra y el centro del donut pasa a mostrar esa
	 * categoría, que es lo que faltaba para poder leer un reparto de ocho colores.
	 * La leyenda es una lista de botones, así que también se recorre con el
	 * teclado y cada entrada dice su porcentaje en voz alta.
	 */
	import { resolve } from '$app/paths';
	import CardHeader from '$lib/ui/card-header.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { buildSlices, toAssetEntries } from '../dashboard';
	import type { AllocationItem } from '$lib/api/types';

	const { allocation = [] }: { allocation: AllocationItem[] } = $props();

	/*
	 * Señalar y fijar son dos cosas distintas y se guardan aparte. Con una sola
	 * variable, hacer clic sobre la entrada que el ratón ya estaba señalando la
	 * apagaba: el puntero la había activado y el clic la alternaba a null.
	 */
	let hovered = $state<string | null>(null);
	let pinned = $state<string | null>(null);

	const activeName = $derived(pinned ?? hovered);

	const assets = $derived(toAssetEntries(allocation));
	const slices = $derived(buildSlices(assets));

	const totalPct = $derived(allocation.reduce((acc, item) => acc + item.percent, 0));
	const totalValue = $derived(assets.reduce((acc, asset) => acc + asset.value, 0));
	const active = $derived(assets.find((asset) => asset.name === activeName) ?? null);

	function fmtMoney(value: number): string {
		return privacy.money(
			'$' +
				new Intl.NumberFormat('es-CO', {
					minimumFractionDigits: 2,
					maximumFractionDigits: 2
				}).format(value)
		);
	}

	function fmtPct(value: number): string {
		return new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		}).format(value);
	}
</script>

<div class="asset-card">
	<CardHeader eyebrow="Distribución" title="Asignación de Activos" />

	{#if allocation.length === 0}
		<EmptyState
			title="Sin posiciones registradas"
			description="Añade activos a un portafolio y verás aquí cómo se reparte tu patrimonio."
		>
			{#snippet icon()}
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
			{/snippet}
		</EmptyState>
	{:else}
		<div class="pie-container">
			<svg
				class="pie-chart"
				viewBox="0 0 200 200"
				preserveAspectRatio="xMidYMid meet"
				aria-hidden="true"
			>
				{#each slices as slice (slice.name)}
					<!-- El SVG entero es `aria-hidden`: la superficie accesible es la
					     leyenda de abajo, que sí es una lista de botones. -->
					<path
						d={slice.d}
						fill={slice.color}
						class="slice"
						class:dimmed={activeName !== null && activeName !== slice.name}
						role="presentation"
						onpointerenter={() => (hovered = slice.name)}
						onpointerleave={() => (hovered = null)}
					/>
				{/each}

				<!-- Centro: el total, o la categoría señalada. -->
				<circle cx="100" cy="100" r="45" fill="#08090a" />
				<text x="100" y="98" text-anchor="middle" class="hole-value">
					{active ? fmtPct(active.percent) : Math.round(totalPct)}%
				</text>
				<text x="100" y="114" text-anchor="middle" class="hole-label">
					{active ? active.name.toUpperCase() : 'DIVERSIFICADO'}
				</text>
			</svg>

			<ul class="pie-legend">
				{#each assets as asset (asset.name)}
					<li>
						<button
							type="button"
							class="legend-item"
							class:active={activeName === asset.name}
							aria-pressed={pinned === asset.name}
							onpointerenter={() => (hovered = asset.name)}
							onpointerleave={() => (hovered = null)}
							onfocus={() => (hovered = asset.name)}
							onblur={() => (hovered = null)}
							onclick={() => (pinned = pinned === asset.name ? null : asset.name)}
						>
							<span class="legend-color" style="background-color: {asset.color}"></span>
							<span class="legend-text">
								<span class="legend-label">{asset.name}</span>
								<span class="legend-value">
									{fmtMoney(asset.value)} ({fmtPct(asset.percent)}%)
								</span>
							</span>
						</button>
					</li>
				{/each}
			</ul>
		</div>

		<p class="total-line">
			<span>Total repartido</span>
			<b>{fmtMoney(totalValue)}</b>
		</p>
	{/if}

	<div class="card-footer">
		<a href={resolve('/dashboard/portfolios')} class="footer-button">Gestionar portafolios</a>
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
		margin-bottom: 1rem;
		align-items: center;
	}

	.pie-chart {
		width: clamp(80px, 20vw, 150px);
		height: clamp(80px, 20vw, 150px);
		flex-shrink: 0;
		filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.3));
	}

	.slice {
		fill-opacity: 0.9;
		stroke: #08090a;
		stroke-width: 2;
		transition:
			fill-opacity 0.2s ease,
			transform 0.2s ease;
		transform-origin: 100px 100px;
	}

	.slice.dimmed {
		fill-opacity: 0.28;
	}

	.hole-value {
		fill: #e8a535;
		font-size: 20px;
		font-weight: 600;
		font-family: var(--font-mono);
	}

	.hole-label {
		fill: #8a8780;
		font-size: 8px;
		letter-spacing: 1px;
		font-family: var(--font-mono);
	}

	.pie-legend {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		list-style: none;
		margin: 0;
		padding: 0;
		min-width: 0;
	}

	.legend-item {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.55rem 0.75rem;
		border: 1px solid transparent;
		border-radius: 8px;
		background: transparent;
		color: inherit;
		font: inherit;
		text-align: left;
		cursor: pointer;
		transition:
			background 0.2s ease,
			border-color 0.2s ease;
	}

	.legend-item:hover,
	.legend-item.active {
		background: var(--surface-2);
		border-color: var(--border);
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
		display: block;
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.legend-value {
		display: block;
		font-family: var(--font-mono);
		font-size: 0.7rem;
		color: var(--text-dim);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		font-variant-numeric: tabular-nums;
	}

	.total-line {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
		margin: 0 0 1.5rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border);
		font-size: 0.75rem;
		color: var(--text-dim);
	}

	.total-line b {
		font-family: var(--font-mono);
		font-size: 0.9rem;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.card-footer {
		border-top: 1px solid var(--border);
		padding-top: 1.5rem;
		margin-top: auto;
	}

	.footer-button {
		display: block;
		width: 100%;
		padding: 0.75rem 1.5rem;
		background: transparent;
		border: 1px solid var(--border-strong);
		color: var(--text);
		border-radius: 6px;
		font-weight: 600;
		font-size: 0.85rem;
		text-align: center;
		text-decoration: none;
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

	@media (prefers-reduced-motion: reduce) {
		.slice {
			transition: none;
		}
	}

	@media (max-width: 1024px) {
		.pie-container {
			flex-direction: column;
			gap: 1.5rem;
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
