<script lang="ts">
	// Panel de marca de la izquierda (solo escritorio): logo, claim y el gráfico
	// decorativo animado. Sin estado ni props: es puramente presentacional.
</script>

<aside class="brand-panel">
	<div class="brand-content">
		<div class="brand-logo">
			<div class="logo-mark" aria-hidden="true">
				<svg width="28" height="28" viewBox="0 0 30 30" fill="none">
					<path
						d="M7 22L12.5 14.5L16.5 18.5L23 9"
						stroke="#0c0a06"
						stroke-width="2.6"
						stroke-linecap="round"
						stroke-linejoin="round"
					/>
				</svg>
			</div>
			<span class="brand-name">FINEXIA</span>
		</div>

		<p class="brand-tagline">Tu patrimonio,<br />bajo control.</p>

		<div class="brand-chart" aria-hidden="true">
			<svg viewBox="0 0 320 160" class="chart-svg" fill="none" xmlns="http://www.w3.org/2000/svg">
				<!-- Subtle grid lines -->
				<line x1="0" y1="40" x2="320" y2="40" stroke="rgba(255,255,255,0.04)" stroke-width="1" />
				<line x1="0" y1="80" x2="320" y2="80" stroke="rgba(255,255,255,0.04)" stroke-width="1" />
				<line x1="0" y1="120" x2="320" y2="120" stroke="rgba(255,255,255,0.04)" stroke-width="1" />
				<!-- Vertical tick marks -->
				<line x1="80" y1="0" x2="80" y2="160" stroke="rgba(255,255,255,0.025)" stroke-width="1" />
				<line x1="160" y1="0" x2="160" y2="160" stroke="rgba(255,255,255,0.025)" stroke-width="1" />
				<line x1="240" y1="0" x2="240" y2="160" stroke="rgba(255,255,255,0.025)" stroke-width="1" />

				<!-- Area fill under the line -->
				<path
					class="chart-area"
					d="M0,140 C30,130 55,118 80,105 C110,90 130,88 155,72 C180,56 205,48 230,35 C255,22 280,14 320,8 L320,160 L0,160 Z"
					fill="url(#area-grad)"
				/>

				<!-- Portfolio trend line -->
				<path
					class="chart-line"
					d="M0,140 C30,130 55,118 80,105 C110,90 130,88 155,72 C180,56 205,48 230,35 C255,22 280,14 320,8"
					stroke="url(#chart-grad)"
					stroke-width="2.5"
					stroke-linecap="round"
				/>

				<!-- End dot -->
				<circle class="chart-dot" cx="320" cy="8" r="4.5" fill="var(--amber)" />
				<circle
					class="chart-dot-ring"
					cx="320"
					cy="8"
					r="8"
					fill="none"
					stroke="var(--amber)"
					stroke-width="1"
				/>

				<defs>
					<linearGradient id="chart-grad" x1="0" y1="0" x2="1" y2="0">
						<stop offset="0%" stop-color="rgba(212,145,42,0.2)" />
						<stop offset="100%" stop-color="rgba(212,145,42,0.9)" />
					</linearGradient>
					<linearGradient id="area-grad" x1="0" y1="0" x2="0" y2="1">
						<stop offset="0%" stop-color="rgba(212,145,42,0.08)" />
						<stop offset="100%" stop-color="rgba(212,145,42,0)" />
					</linearGradient>
				</defs>
			</svg>
		</div>

		<p class="brand-footnote">Visualiza, analiza y optimiza<br />tu riqueza en tiempo real.</p>
	</div>
</aside>

<style>
	.logo-mark {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 52px;
		height: 52px;
		border-radius: 13px;
		background: var(--amber);
		border: 1px solid rgba(232, 165, 53, 0.4);
		box-shadow:
			0 0 25px rgba(212, 145, 42, 0.25),
			inset 0 1px 2px rgba(255, 255, 255, 0.2);
		flex-shrink: 0;
	}

	.brand-panel {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: clamp(3rem, 6vw, 5rem) clamp(2.5rem, 5vw, 4.5rem);
		background: #07080a;
		border-right: 1px solid var(--border);
		position: relative;
		overflow: hidden;
	}

	.brand-panel::before {
		content: '';
		position: absolute;
		top: -20%;
		right: -10%;
		width: 420px;
		height: 420px;
		border-radius: 50%;
		background: radial-gradient(circle, rgba(212, 145, 42, 0.06) 0%, transparent 65%);
		pointer-events: none;
	}

	.brand-content {
		position: relative;
		z-index: 1;
		width: 100%;
		max-width: 420px;
	}

	.brand-logo {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 2.5rem;
	}

	.brand-name {
		font-family: var(--font-display);
		font-size: clamp(2rem, 3.5vw, 3rem);
		font-weight: 600;
		letter-spacing: 0.12em;
		color: var(--text);
		line-height: 1;
	}

	.brand-tagline {
		font-family: var(--font-display);
		font-size: clamp(1.5rem, 2.8vw, 2.25rem);
		font-weight: 300;
		color: var(--text-muted);
		line-height: 1.4;
		margin: 0;
		letter-spacing: 0.01em;
	}

	.brand-chart {
		margin-top: 3rem;
		margin-bottom: 2rem;
	}

	.chart-svg {
		width: 100%;
		overflow: visible;
	}

	.chart-area {
		opacity: 0;
		animation: fade-area 0.6s 2s ease forwards;
	}

	.chart-line {
		stroke-dasharray: 700;
		stroke-dashoffset: 700;
		animation: draw-chart 1.8s 0.4s cubic-bezier(0.4, 0, 0.2, 1) forwards;
	}

	.chart-dot {
		opacity: 0;
		animation: fade-dot 0.4s 2.1s ease forwards;
	}

	.chart-dot-ring {
		opacity: 0;
		animation: pulse-ring 1.5s 2.3s ease-out infinite;
	}

	@keyframes draw-chart {
		to {
			stroke-dashoffset: 0;
		}
	}

	@keyframes fade-area {
		to {
			opacity: 1;
		}
	}

	@keyframes fade-dot {
		to {
			opacity: 1;
		}
	}

	@keyframes pulse-ring {
		0% {
			opacity: 0.6;
			transform: scale(1);
			transform-origin: 320px 8px;
		}
		100% {
			opacity: 0;
			transform: scale(2.2);
			transform-origin: 320px 8px;
		}
	}

	.brand-footnote {
		font-size: 0.8rem;
		color: var(--text-dim);
		line-height: 1.6;
		letter-spacing: 0.3px;
		margin: 0;
	}

	.brand-logo .logo-mark {
		width: 44px;
		height: 44px;
		border-radius: 11px;
		animation: logo-float 3s ease-in-out infinite;
	}

	@keyframes logo-float {
		0%,
		100% {
			transform: translateY(0px);
		}
		50% {
			transform: translateY(-4px);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.chart-line {
			animation: none;
			stroke-dashoffset: 0;
		}
		.chart-area,
		.chart-dot {
			animation: none;
			opacity: 1;
		}
		.chart-dot-ring {
			animation: none;
			opacity: 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.brand-logo .logo-mark {
			animation: none;
		}
	}

	@media (max-width: 768px) {
		.brand-panel {
			display: none;
		}
	}
</style>
