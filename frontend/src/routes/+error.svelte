<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';

	const is404 = $derived(page.status === 404);
</script>

<svelte:head>
	<title>{page.status} — Finexia</title>
</svelte:head>

<div class="error-page">
	<a href={resolve('/')} class="brand" aria-label="Finexia — Inicio">
		<svg width="28" height="28" viewBox="0 0 30 30" fill="none" aria-hidden="true">
			<rect width="30" height="30" rx="7" fill="var(--amber)" />
			<path
				d="M7 22L12.5 14.5L16.5 18.5L23 9"
				stroke="#0c0a06"
				stroke-width="2.6"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
		</svg>
		<span>FINEXIA</span>
	</a>

	<div class="chart-card">
		<div class="chart-header">
			<span class="chart-label">Portafolio principal</span>
			<span class="chart-tag-error">{is404 ? 'No encontrado' : 'Error del servidor'}</span>
		</div>

		<svg class="chart-svg" viewBox="0 0 420 130" fill="none" aria-hidden="true" role="presentation">
			<!-- Subtle grid lines -->
			<line x1="0" y1="25" x2="420" y2="25" stroke="rgba(255,255,255,0.03)" stroke-width="1" />
			<line x1="0" y1="65" x2="420" y2="65" stroke="rgba(255,255,255,0.03)" stroke-width="1" />
			<line x1="0" y1="106" x2="420" y2="106" stroke="rgba(255,255,255,0.05)" stroke-width="1" />

			<!-- Area fill under rise (amber glow) -->
			<path
				class="rise-area"
				d="M0,118 C30,110 60,94 95,78 S145,52 185,36 S220,22 250,18 L250,118 Z"
				fill="rgba(212,145,42,0.05)"
			/>

			<!-- Rise path — portfolio climbing (amber) -->
			<path
				class="rise-path"
				d="M0,118 C30,110 60,94 95,78 S145,52 185,36 S220,22 250,18"
				stroke="var(--amber)"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				pathLength="1"
			/>

			<!-- Crash path — sharp drop (red) -->
			<path
				class="crash-path"
				d="M250,18 L310,106"
				stroke="#e05a5a"
				stroke-width="2"
				stroke-linecap="round"
				pathLength="1"
			/>

			<!-- Flatline — stays at 404 level (dashed muted red) -->
			<path
				class="flat-path"
				d="M310,106 L420,106"
				stroke="rgba(224,90,90,0.4)"
				stroke-width="1.5"
				stroke-dasharray="5 7"
				stroke-linecap="round"
			/>

			<!-- Dot at crash landing -->
			<circle class="dot-crash" cx="310" cy="106" r="3.5" fill="#e05a5a" />

			<!-- 404 price-level annotation -->
			<g class="label-404">
				<rect
					x="363"
					y="94"
					width="55"
					height="22"
					rx="3"
					fill="rgba(224,90,90,0.1)"
					stroke="rgba(224,90,90,0.25)"
					stroke-width="1"
				/>
				<text
					x="390"
					y="109"
					font-family="'JetBrains Mono', monospace"
					font-size="11.5"
					font-weight="600"
					fill="#e05a5a"
					text-anchor="middle"
				>
					404
				</text>
			</g>
		</svg>
	</div>

	<div class="copy">
		<p class="eyebrow">Error &middot; {page.status}</p>
		<h1 class="headline">
			{#if is404}
				Esta página no está<br /><em>en tu mapa.</em>
			{:else}
				Algo salió<br /><em>inesperadamente</em> mal.
			{/if}
		</h1>
		<p class="desc">
			{#if is404}
				La ruta que buscas no existe o fue movida.<br />Vuelve al inicio para retomar el rumbo.
			{:else}
				Ocurrió un error interno en el servidor.<br />Inténtalo de nuevo en unos momentos.
			{/if}
		</p>
		<a href={resolve('/')} class="btn-back">Volver al inicio</a>
	</div>
</div>

<style>
	.error-page {
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 24px;
		gap: 32px;
	}

	/* Brand */
	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
		text-decoration: none;
		color: var(--text);
		font-family: var(--font-display);
		font-weight: 600;
		font-size: 18px;
		letter-spacing: 0.1em;
		margin-bottom: 4px;
	}

	/* Chart card */
	.chart-card {
		width: 100%;
		max-width: 480px;
		padding: 20px 20px 16px;
		border-radius: 12px;
		background: rgba(255, 255, 255, 0.038);
		border: 1px solid rgba(255, 255, 255, 0.12);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
	}

	.chart-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 14px;
	}

	.chart-label {
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.chart-tag-error {
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #e05a5a;
		background: rgba(224, 90, 90, 0.09);
		border: 1px solid rgba(224, 90, 90, 0.2);
		padding: 4px 9px;
		border-radius: 4px;
	}

	.chart-svg {
		width: 100%;
		height: auto;
		display: block;
		overflow: visible;
	}

	/* SVG path animations */
	.rise-area {
		opacity: 0;
		animation: fade-in 0.6s ease 0.3s forwards;
	}

	.rise-path {
		stroke-dasharray: 1;
		stroke-dashoffset: 1;
		animation: draw-path 1.2s ease-out 0.3s forwards;
	}

	.crash-path {
		stroke-dasharray: 1;
		stroke-dashoffset: 1;
		animation: draw-path 0.35s ease-in 1.4s forwards;
	}

	.flat-path {
		opacity: 0;
		animation: fade-in 0.4s ease 1.7s forwards;
	}

	.dot-crash {
		opacity: 0;
		animation: fade-in 0.3s ease 1.72s forwards;
	}

	.label-404 {
		opacity: 0;
		animation: fade-in 0.4s ease 1.82s forwards;
	}

	@keyframes draw-path {
		to {
			stroke-dashoffset: 0;
		}
	}

	@keyframes fade-in {
		to {
			opacity: 1;
		}
	}

	/* Copy section */
	.copy {
		text-align: center;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 16px;
		opacity: 0;
		animation: fade-in 0.7s ease 2.1s forwards;
	}

	.eyebrow {
		font-family: var(--font-mono);
		font-size: 11px;
		font-weight: 500;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--amber);
		margin: 0;
	}

	.headline {
		font-family: var(--font-display);
		font-weight: 300;
		font-size: clamp(28px, 5vw, 46px);
		line-height: 1.1;
		letter-spacing: -0.022em;
		margin: 0;
		text-wrap: balance;
	}

	.headline em {
		font-style: italic;
		font-weight: 500;
		color: var(--amber-light);
	}

	.desc {
		font-size: 15px;
		color: var(--text-muted);
		font-weight: 300;
		line-height: 1.7;
		margin: 0;
	}

	.btn-back {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 12px 22px;
		border-radius: 6px;
		border: none;
		cursor: pointer;
		font-family: var(--font-body);
		font-size: 14px;
		font-weight: 600;
		color: #0d0800;
		background: var(--amber);
		text-decoration: none;
		transition:
			background 0.2s,
			transform 0.15s;
		margin-top: 4px;
	}

	.btn-back:hover {
		background: var(--amber-light);
		transform: translateY(-1px);
	}

	.btn-back:active {
		transform: none;
	}

	/* Reduced motion */
	@media (prefers-reduced-motion: reduce) {
		.rise-area,
		.flat-path,
		.dot-crash,
		.label-404,
		.copy {
			opacity: 1;
			animation: none;
		}

		.rise-path,
		.crash-path {
			stroke-dashoffset: 0;
			animation: none;
		}
	}

	/* Mobile */
	@media (max-width: 480px) {
		.error-page {
			gap: 24px;
			padding: 36px 18px;
		}

		.chart-card {
			padding: 16px 16px 12px;
		}
	}
</style>
