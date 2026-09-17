<script lang="ts">
	/*
	 * El hero dice qué es Finexia en una frase y lo enseña debajo: el mapa de
	 * posiciones que se reagrupa. El formulario va al lado del argumento, a la
	 * altura de la primera pantalla, porque apuntarse es a lo que viene casi todo
	 * el mundo que llega aquí.
	 */
	import WaitlistForm from './waitlist-form.svelte';
	import HoldingsMap from './holdings-map.svelte';
</script>

<section class="hero" aria-labelledby="hero-title">
	<div class="lp-wrap">
		<h1 class="title" id="hero-title">Tu patrimonio, ordenado como tú lo piensas.</h1>

		<div class="pitch">
			<p class="lp-lead">
				Anota qué tienes en cada broker, exchange o banco y agrúpalo en tus propios portafolios.
				Finexia no se conecta a tus cuentas ni te pide ninguna contraseña.
			</p>
			<WaitlistForm anchor="waitlist" />
		</div>

		<HoldingsMap />
	</div>
</section>

<style>
	.hero {
		padding-block: 64px 8px;
	}

	.title {
		max-width: 14ch;
		margin: 0;
		font-size: clamp(44px, 6.6vw, 92px);
		font-stretch: 118%;
		font-weight: 640;
		line-height: 0.95;
		letter-spacing: -0.035em;
		text-wrap: balance;
	}

	.pitch {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 520px);
		gap: 32px 64px;
		align-items: start;
		margin: 36px 0 48px;
	}

	@media (max-width: 900px) {
		.hero {
			padding-block: 44px 8px;
		}
		.pitch {
			grid-template-columns: minmax(0, 1fr);
			margin: 28px 0 44px;
		}
	}

	@media (max-width: 640px) {
		.hero {
			padding-block: 28px 8px;
		}
	}
</style>
