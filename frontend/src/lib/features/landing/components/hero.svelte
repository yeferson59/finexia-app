<script lang="ts">
	import WaitlistForm from './waitlist-form.svelte';
	import HeroMap from './hero-map.svelte';
</script>

<section class="hero-shell">
	<!--
		Atmósfera del hero. Sustituye a `.scanlines`, que a 0.006 de alfa no se
		veía en ninguna pantalla: una retícula de mapa enmascarada hacia los
		bordes y un halo ámbar en la esquina de entrada.
	-->
	<div class="hero-grid" aria-hidden="true"></div>
	<div class="hero-glow" aria-hidden="true"></div>

	<div class="hero wrap">
		<div class="hero-left">
			<div class="hero-badge reveal">
				<span class="dot"></span>
				Acceso anticipado · Lanzamiento 1 oct 2026
			</div>

			<h1 class="hero-title reveal">
				Todo tu patrimonio,<br /><em>en tu mapa.</em>
			</h1>
			<p class="hero-sub reveal">
				Registra manualmente dónde tienes tus activos y agrúpalos en los portafolios que tú
				imaginas, aunque estén en distintas plataformas. Sin conectar cuentas, sin dar acceso a
				nadie.
			</p>

			<WaitlistForm anchor="waitlist" />

			<ul class="hero-proof reveal">
				<li>Sin credenciales de tu broker</li>
				<li>Sin tarjeta</li>
				<li>Tus datos, exportables</li>
			</ul>
		</div>

		<HeroMap />
	</div>
</section>

<style>
	.hero-shell {
		position: relative;
		overflow: hidden;
	}

	.hero-grid {
		position: absolute;
		inset: 0;
		background-image:
			linear-gradient(rgba(255, 255, 255, 0.028) 1px, transparent 1px),
			linear-gradient(90deg, rgba(255, 255, 255, 0.028) 1px, transparent 1px);
		background-size: 72px 72px;
		-webkit-mask-image: radial-gradient(ellipse 62% 78% at 68% 38%, #000 0%, transparent 72%);
		mask-image: radial-gradient(ellipse 62% 78% at 68% 38%, #000 0%, transparent 72%);
		pointer-events: none;
	}

	.hero-glow {
		position: absolute;
		inset: 0;
		background: radial-gradient(
			ellipse 46% 62% at 18% 6%,
			rgba(212, 145, 42, 0.1),
			transparent 62%
		);
		pointer-events: none;
	}

	.hero {
		position: relative;
		display: grid;
		grid-template-columns: 55fr 45fr;
		gap: 56px;
		align-items: center;
		padding-block: 84px 92px;
	}

	.hero-left {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
	}

	.hero-badge {
		display: inline-flex;
		align-items: center;
		gap: 9px;
		padding: 6px 14px;
		border: 1px solid var(--border-strong);
		border-radius: 999px;
		background: var(--surface);
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--text-muted);
	}

	.hero-badge .dot {
		width: 6px;
		height: 6px;
		flex-shrink: 0;
		border-radius: 999px;
		background: var(--amber);
	}

	.hero-title {
		font-family: var(--font-display);
		font-weight: 300;
		font-size: clamp(42px, 5.6vw, 76px);
		line-height: 1.06;
		letter-spacing: -0.022em;
		margin: 26px 0 0;
	}

	.hero-title em {
		font-style: italic;
		font-weight: 500;
		color: var(--amber-light);
	}

	.hero-sub {
		margin: 20px 0 0;
		max-width: 46ch;
		font-size: clamp(15px, 1.6vw, 17px);
		color: var(--text-muted);
		font-weight: 300;
		line-height: 1.7;
		text-wrap: pretty;
	}

	/*
	 * El formulario es lo único grande que queda bajo el titular: con la cuenta
	 * atrás reducida a una línea, entra completo en la primera pantalla también
	 * en un móvil de 844px de alto.
	 */
	.hero :global(.waitlist) {
		margin-top: 32px;
	}

	.hero-proof {
		display: flex;
		flex-wrap: wrap;
		gap: 8px 20px;
		margin: 26px 0 0;
		padding: 0;
		list-style: none;
	}

	.hero-proof li {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 12.5px;
		color: var(--text-dim);
	}

	.hero-proof li::before {
		content: '';
		width: 12px;
		height: 12px;
		flex-shrink: 0;
		background: var(--green);
		mask: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%23000' stroke-width='3.2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='M20 6 9 17l-5-5'/%3E%3C/svg%3E")
			center / contain no-repeat;
	}

	@media (max-width: 940px) {
		.hero {
			grid-template-columns: 1fr;
			gap: 40px;
			padding-block: 60px 68px;
		}
		.hero-grid {
			background-size: 56px 56px;
			-webkit-mask-image: radial-gradient(ellipse 80% 46% at 60% 24%, #000 0%, transparent 74%);
			mask-image: radial-gradient(ellipse 80% 46% at 60% 24%, #000 0%, transparent 74%);
		}
	}

	@media (max-width: 640px) {
		.hero {
			gap: 32px;
			padding-block: 40px 52px;
		}
		.hero :global(.waitlist) {
			margin-top: 24px;
		}
	}

	@media (max-width: 480px) {
		.hero-title {
			margin-top: 18px;
		}
	}
</style>
