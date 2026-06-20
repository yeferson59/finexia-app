<script lang="ts">
	import { onMount } from 'svelte';

	let metricsEl: HTMLElement;

	onMount(() => {
		const io = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						metricsEl.classList.add('in');
						io.unobserve(entry.target);
					}
				});
			},
			{ threshold: 0.15 }
		);
		io.observe(metricsEl);
		return () => io.disconnect();
	});
</script>

<section class="block wrap">
	<div class="sec-head reveal">
		<div class="eyebrow">Métricas que importan</div>
		<h2 class="sec-title">Mira crecer tu patrimonio<br />con claridad</h2>
		<p class="sec-desc">
			Valor, peso, rendimiento y distribución de cada portafolio, en una vista hecha para entender
			de un vistazo.
		</p>
	</div>

	<div class="metrics-wrap reveal" bind:this={metricsEl}>
		<div class="metrics-top">
			<div class="metrics-headline">
				<div class="eyebrow" style="color: var(--green); text-align:left">Patrimonio total</div>
				<h2>$248.500 <span class="growth-badge">+12,4%</span></h2>
				<p>
					Sigue la evolución de tu patrimonio agregado y el peso de cada activo dentro de tus
					portafolios.
				</p>
			</div>
			<div class="metric-chips">
				<div class="mchip">
					<div class="lbl">Valor</div>
					<div class="val">$248,5K</div>
				</div>
				<div class="mchip">
					<div class="lbl">Rendimiento</div>
					<div class="val">+12,4%<span class="delta up">▲</span></div>
				</div>
				<div class="mchip">
					<div class="lbl">Peso mayor</div>
					<div class="val">38%<span class="delta amb">Acciones</span></div>
				</div>
			</div>
		</div>

		<div class="chart-zone">
			<svg class="chart-svg" viewBox="0 0 1000 280" preserveAspectRatio="none" aria-hidden="true">
				<defs>
					<linearGradient id="growthFill" x1="0" y1="0" x2="0" y2="1">
						<stop offset="0%" stop-color="rgba(212,145,42,0.2)" />
						<stop offset="100%" stop-color="rgba(212,145,42,0)" />
					</linearGradient>
					<linearGradient id="growthStroke" x1="0" y1="0" x2="1" y2="0">
						<stop offset="0%" stop-color="#22c97e" />
						<stop offset="100%" stop-color="#d4912a" />
					</linearGradient>
				</defs>
				<g class="chart-grid">
					<line x1="0" y1="40" x2="1000" y2="40" />
					<line x1="0" y1="110" x2="1000" y2="110" />
					<line x1="0" y1="180" x2="1000" y2="180" />
					<line x1="0" y1="250" x2="1000" y2="250" />
				</g>
				<path
					class="chart-area"
					d="M0,230 C90,225 150,210 230,200 C320,188 360,170 450,150 C540,130 590,140 670,110 C760,78 820,72 900,52 C935,42 960,34 982,26 L982,280 L0,280 Z"
				/>
				<path
					class="chart-line"
					d="M0,230 C90,225 150,210 230,200 C320,188 360,170 450,150 C540,130 590,140 670,110 C760,78 820,72 900,52 C935,42 960,34 982,26"
				/>
				<g class="chart-dot">
					<circle class="pulse-ring" cx="982" cy="26" r="6" fill="rgba(212,145,42,0.35)" />
					<circle cx="982" cy="26" r="5" fill="var(--amber)" stroke="#08090a" stroke-width="2.5" />
				</g>
			</svg>
			<div class="chart-flag">
				<div class="t">Este año</div>
				<div class="v">+ $27.400</div>
			</div>
		</div>
	</div>
</section>

<style>
	.metrics-wrap {
		position: relative;
		border-radius: 14px;
		overflow: hidden;
		border: 1px solid var(--border-strong);
		background: var(--surface);
		padding: 44px 44px 0;
	}
	.metrics-top {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 28px;
		flex-wrap: wrap;
		position: relative;
		z-index: 2;
	}
	.metrics-headline {
		max-width: 30ch;
	}
	.metrics-headline h2 {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: clamp(24px, 3.2vw, 36px);
		line-height: 1.06;
		letter-spacing: -0.02em;
		margin-top: 10px;
		font-variant-numeric: tabular-nums;
	}
	.growth-badge {
		color: var(--green);
		font-size: 0.52em;
		font-weight: 600;
		letter-spacing: 0;
	}
	.metrics-headline p {
		margin-top: 12px;
		font-size: 14.5px;
		color: var(--text-muted);
		font-weight: 300;
	}
	.metric-chips {
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}
	.mchip {
		padding: 14px 18px;
		border-radius: 8px;
		min-width: 118px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid var(--border);
	}
	.mchip .lbl {
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 500;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-dim);
	}
	.mchip .val {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: 21px;
		letter-spacing: -0.02em;
		margin-top: 6px;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}
	.delta {
		font-size: 12px;
		font-weight: 600;
		margin-left: 6px;
	}
	.delta.up {
		color: var(--green);
	}
	.delta.amb {
		color: var(--amber);
	}

	/* CHART */
	.chart-zone {
		position: relative;
		margin-top: 32px;
		height: 260px;
	}
	.chart-svg {
		display: block;
		width: 100%;
		height: 100%;
	}
	.chart-grid line {
		stroke: rgba(255, 255, 255, 0.04);
		stroke-width: 1;
	}
	.chart-area {
		fill: url(#growthFill);
		opacity: 0;
		transition: opacity 1.2s ease 0.5s;
	}
	:global(.metrics-wrap.in) .chart-area {
		opacity: 1;
	}
	.chart-line {
		fill: none;
		stroke: url(#growthStroke);
		stroke-width: 2;
		stroke-linecap: round;
		stroke-linejoin: round;
		stroke-dasharray: 1400;
		stroke-dashoffset: 1400;
		transition: stroke-dashoffset 1.8s cubic-bezier(0.4, 0, 0.2, 1) 0.2s;
	}
	:global(.metrics-wrap.in) .chart-line {
		stroke-dashoffset: 0;
	}
	.chart-dot {
		opacity: 0;
		transition: opacity 0.4s ease 1.6s;
	}
	:global(.metrics-wrap.in) .chart-dot {
		opacity: 1;
	}
	.pulse-ring {
		transform-box: fill-box;
		transform-origin: center;
		animation: dotpulse 2.2s ease-out infinite;
	}
	@keyframes dotpulse {
		0% {
			transform: scale(0.6);
			opacity: 0.7;
		}
		70% {
			transform: scale(2.4);
			opacity: 0;
		}
		100% {
			opacity: 0;
		}
	}
	.chart-flag {
		position: absolute;
		top: 8%;
		right: 4%;
		padding: 8px 12px;
		border-radius: 8px;
		background: rgba(8, 9, 10, 0.88);
		border: 1px solid rgba(34, 201, 126, 0.28);
		backdrop-filter: blur(8px);
		opacity: 0;
		transform: translateY(6px);
		transition:
			opacity 0.5s ease 1.5s,
			transform 0.5s ease 1.5s;
	}
	:global(.metrics-wrap.in) .chart-flag {
		opacity: 1;
		transform: none;
	}
	.chart-flag .t {
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--text-dim);
	}
	.chart-flag .v {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: 17px;
		color: var(--green);
		margin-top: 3px;
	}
	@media (max-width: 620px) {
		.metrics-wrap {
			padding: 28px 20px 0;
		}
		.chart-zone {
			height: 200px;
		}
		.mchip {
			min-width: 0;
			flex: 1;
			padding: 12px 14px;
		}
	}
</style>
