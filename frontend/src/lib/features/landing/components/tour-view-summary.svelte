<script lang="ts">
	/* Maqueta de `/dashboard`: patrimonio neto + gráfico de crecimiento. */
	import { TOUR_NET_WORTH, TOUR_GROWTH_SERIES } from '../product-tour';

	const W = 100;
	const H = 34;

	const line = TOUR_GROWTH_SERIES.map((v, i) => {
		const x = (i / (TOUR_GROWTH_SERIES.length - 1)) * W;
		const y = H - (v / 100) * H;
		return `${x.toFixed(2)},${y.toFixed(2)}`;
	}).join(' ');

	const area = `0,${H} ${line} ${W},${H}`;

	/* El capital invertido va por debajo del valor de mercado: la diferencia
	   entre las dos líneas es justo la ganancia que anuncia la tarjeta. */
	const costLine = TOUR_GROWTH_SERIES.map((v, i) => {
		const x = (i / (TOUR_GROWTH_SERIES.length - 1)) * W;
		const y = H - ((v * 0.62 + 6) / 100) * H;
		return `${x.toFixed(2)},${y.toFixed(2)}`;
	}).join(' ');

	const periods = ['1M', '6M', '1A', 'Todo'];
</script>

<div class="summary">
	<div class="mk-card net">
		<div class="net-top">
			<div>
				<div class="mk-eyebrow">Patrimonio total</div>
				<div class="net-value">{TOUR_NET_WORTH.total}</div>
				<div class="net-delta mk-up">{TOUR_NET_WORTH.delta}</div>
			</div>
			<div class="mk-pill">3 portafolios · 4 plataformas</div>
		</div>
		<div class="mk-stats">
			{#each TOUR_NET_WORTH.stats as stat, i (stat.label)}
				<div>
					<div class="mk-stat-label">{stat.label}</div>
					<div class="mk-stat-value" class:mk-up={i === 2} class:mk-amber={i === 1}>
						{stat.value}
					</div>
				</div>
			{/each}
		</div>
	</div>

	<div class="mk-card">
		<div class="chart-top">
			<div>
				<div class="mk-eyebrow">Portafolio</div>
				<div class="mk-title">Crecimiento del portafolio</div>
			</div>
			<div class="periods">
				{#each periods as p (p)}
					<span class:on={p === 'Todo'}>{p}</span>
				{/each}
			</div>
		</div>

		<svg class="chart" viewBox="0 0 {W} {H}" preserveAspectRatio="none">
			<defs>
				<linearGradient id="tourGrowth" x1="0" y1="0" x2="0" y2="1">
					<stop offset="0%" stop-color="rgba(212,145,42,0.22)" />
					<stop offset="100%" stop-color="rgba(212,145,42,0)" />
				</linearGradient>
			</defs>
			<polygon points={area} fill="url(#tourGrowth)" />
			<polyline
				points={costLine}
				fill="none"
				stroke="rgba(236,234,229,0.26)"
				stroke-width="0.6"
				stroke-dasharray="2 1.6"
				vector-effect="non-scaling-stroke"
			/>
			<polyline
				points={line}
				fill="none"
				stroke="var(--amber)"
				stroke-width="1"
				stroke-linecap="round"
				stroke-linejoin="round"
				vector-effect="non-scaling-stroke"
			/>
		</svg>

		<div class="legend">
			<span><i class="ln amber"></i>Valor de mercado</span>
			<span><i class="ln gray"></i>Capital invertido</span>
		</div>
	</div>
</div>

<style>
	.summary {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.net {
		position: relative;
		overflow: hidden;
	}

	.net::before {
		content: '';
		position: absolute;
		inset: 0;
		background: radial-gradient(
			ellipse 60% 90% at 0% 0%,
			rgba(212, 145, 42, 0.09),
			transparent 58%
		);
		pointer-events: none;
	}

	.net > * {
		position: relative;
	}

	.net-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 14px;
		padding-bottom: 14px;
		margin-bottom: 14px;
		border-bottom: 1px solid var(--border);
	}

	.net-value {
		margin-top: 6px;
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: clamp(22px, 3.4vw, 32px);
		line-height: 1;
		letter-spacing: -0.03em;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.net-delta {
		margin-top: 7px;
		font-size: 11.5px;
	}

	.chart-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
		padding-bottom: 12px;
		border-bottom: 1px solid var(--border);
	}

	.periods {
		display: flex;
		gap: 2px;
		padding: 2px;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: rgba(255, 255, 255, 0.03);
	}

	.periods span {
		padding: 3px 7px;
		border-radius: 4px;
		font-family: var(--font-mono);
		font-size: 9px;
		font-weight: 600;
		color: var(--text-dim);
	}

	.periods span.on {
		background: rgba(212, 145, 42, 0.18);
		color: var(--amber-light);
	}

	.chart {
		display: block;
		width: 100%;
		height: 118px;
		margin: 14px 0 10px;
	}

	.legend {
		display: flex;
		gap: 18px;
		font-family: var(--font-mono);
		font-size: 9px;
		color: var(--text-dim);
	}

	.legend span {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.ln {
		width: 16px;
		height: 0;
	}

	.ln.amber {
		border-top: 2px solid var(--amber);
	}

	.ln.gray {
		border-top: 1.5px dashed rgba(236, 234, 229, 0.28);
	}

	@media (max-width: 560px) {
		.chart {
			height: 92px;
		}

		.net-top {
			flex-direction: column;
			gap: 10px;
		}
	}
</style>
