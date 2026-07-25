<script lang="ts">
	import HeroCountdown from './hero-countdown.svelte';
	import HeroWaitlist from './hero-waitlist.svelte';
	import HeroPreviewCards from './hero-preview-cards.svelte';
</script>

<section class="hero wrap">
	<div class="hero-left">
		<h1 class="hero-title reveal">
			Todo tu patrimonio,<br /><em>en tu mapa.</em>
		</h1>
		<p class="hero-sub reveal">
			Registra manualmente dónde tienes tus activos y agrúpalos en los portafolios que tú imaginas,
			aunque estén en distintas plataformas. Sin conectar cuentas, sin dar acceso a nadie.
		</p>

		<HeroCountdown />

		<HeroWaitlist />
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
