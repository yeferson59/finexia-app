<script lang="ts">
	/*
	 * Torta de concentración: cuánto pesa cada activo sobre todo lo que el
	 * usuario tiene, sin importar en qué portafolio esté.
	 *
	 * La porción y su entrada en la leyenda son el mismo dato —señalar cualquiera
	 * resalta la otra y el centro pasa a mostrar ese activo—, que es lo que
	 * permite leer un reparto de siete colores. La leyenda es una lista de
	 * botones: se recorre con el teclado y cada entrada dice su peso en voz alta,
	 * así que la identidad nunca depende solo del color.
	 */
	import Card from '$lib/ui/card.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCompactCurrency } from '$lib/shared/format/money';
	import { buildConcentrationSlices, type AssetHoldingRow } from '../asset-holdings';

	let {
		rows,
		displayCurrency,
		formatCurrency
	}: {
		rows: AssetHoldingRow[];
		/** Moneda de los importes: la misma en todas las filas. */
		displayCurrency: string;
		formatCurrency: (value: number) => string;
	} = $props();

	// Señalar y fijar se guardan aparte: con una sola variable, hacer clic sobre
	// la entrada que el ratón ya señalaba la apagaba.
	let hovered = $state<string | null>(null);
	let pinned = $state<string | null>(null);

	const activeKey = $derived(pinned ?? hovered);

	const slices = $derived(buildConcentrationSlices(rows));
	const total = $derived(slices.reduce((sum, slice) => sum + slice.value, 0));
	const active = $derived(slices.find((slice) => slice.key === activeKey) ?? null);

	function fmtPct(value: number): string {
		return new Intl.NumberFormat('es-CO', {
			minimumFractionDigits: 1,
			maximumFractionDigits: 1
		}).format(value);
	}
</script>

<Card variant="elevated" padding="none">
	<div class="concentration">
		<header class="panel-header">
			<h2>Concentración por activo</h2>
			<span>{rows.length} {rows.length === 1 ? 'activo' : 'activos'}</span>
		</header>

		{#if slices.length === 0}
			<EmptyState
				title="Aún no tienes activos"
				description="Registra una posición en cualquier portafolio y aquí verás cuánto pesa sobre el total."
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
			<div class="chart-body">
				<svg
					class="pie"
					viewBox="0 0 200 200"
					preserveAspectRatio="xMidYMid meet"
					aria-hidden="true"
				>
					<!-- El SVG entero es `aria-hidden`: la superficie accesible es la
					     leyenda de abajo, que sí es una lista de botones. -->
					{#each slices as slice (slice.key)}
						<path
							d={slice.d}
							fill={slice.color}
							class="slice"
							class:dimmed={activeKey !== null && activeKey !== slice.key}
							role="presentation"
							onpointerenter={() => (hovered = slice.key)}
							onpointerleave={() => (hovered = null)}
						>
							<title>{slice.label}: {fmtPct(slice.percent)}%</title>
						</path>
					{/each}

					<!-- Centro: el total, o el activo señalado. -->
					<circle cx="100" cy="100" r="52" fill="var(--bg)" />
					<!-- El total va abreviado: la cifra entera no cabe en el hueco, y
					     está completa en la tarjeta de arriba. -->
					<text x="100" y="96" text-anchor="middle" class="hole-value">
						{active
							? `${fmtPct(active.percent)}%`
							: privacy.money(formatCompactCurrency(total, displayCurrency))}
					</text>
					<text x="100" y="112" text-anchor="middle" class="hole-label">
						{active ? active.label.toUpperCase() : 'VALOR TOTAL'}
					</text>
				</svg>

				<ul class="legend">
					{#each slices as slice (slice.key)}
						<li>
							<button
								type="button"
								class="legend-item"
								class:active={activeKey === slice.key}
								aria-pressed={pinned === slice.key}
								onpointerenter={() => (hovered = slice.key)}
								onpointerleave={() => (hovered = null)}
								onfocus={() => (hovered = slice.key)}
								onblur={() => (hovered = null)}
								onclick={() => (pinned = pinned === slice.key ? null : slice.key)}
							>
								<span class="legend-color" style="background-color: {slice.color}"></span>
								<span class="legend-text">
									<span class="legend-label">
										{slice.label}
										{#if slice.assets > 1}
											<span class="legend-count">· {slice.assets} activos</span>
										{/if}
									</span>
									<span class="legend-value">
										{privacy.money(formatCurrency(slice.value))} ({fmtPct(slice.percent)}%)
									</span>
								</span>
							</button>
						</li>
					{/each}
				</ul>
			</div>
		{/if}
	</div>
</Card>

<style>
	.concentration {
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
		font-size: 1.15rem;
		color: var(--text);
	}

	.panel-header span {
		font-size: 0.8rem;
		color: rgba(236, 234, 229, 0.52);
	}

	.chart-body {
		display: flex;
		align-items: center;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.pie {
		width: clamp(160px, 22vw, 210px);
		height: clamp(160px, 22vw, 210px);
		flex-shrink: 0;
		filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.3));
	}

	/* El borde del color del fondo es el separador de 2px entre porciones: sin
	   él dos tonos contiguos se leen como una sola mancha. */
	.slice {
		fill-opacity: 0.92;
		stroke: var(--bg);
		stroke-width: 2;
		transition: fill-opacity 0.2s ease;
	}

	.slice.dimmed {
		fill-opacity: 0.26;
	}

	.hole-value {
		fill: var(--amber-light);
		font-size: 17px;
		font-weight: 600;
		font-family: var(--font-mono);
	}

	.hole-label {
		fill: #8a8780;
		font-size: 8px;
		letter-spacing: 1px;
		font-family: var(--font-mono);
	}

	/* En pantallas anchas la leyenda se reparte en columnas: en una sola, siete
	   entradas dejaban vacía media tarjeta y la torta parecía perdida. */
	.legend {
		flex: 1;
		min-width: 240px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: 0.2rem;
		list-style: none;
		margin: 0;
		padding: 0;
	}

	.legend-item {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 0.7rem;
		padding: 0.5rem 0.7rem;
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
		font-size: 0.82rem;
		font-weight: 600;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.legend-count {
		font-weight: 400;
		color: var(--text-muted);
	}

	.legend-value {
		display: block;
		font-family: var(--font-mono);
		font-size: 0.72rem;
		color: var(--text-dim);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		font-variant-numeric: tabular-nums;
	}

	@media (prefers-reduced-motion: reduce) {
		.slice {
			transition: none;
		}
	}

	@media (max-width: 900px) {
		.chart-body {
			flex-direction: column;
			gap: 1.25rem;
		}

		.legend {
			width: 100%;
		}
	}
</style>
