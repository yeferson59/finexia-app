<script lang="ts">
	/*
	 * El cierre. El formulario está donde se toma la decisión, y es el mismo
	 * componente que el del hero, así que no hay dos acciones distintas que
	 * mantener. Va sobre el verde y enlaza con el pie, que sigue en el mismo
	 * fondo: la página termina en un solo bloque.
	 */
	import WaitlistForm from './waitlist-form.svelte';
</script>

<section class="cta lp-forest" aria-labelledby="cta-title">
	<div class="lp-wrap grid">
		<h2 class="lp-h2" id="cta-title">Te avisamos el día que abramos.</h2>
		<div class="side">
			<p class="lp-lead">
				Deja tu correo y el 1 de octubre te llega el acceso. Apuntarte es gratis y no pide tarjeta.
			</p>
			<WaitlistForm tone="forest" />
		</div>
	</div>
</section>

<style>
	.cta {
		padding-block: 112px 96px;
	}

	.grid {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 480px);
		gap: 32px 72px;
		align-items: end;
	}

	.grid .lp-h2 {
		font-size: clamp(38px, 5.2vw, 72px);
		font-stretch: 118%;
		line-height: 0.95;
	}

	.side .lp-lead {
		margin-bottom: 24px;
	}

	@media (max-width: 900px) {
		.cta {
			padding-block: 80px 64px;
		}
		.grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
