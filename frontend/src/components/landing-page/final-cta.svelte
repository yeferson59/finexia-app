<section class="block wrap">
	<div class="final-cta reveal">
		<div class="cta-eyebrow">Lista de espera</div>
		<h2>Sé de los primeros<br />en tener el control.</h2>
		<p>Cupos de acceso anticipado limitados. Reserva tu lugar antes del lanzamiento.</p>
		<a href="#waitlist" class="btn-amber btn-amber-lg">Unirme a la lista de espera</a>
	</div>
</section>

<style>
	.block {
		padding: 88px 0;
	}
	.final-cta {
		text-align: center;
		padding: 92px 40px;
		border-radius: 14px;
		position: relative;
		overflow: hidden;
		background: var(--surface);
		border: 1px solid var(--border-strong);
	}
	.final-cta::before {
		content: '';
		position: absolute;
		inset: 0;
		background: radial-gradient(
			ellipse 65% 110% at 50% -5%,
			rgba(212, 145, 42, 0.07),
			transparent 58%
		);
		pointer-events: none;
	}
	.cta-eyebrow {
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--text-dim);
		margin-bottom: 20px;
		position: relative;
	}
	.final-cta h2 {
		font-family: var(--font-display);
		font-weight: 300;
		font-size: clamp(30px, 5vw, 58px);
		line-height: 1.06;
		letter-spacing: -0.025em;
		text-wrap: balance;
		position: relative;
	}
	.final-cta p {
		margin: 16px auto 0;
		max-width: 44ch;
		font-size: 16px;
		color: var(--text-muted);
		font-weight: 300;
		position: relative;
	}
	.final-cta :global(.btn-amber-lg) {
		margin-top: 32px;
		position: relative;
	}
</style>
