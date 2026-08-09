<script lang="ts">
	import HeroCountdown from './hero-countdown.svelte';
	import HeroWaitlist from './hero-waitlist.svelte';
	import HeroPreviewCards from './hero-preview-cards.svelte';
</script>

<section class="hero wrap">
	<div class="hero-left">
		<div class="hero-badge reveal">Acceso anticipado · Lanzamiento el 1 de octubre de 2026</div>

		<h1 class="hero-title reveal">
			Todo tu patrimonio,<br /><em>en tu mapa.</em>
		</h1>
		<p class="hero-sub reveal">
			Registra manualmente dónde tienes tus activos y agrúpalos en los portafolios que tú imaginas,
			aunque estén en distintas plataformas. Sin conectar cuentas, sin dar acceso a nadie.
		</p>

		<HeroCountdown />

		<HeroWaitlist />

		<ul class="hero-proof reveal">
			<li>Sin credenciales de tu broker</li>
			<li>Sin tarjeta</li>
			<li>Tus datos, exportables</li>
		</ul>
	</div>

	<HeroPreviewCards />
</section>

<style>
	.hero {
		display: grid;
		grid-template-columns: 55fr 45fr;
		gap: 56px;
		align-items: center;
		padding: 84px 0 92px;
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

	.hero-proof {
		display: flex;
		flex-wrap: wrap;
		gap: 8px 20px;
		margin: 22px 0 0;
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

	.hero-title {
		font-family: var(--font-display);
		font-weight: 300;
		font-size: clamp(42px, 5.6vw, 76px);
		line-height: 1.06;
		letter-spacing: -0.022em;
		margin: 28px 0 0;
	}

	.hero-title em {
		font-style: italic;
		font-weight: 500;
		color: var(--amber-light);
	}

	.hero-sub {
		margin: 22px 0 0;
		max-width: 46ch;
		font-size: clamp(15px, 1.6vw, 17px);
		color: var(--text-muted);
		font-weight: 300;
		line-height: 1.7;
	}

	/* COUNTDOWN */

	/* WAITLIST */

	/* HERO RIGHT — MOCK CARDS */

	@media (max-width: 940px) {
		.hero {
			grid-template-columns: 1fr;
			gap: 48px;
			padding: 60px 0 68px;
		}
	}

	@media (max-width: 640px) {
		.hero {
			gap: 40px;
			padding: 44px 0 52px;
		}
	}

	@media (max-width: 480px) {
		.hero-title {
			margin-top: 16px;
		}
	}
</style>
